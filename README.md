# 管理栄養士業務システム（kyuushoku-system）

保育園3〜5歳児向け給食管理システム。献立作成・栄養計算・発注集計・まとめ買い計画をサポートします。

## 技術スタック

| 層 | 技術 |
|----|------|
| バックエンド | Go, SQLite, REST API |
| フロントエンド | React, TypeScript, Vite |
| データ | CSV → SQLite |

## 機能

- **献立一覧** - 月別表示
- **献立入力** - 登録・更新
- **栄養計算** - 日付指定で栄養価表示
- **発注集計** - 1日単位 / 期間、調味料除外オプション
- **まとめ買い** - 米・調味料・乾物の発注目安

## クイックスタート

### SQLite（デフォルト）

```bash
# 1. データ投入（初回のみ）
python3 tools/migrate_csv_to_sqlite.py

# 2. API サーバ
cd backend_go && ./menu-system serve
# → http://localhost:8080

# 3. フロントエンド（別ターミナル）
cd frontend_web && npm install && npm run dev
# → http://localhost:5173
```

### PostgreSQL（Phase4）

```bash
# 1. DB 作成: createdb kyuushoku
# 2. 移行: pip install psycopg2-binary && python3 tools/migrate_sqlite_to_pg.py
# 3. 起動: export DATABASE_URL="postgres://localhost:5432/kyuushoku" && cd backend_go && ./menu-system serve
```

## 構成

```
kyuushoku-system/
├── backend_go/     # Go API サーバ
├── frontend_web/   # React フロントエンド
├── data/           # SQLite DB・CSV
├── tools/          # 移行スクリプト
├── docs/           # 設計書・API仕様
└── phase1_python/  # Python CLI（レガシー）
```

## ドキュメント

- [API仕様書](docs/API仕様書.md)
- [再開用コンテキスト](docs/再開用_コンテキスト.md)
- [Phase4 設計書](docs/Phase4_設計書.md)

## ライセンス

MIT License
