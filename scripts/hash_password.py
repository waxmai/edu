#!/usr/bin/env python3
import argparse
import sys

try:
    import bcrypt
except ImportError:
    print('python3 package bcrypt is required', file=sys.stderr)
    raise SystemExit(1)


def main():
    parser = argparse.ArgumentParser(description='Generate bcrypt hash for bootstrap passwords')
    parser.add_argument('password', help='plain text password to hash')
    args = parser.parse_args()
    hashed = bcrypt.hashpw(args.password.encode('utf-8'), bcrypt.gensalt(rounds=10))
    print(hashed.decode('utf-8'))


if __name__ == '__main__':
    main()
