#!/usr/bin/env python3
import os
import pathlib
import subprocess
import sys

MODE = sys.argv[1]
TARGET = pathlib.Path(sys.argv[2]).resolve()
HOST = os.environ.get('MYSQL_HOST', '127.0.0.1')
DB = os.environ['MYSQL_DB']
USER = os.environ['MYSQL_USER']
PASSWORD = os.getenv('MYSQL_PASSWORD', '')
PORT = os.getenv('MYSQL_PORT', '3306')
CONTAINER = os.getenv('MYSQL_CONTAINER', 'edu-schedule-system-mysql-1')


def has_binary(name: str) -> bool:
    return subprocess.call(['bash', '-lc', f'command -v {name} >/dev/null 2>&1']) == 0


def run(cmd, **kwargs):
    subprocess.run(cmd, check=True, **kwargs)


def docker_exec_base():
    base = ['docker', 'exec', '-i']
    if PASSWORD:
        base.extend(['-e', f'MYSQL_PWD={PASSWORD}'])
    base.append(CONTAINER)
    return base


def dump_with_docker(target: pathlib.Path):
    target.parent.mkdir(parents=True, exist_ok=True)
    cmd = docker_exec_base() + [
        'mysqldump',
        f'--host={HOST}',
        f'--port={PORT}',
        f'--user={USER}',
        '--single-transaction',
        '--skip-lock-tables',
        '--set-gtid-purged=OFF',
        DB,
    ]
    with target.open('wb') as fh:
        run(cmd, stdout=fh)


def restore_with_docker(source: pathlib.Path):
    cmd = docker_exec_base() + [
        'mysql',
        f'--host={HOST}',
        f'--port={PORT}',
        f'--user={USER}',
        DB,
    ]
    with source.open('rb') as fh:
        run(cmd, stdin=fh)


def dump_with_local(target: pathlib.Path):
    target.parent.mkdir(parents=True, exist_ok=True)
    env = os.environ.copy()
    if PASSWORD:
        env['MYSQL_PWD'] = PASSWORD
    cmd = [
        'mysqldump',
        f'--host={HOST}',
        f'--port={PORT}',
        f'--user={USER}',
        '--single-transaction',
        '--skip-lock-tables',
        '--set-gtid-purged=OFF',
        DB,
    ]
    with target.open('wb') as fh:
        run(cmd, stdout=fh, env=env)


def restore_with_local(source: pathlib.Path):
    env = os.environ.copy()
    if PASSWORD:
        env['MYSQL_PWD'] = PASSWORD
    cmd = [
        'mysql',
        f'--host={HOST}',
        f'--port={PORT}',
        f'--user={USER}',
        DB,
    ]
    with source.open('rb') as fh:
        run(cmd, stdin=fh, env=env)


if MODE == 'backup':
    if has_binary('mysqldump'):
        dump_with_local(TARGET)
    else:
        dump_with_docker(TARGET)
elif MODE == 'restore':
    if not TARGET.exists():
        raise SystemExit(f'backup file not found: {TARGET}')
    if has_binary('mysql'):
        restore_with_local(TARGET)
    else:
        restore_with_docker(TARGET)
else:
    raise SystemExit('unknown mode')
