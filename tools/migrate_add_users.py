#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Phase4 Task3: users テーブル追加
既存の menu.db に users テーブルを追加する。

使い方:
  python3 tools/migrate_add_users.py [data/menu.db]

※ 初回ユーザーは POST /api/auth/register で作成してください。
"""

import os
import sqlite3
import sys

PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DEFAULT_DB = os.path.join(PROJECT_ROOT, "data", "menu.db")


def migrate_sqlite(db_path: str) -> None:
    conn = sqlite3.connect(db_path)

    cur = conn.execute(
        "SELECT name FROM sqlite_master WHERE type='table' AND name='users'"
    )
    if cur.fetchone():
        print("users テーブルは既に存在します。スキップします。")
        conn.close()
        return

    print("Phase4 Task3: users テーブルを追加します")

    conn.execute("""
        CREATE TABLE users (
            id INTEGER PRIMARY KEY,
            email TEXT UNIQUE NOT NULL,
            password_hash TEXT NOT NULL,
            facility_id INTEGER REFERENCES facilities(id),
            role TEXT NOT NULL,
            created_at TEXT DEFAULT (datetime('now'))
        )
    """)

    conn.commit()
    conn.close()
    print("-" * 40)
    print("移行完了")


def main() -> None:
    db_path = sys.argv[1] if len(sys.argv) > 1 else DEFAULT_DB
    if not os.path.exists(db_path):
        print(f"エラー: {db_path} が存在しません")
        sys.exit(1)
    migrate_sqlite(db_path)


if __name__ == "__main__":
    main()
