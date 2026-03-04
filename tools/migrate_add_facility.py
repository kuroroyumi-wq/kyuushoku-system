#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Phase4 Task2: 既存 SQLite DB に facilities と facility_id を追加
既存の menu.db をその場で更新する。

使い方:
  python3 tools/migrate_add_facility.py [data/menu.db]
"""

import os
import sqlite3
import sys

PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DEFAULT_DB = os.path.join(PROJECT_ROOT, "data", "menu.db")


def migrate_sqlite(db_path: str) -> None:
    conn = sqlite3.connect(db_path)
    conn.row_factory = sqlite3.Row

    # facilities が既にあるか確認
    cur = conn.execute(
        "SELECT name FROM sqlite_master WHERE type='table' AND name='facilities'"
    )
    if cur.fetchone():
        print("facilities テーブルは既に存在します。スキップします。")
        conn.close()
        return

    print("Phase4: facilities と facility_id を追加します")

    # 1. facilities テーブル作成
    conn.execute("""
        CREATE TABLE facilities (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            code TEXT UNIQUE,
            created_at TEXT DEFAULT (datetime('now'))
        )
    """)
    conn.execute(
        "INSERT INTO facilities (id, name, code) VALUES (1, 'デフォルト', 'default')"
    )
    print("  facilities: 1 件追加")

    # 2. menus に facility_id があるか確認
    cur = conn.execute("PRAGMA table_info(menus)")
    cols = [row[1] for row in cur.fetchall()]
    if "facility_id" in cols:
        print("  menus.facility_id: 既に存在")
    else:
        # menus を再作成（date → date, facility_id の複合PK）
        conn.execute("""
            CREATE TABLE menus_new (
                date TEXT NOT NULL,
                facility_id INTEGER NOT NULL DEFAULT 1 REFERENCES facilities(id),
                staple_id INTEGER NOT NULL,
                main_id INTEGER NOT NULL,
                side_id INTEGER NOT NULL,
                soup_id INTEGER NOT NULL,
                dessert_id INTEGER NOT NULL,
                note TEXT,
                PRIMARY KEY (date, facility_id),
                FOREIGN KEY (staple_id) REFERENCES dishes(id),
                FOREIGN KEY (main_id) REFERENCES dishes(id),
                FOREIGN KEY (side_id) REFERENCES dishes(id),
                FOREIGN KEY (soup_id) REFERENCES dishes(id),
                FOREIGN KEY (dessert_id) REFERENCES dishes(id)
            )
        """)
        conn.execute("""
            INSERT INTO menus_new (date, facility_id, staple_id, main_id, side_id, soup_id, dessert_id, note)
            SELECT date, 1, staple_id, main_id, side_id, soup_id, dessert_id, COALESCE(note, '')
            FROM menus
        """)
        conn.execute("DROP TABLE menus")
        conn.execute("ALTER TABLE menus_new RENAME TO menus")
        conn.execute("CREATE INDEX IF NOT EXISTS idx_menus_facility_date ON menus(facility_id, date)")
        print("  menus: facility_id 追加、既存データを facility_id=1 に紐づけ")

    # 3. ingredients, dishes に facility_id 追加（NULL = 共通マスタ）
    for table in ["ingredients", "dishes"]:
        cur = conn.execute(f"PRAGMA table_info({table})")
        cols = [row[1] for row in cur.fetchall()]
        if "facility_id" in cols:
            print(f"  {table}.facility_id: 既に存在")
        else:
            conn.execute(f"ALTER TABLE {table} ADD COLUMN facility_id INTEGER")
            print(f"  {table}: facility_id 追加 (NULL=共通)")

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
