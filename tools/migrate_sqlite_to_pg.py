#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
SQLite → PostgreSQL 移行ツール
Phase4 用。data/menu.db のデータを PostgreSQL にコピーする。

使い方:
  export DATABASE_URL="postgres://user:pass@localhost:5432/kyuushoku?sslmode=disable"
  python3 tools/migrate_sqlite_to_pg.py

事前準備:
  - PostgreSQL で DB 作成: createdb kyuushoku
  - pip install psycopg2-binary
"""

import os
import sqlite3
import sys

PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DATA_DIR = os.path.join(PROJECT_ROOT, "data")
SQLITE_PATH = os.path.join(DATA_DIR, "menu.db")
SCHEMA_PG = os.path.join(DATA_DIR, "schema_pg.sql")


def migrate() -> None:
    db_url = os.environ.get("DATABASE_URL", "postgres://localhost:5432/kyuushoku")
    if not db_url.startswith("postgres"):
        print("エラー: DATABASE_URL が postgres:// で始まっていません")
        sys.exit(1)

    if not os.path.exists(SQLITE_PATH):
        print(f"エラー: {SQLITE_PATH} が存在しません。先に tools/migrate_csv_to_sqlite.py を実行してください")
        sys.exit(1)

    try:
        import psycopg2
        from psycopg2 import sql
    except ImportError:
        print("エラー: psycopg2 がインストールされていません。pip install psycopg2-binary を実行してください")
        sys.exit(1)

    sqlite_conn = sqlite3.connect(SQLITE_PATH)
    sqlite_conn.row_factory = sqlite3.Row

    pg_conn = psycopg2.connect(db_url)
    pg_conn.autocommit = False

    try:
        # PostgreSQL スキーマ作成
        with open(SCHEMA_PG, "r", encoding="utf-8") as f:
            schema = f.read()
        with pg_conn.cursor() as cur:
            cur.execute(schema)
        print("PostgreSQL スキーマ作成完了")

        # 既存データ削除（依存順）
        truncate_order = ["menus", "recipe", "bulk_purchase_guide", "dishes", "ingredients"]
        with pg_conn.cursor() as cur:
            for t in truncate_order:
                try:
                    cur.execute(f"TRUNCATE TABLE {t} CASCADE")
                except Exception:
                    pass

        # データコピー（外部キー順）
        tables = [
            ("ingredients", ["id", "name", "ingredient_category", "energy", "protein", "fat",
             "carbohydrate", "salt", "unit", "waste_rate", "note"]),
            ("dishes", ["id", "name", "menu_category", "serving_size", "note"]),
            ("recipe", ["dish_id", "ingredient_id", "amount"]),
            ("menus", ["date", "staple_id", "main_id", "side_id", "soup_id", "dessert_id", "note"]),
            ("bulk_purchase_guide", ["ingredient_id", "order_unit_g", "order_unit_name", "bulk_category"]),
        ]

        for table_name, columns in tables:
            cur_sqlite = sqlite_conn.execute(f"SELECT {', '.join(columns)} FROM {table_name}")
            rows = cur_sqlite.fetchall()
            if not rows:
                print(f"  {table_name}: 0 件")
                continue

            placeholders = ", ".join(["%s"] * len(columns))
            col_str = ", ".join(columns)
            insert_sql = f"INSERT INTO {table_name} ({col_str}) VALUES ({placeholders})"

            with pg_conn.cursor() as cur:
                for row in rows:
                    cur.execute(insert_sql, list(row))
            print(f"  {table_name}: {len(rows)} 件")

        # version テーブル
        cur = sqlite_conn.execute("SELECT schema_version, updated_at FROM version LIMIT 1")
        ver_row = cur.fetchone()
        if ver_row:
            from datetime import datetime
            updated_at = ver_row[1] or datetime.now().isoformat()
            with pg_conn.cursor() as cur:
                cur.execute("DELETE FROM version")
                cur.execute(
                    "INSERT INTO version (schema_version, updated_at) VALUES (%s, %s)",
                    (ver_row[0], updated_at),
                )
            print("  version: 1 件")

        # シーケンスリセット（ingredients, dishes の id）
        with pg_conn.cursor() as cur:
            for seq_table in ["ingredients", "dishes"]:
                try:
                    cur.execute(
                        f"SELECT setval(pg_get_serial_sequence('{seq_table}', 'id'), "
                        f"COALESCE((SELECT MAX(id) FROM {seq_table}), 1))"
                    )
                except Exception:
                    pass

        pg_conn.commit()
        print("-" * 40)
        print("SQLite → PostgreSQL 移行完了")

    except Exception as e:
        pg_conn.rollback()
        raise
    finally:
        sqlite_conn.close()
        pg_conn.close()


def main() -> None:
    if not os.path.exists(SCHEMA_PG):
        print(f"エラー: {SCHEMA_PG} が存在しません")
        sys.exit(1)

    print("SQLite → PostgreSQL 移行")
    print("-" * 40)
    migrate()


if __name__ == "__main__":
    main()
