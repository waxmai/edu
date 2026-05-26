#!/usr/bin/env python3
import json
import os
import sys
import urllib.request
import urllib.error

BASE = os.getenv('API_BASE_URL', 'http://127.0.0.1:9999/api/v1').rstrip('/')
ADMIN_USERNAME = os.getenv('SMOKE_ADMIN_USERNAME', 'platform_admin')
ADMIN_PASSWORD = os.getenv('SMOKE_ADMIN_PASSWORD', '')
ADMIN_NEW_PASSWORD = os.getenv('SMOKE_ADMIN_NEW_PASSWORD', 'AdminSmoke123!')
ADMIN_TOKEN = os.getenv('SMOKE_ADMIN_TOKEN', '').strip()
ADMIN_REFRESH_TOKEN = os.getenv('SMOKE_ADMIN_REFRESH_TOKEN', '').strip()
TEACHER_USERNAME = os.getenv('SMOKE_TEACHER_USERNAME', 'teacher01')
TEACHER_PASSWORD = os.getenv('SMOKE_TEACHER_PASSWORD', '')
TEACHER_NEW_PASSWORD = os.getenv('SMOKE_TEACHER_NEW_PASSWORD', 'TeacherSmoke123!')
TEACHER_TOKEN = os.getenv('SMOKE_TEACHER_TOKEN', '').strip()
TEACHER_REFRESH_TOKEN = os.getenv('SMOKE_TEACHER_REFRESH_TOKEN', '').strip()


def fail(message):
    print(json.dumps({'ok': False, 'message': message}, ensure_ascii=False))
    raise SystemExit(1)


def request(method, path, payload=None, token=None, refresh_token=None):
    headers = {'Content-Type': 'application/json'}
    if token:
        headers['Authorization'] = f'Bearer {token}'
    if refresh_token:
        headers['X-Refresh-Token'] = refresh_token
    data = None if payload is None else json.dumps(payload).encode('utf-8')
    req = urllib.request.Request(BASE + path, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            body = resp.read().decode('utf-8')
            return resp.status, json.loads(body) if body else {}
    except urllib.error.HTTPError as e:
        body = e.read().decode('utf-8')
        try:
            parsed = json.loads(body) if body else {}
        except Exception:
            parsed = {'raw': body}
        return e.code, parsed


def expect_status(actual, expected, context, payload=None):
    if actual != expected:
        fail(f'{context} expected {expected}, got {actual}, payload={json.dumps(payload, ensure_ascii=False)}')


def login(username, password):
    status, payload = request('POST', '/auth/login', {'username': username, 'password': password})
    expect_status(status, 200, f'login({username})', payload)
    data = payload.get('data') or {}
    if not data.get('token') or not data.get('userInfo'):
        fail(f'login({username}) missing token or userInfo: {json.dumps(payload, ensure_ascii=False)}')
    return data


def ensure_password_ready(username, password, new_password, token='', refresh_token=''):
    if token:
        status, payload = request('GET', '/auth/me', token=token)
        expect_status(status, 200, f'auth/me({username}) with provided token', payload)
        user_info = payload.get('data') or {}
        return {
            'token': token,
            'refreshToken': refresh_token,
            'userInfo': user_info,
        }

    login_data = login(username, password)
    if login_data.get('userInfo', {}).get('mustChangePassword'):
        status, payload = request(
            'POST',
            '/auth/change-password',
            {'oldPassword': password, 'newPassword': new_password},
            token=login_data['token'],
            refresh_token=login_data.get('refreshToken'),
        )
        expect_status(status, 200, f'change-password({username})', payload)
        login_data = payload.get('data') or {}
        if not login_data.get('token') or not login_data.get('userInfo'):
            fail(f'change-password({username}) missing new session data: {json.dumps(payload, ensure_ascii=False)}')
    return login_data


def main():
    admin = ensure_password_ready(ADMIN_USERNAME, ADMIN_PASSWORD, ADMIN_NEW_PASSWORD, ADMIN_TOKEN, ADMIN_REFRESH_TOKEN)
    teacher = ensure_password_ready(TEACHER_USERNAME, TEACHER_PASSWORD, TEACHER_NEW_PASSWORD, TEACHER_TOKEN, TEACHER_REFRESH_TOKEN)

    checks = []

    status, payload = request('GET', '/auth/me', token=admin['token'])
    expect_status(status, 200, 'admin auth/me', payload)
    checks.append({'case': 'admin_me', 'status': status})

    status, payload = request('GET', '/users', token=admin['token'])
    expect_status(status, 200, 'admin users list', payload)
    checks.append({'case': 'admin_users_list', 'status': status})

    status, payload = request('GET', '/users', token=teacher['token'])
    expect_status(status, 403, 'teacher users list forbidden', payload)
    checks.append({'case': 'teacher_users_forbidden', 'status': status, 'message': payload.get('message')})

    status, payload = request('GET', '/courses', token=teacher['token'])
    expect_status(status, 200, 'teacher courses list', payload)
    checks.append({'case': 'teacher_courses_list', 'status': status})

    status, payload = request('GET', '/lesson-packages', token=teacher['token'])
    expect_status(status, 200, 'teacher lesson packages list', payload)
    checks.append({'case': 'teacher_lesson_packages_list', 'status': status})

    status, payload = request('GET', '/payment-records', token=teacher['token'])
    expect_status(status, 403, 'teacher payment records list forbidden', payload)
    checks.append({'case': 'teacher_payment_records_forbidden', 'status': status, 'message': payload.get('message')})

    status, payload = request('POST', '/courses', {'courseName': 'SmokeCourse', 'subject': '数学', 'courseType': 'group'}, token=teacher['token'])
    expect_status(status, 403, 'teacher create course forbidden', payload)
    checks.append({'case': 'teacher_create_course_forbidden', 'status': status, 'message': payload.get('message')})

    status, payload = request('POST', '/payment-records', {'studentId': 1, 'lessonPackageId': 1, 'amount': 100, 'paymentType': 'tuition', 'paymentMethod': 'cash', 'paymentTime': '2026-05-10 10:00:00', 'paymentStatus': 'paid'}, token=teacher['token'])
    expect_status(status, 403, 'teacher create payment forbidden', payload)
    checks.append({'case': 'teacher_create_payment_forbidden', 'status': status, 'message': payload.get('message')})

    status, payload = request('POST', '/auth/logout', token=admin['token'], refresh_token=admin.get('refreshToken'))
    expect_status(status, 200, 'admin logout', payload)
    checks.append({'case': 'admin_logout', 'status': status})

    status, payload = request('GET', '/auth/me', token=admin['token'])
    expect_status(status, 401, 'admin token invalid after logout', payload)
    checks.append({'case': 'admin_token_revoked', 'status': status})

    print(json.dumps({'ok': True, 'base': BASE, 'checks': checks}, ensure_ascii=False, indent=2))


if __name__ == '__main__':
    main()
