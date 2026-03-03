#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
CSV → SQLite 移行ツール
Phase2 準備。data/ の CSV を SQLite にインポートする。
"""

import csv
import json
import os
import sqlite3
import sys

# プロジェクトルート（tools の親）
PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DATA_DIR = os.path.join(PROJECT_ROOT, "data")
SCHEMA_SQL = os.path.join(DATA_DIR, "schema.sql")
DEFAULT_DB_PATH = os.path.join(PROJECT_ROOT, "data", "menu.db")


def load_schema_version() -> str:
    """schema.json から schema_version を取得"""
    path = os.path.join(DATA_DIR, "schema.json")
    with open(path, "r", encoding="utf-8") as f:
        data = json.load(f)
    return data["schema_version"]


def load_csv(filepath: str) -> list[dict]:
    """CSV を読み込み、辞書のリストを返す"""
    rows = []
    with open(filepath, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            if any(row.values()):
                rows.append(row)
    return rows


def migrate(db_path: str) -> None:
    """CSV を SQLite に移行"""
    # スキーマ実行
    with open(SCHEMA_SQL, "r", encoding="utf-8") as f:
        schema_sql = f.read()

    conn = sqlite3.connect(db_path)
    conn.executescript(schema_sql)

    # version テーブル（schema.json と同期）
    schema_version = load_schema_version()
    conn.execute("DELETE FROM version")
    conn.execute(
        "INSERT INTO version (schema_version, updated_at) VALUES (?, date('now'))",
        (schema_version,),
    )

    # 各 CSV をインポート（既存データは削除して再投入）
    tables = [
        ("ingredients.csv", "ingredients", [
            "id", "name", "ingredient_category", "energy", "protein", "fat",
            "carbohydrate", "salt", "unit", "waste_rate", "note"
        ]),
        ("dishes.csv", "dishes", ["id", "name", "menu_category", "serving_size", "note"]),
        ("recipe.csv", "recipe", ["dish_id", "ingredient_id", "amount"]),
        ("menus.csv", "menus", ["date", "staple_id", "main_id", "side_id", "soup_id", "dessert_id", "note"]),
    ]

    # 外部キー順に削除（recipe → menus は独立、ingredients/dishes は recipe から参照される）
    for _, table_name, _ in reversed(tables):
        conn.execute(f"DELETE FROM {table_name}")

    for csv_name, table_name, columns in tables:
        csv_path = os.path.join(DATA_DIR, csv_name)
        if not os.path.exists(csv_path):
            print(f"スキップ: {csv_name} が存在しません")
            continue

        rows = load_csv(csv_path)
        if not rows:
            print(f"  {csv_name} → {table_name}: 0 件（空）")
            continue

        placeholders = ", ".join(["?"] * len(columns))
        col_str = ", ".join(columns)
        for row in rows:
            values = [row.get(c, "") for c in columns]
            conn.execute(
                f"INSERT INTO {table_name} ({col_str}) VALUES ({placeholders})",
                values,
            )
        print(f"  {csv_name} → {table_name}: {len(rows)} 件")

    conn.commit()
    conn.close()
    print(f"移行完了: {db_path}")


def main() -> None:
    db_path = sys.argv[1] if len(sys.argv) > 1 else DEFAULT_DB_PATH
    if not os.path.exists(DATA_DIR):
        print(f"エラー: {DATA_DIR} が存在しません")
        sys.exit(1)
    if not os.path.exists(SCHEMA_SQL):
        print(f"エラー: {SCHEMA_SQL} が存在しません")
        sys.exit(1)

    print("CSV → SQLite 移行")
    print("-" * 40)
    migrate(db_path)


if __name__ == "__main__":
    main()
