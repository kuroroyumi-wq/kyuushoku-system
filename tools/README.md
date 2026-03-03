# 移行ツール

## CSV → SQLite 移行

Phase2 で使用する SQLite データベースを、data/ の CSV から作成する。

### 実行方法

```bash
cd /Applications/cursorフォルダ/idea
python3 tools/migrate_csv_to_sqlite.py
```

### 出力

- `data/menu.db` に SQLite データベースが作成される
- 既存の menu.db がある場合は上書き（全テーブルを再投入）

### 前提

- data/schema.sql が存在すること
- data/schema.json が存在すること
- data/*.csv が存在すること
