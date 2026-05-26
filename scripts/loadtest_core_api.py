import json
import os
import random
import threading
import time
import urllib.request
import urllib.error

BASE = os.environ['API_BASE_URL'].rstrip('/')
TOKEN = os.environ['AUTH_TOKEN'].strip()
HEADERS = {
    'Authorization': f'Bearer {TOKEN}',
    'Content-Type': 'application/json',
}
THREADS = int(os.getenv('LOADTEST_THREADS', '8'))
ITERATIONS = int(os.getenv('LOADTEST_ITERATIONS', '10'))
SCHEDULE_ID = int(os.getenv('LOADTEST_SCHEDULE_ID', '1'))
PAYMENT_ID = int(os.getenv('LOADTEST_PAYMENT_ID', '1'))
LESSON_PACKAGE_ID = int(os.getenv('LOADTEST_LESSON_PACKAGE_ID', '1'))
STUDENT_ID = int(os.getenv('LOADTEST_STUDENT_ID', '1'))
COURSE_ID = int(os.getenv('LOADTEST_COURSE_ID', '1'))
TEACHER_ID = int(os.getenv('LOADTEST_TEACHER_ID', '1'))

results = []
lock = threading.Lock()


def request(method, path, payload=None):
    data = None
    if payload is not None:
        data = json.dumps(payload).encode('utf-8')
    req = urllib.request.Request(BASE + path, data=data, method=method, headers=HEADERS)
    started = time.time()
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            body = resp.read().decode('utf-8')
            return resp.status, time.time() - started, body
    except urllib.error.HTTPError as e:
        return e.code, time.time() - started, e.read().decode('utf-8')
    except urllib.error.URLError as e:
        return 0, time.time() - started, str(e)


def record_result(case, status, cost, body, payload):
    with lock:
        results.append((case, status, cost, body, payload))


def worker_create_schedule(worker_id):
    for i in range(ITERATIONS):
        minute = (worker_id * ITERATIONS + i) % 50
        start = f'2026-05-20 10:{minute:02d}:00'
        end = f'2026-05-20 11:{minute:02d}:00'
        payload = {
            'studentId': STUDENT_ID,
            'courseId': COURSE_ID,
            'teacherId': TEACHER_ID,
            'lessonPackageId': LESSON_PACKAGE_ID,
            'classDate': '2026-05-20',
            'startTime': start,
            'endTime': end,
            'classroom': f'loadtest-room-{worker_id}',
            'scheduleStatus': 'scheduled'
        }
        status, cost, body = request('POST', '/schedules', payload)
        record_result('schedule_create', status, cost, body, payload)


def worker_update_payment(_worker_id):
    for i in range(ITERATIONS):
        amount = 100 + (i % 3)
        payload = {
            'amount': amount,
            'remark': f'loadtest-{i}'
        }
        status, cost, body = request('PUT', f'/payment-records/{PAYMENT_ID}', payload)
        record_result('payment_update', status, cost, body, payload)


def worker_complete_schedule(_worker_id):
    for i in range(ITERATIONS):
        payload = {
            'scheduleStatus': 'completed' if i % 2 == 0 else 'scheduled'
        }
        status, cost, body = request('PUT', f'/schedules/{SCHEDULE_ID}', payload)
        record_result('schedule_complete_toggle', status, cost, body, payload)


def run_group(name, fn):
    threads = [threading.Thread(target=fn, args=(i,)) for i in range(THREADS)]
    for t in threads:
        t.start()
    for t in threads:
        t.join()
    subset = [r for r in results if r[0] == name]
    ok = sum(1 for _, status, _, _, _ in subset if 200 <= status < 300)
    conflicts = sum(1 for _, status, _, _, _ in subset if status == 409)
    client_errors = sum(1 for _, status, _, _, _ in subset if 400 <= status < 500 and status != 409)
    errors = sum(1 for _, status, _, _, _ in subset if status >= 500 or status == 0)
    avg = sum(cost for _, _, cost, _, _ in subset) / len(subset) if subset else 0
    sample_failures = [
        {'status': status, 'body': body[:300], 'payload': payload}
        for _, status, _, body, payload in subset if status >= 400 or status == 0
    ][:5]
    print(json.dumps({
        'case': name,
        'count': len(subset),
        'ok': ok,
        'conflicts': conflicts,
        'client_errors': client_errors,
        'server_errors': errors,
        'avg_seconds': round(avg, 4),
        'sample_failures': sample_failures,
    }, ensure_ascii=False))


if __name__ == '__main__':
    run_group('schedule_create', worker_create_schedule)
    run_group('payment_update', worker_update_payment)
    run_group('schedule_complete_toggle', worker_complete_schedule)
