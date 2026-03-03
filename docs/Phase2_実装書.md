# 管理栄養士業務システム Phase2 実装書

**版数**：1.0  
**作成日**：2026年3月3日  
**対象**：Phase2（Go + SQLite）の実装仕様

---

## 1. 概要

Phase2 は Phase1 の機能を Go と SQLite で再実装し、単体バイナリ化とデータ整合性強化を実現する。

| 項目 | 内容 |
|------|------|
| 言語 | Go 1.21以上 |
| データ | SQLite（data/menu.db） |
| データ投入 | CSV → SQLite 移行ツール（Python） |
| 実行形態 | CLI（単体バイナリ） |

---

## 2. ファイル構成

```
project-root/
├── data/
│   ├── schema.sql      # SQLite スキーマ定義
│   ├── schema.json     # schema_version（CSV と共通）
│   ├── *.csv           # マスタデータ（移行元）
│   └── menu.db         # SQLite データベース（移行後）
├── tools/
│   ├── migrate_csv_to_sqlite.py  # CSV → SQLite 移行ツール
│   └── README.md
└── backend_go/
    ├── main.go         # CLI エントリポイント
    ├── go.mod
    └── README.md
```

---

## 3. 移行ツール（Python）

### 3.1 役割

data/ の CSV を SQLite にインポートする。Phase2 実行前に必須。

### 3.2 実行

```bash
cd /path/to/kyuushoku-system
python3 tools/migrate_csv_to_sqlite.py
```

### 3.3 処理内容

1. data/schema.sql でテーブル作成
2. version テーブルに schema.json の schema_version を保存
3. ingredients, dishes, recipe, menus を CSV から投入

---

## 4. SQLite スキーマ

### 4.1 テーブル一覧

| テーブル | 説明 |
|----------|------|
| version | schema_version, updated_at |
| ingredients | 食材マスタ |
| dishes | 献立候補（料理マスタ） |
| recipe | レシピ（dish_id, ingredient_id, amount） |
| menus | 献立（date, staple_id, main_id, side_id, soup_id, dessert_id, note） |

### 4.2 外部キー

- recipe.dish_id → dishes.id
- recipe.ingredient_id → ingredients.id
- menus.staple_id, main_id, side_id, soup_id, dessert_id → dishes.id

---

## 5. Go CLI コマンド

| コマンド | 説明 | 例 |
|----------|------|-----|
| （引数なし） | ステップ1動作確認 | `./menu-system` |
| `calc` | 献立の栄養価計算 | `./menu-system calc --date 2026-04-01` |
| `create-menu` | 献立登録（対話式） | `./menu-system create-menu --date 2026-04-01` |
| `order` | 発注集計 | `./menu-system order --start 2026-04-01 --end 2026-04-30 --people 120` |
| `export` | 献立表Excel出力 | `./menu-system export --month 2026-04 --output 献立表.xlsx` |

---

## 6. データベース検索パス

`data/menu.db` を以下の順で検索：

1. カレントディレクトリ/data/menu.db
2. 親ディレクトリ/data/menu.db
3. 親の親ディレクトリ/data/menu.db

プロジェクトルートで実行することを推奨。

---

## 7. Phase1 との互換性

| 項目 | 互換内容 |
|------|----------|
| 栄養計算 | 同一計算式・丸めルール（Phase1_実装書準拠） |
| 発注集計 | 同一食材合算・人数倍率・出力順（ingredient_id 昇順） |
| Excel出力 | 同一必須列・シート名形式 |
| 献立登録 | 同日登録禁止、5カテゴリ選択 |

---

## 8. 依存パッケージ

| パッケージ | 用途 |
|------------|------|
| modernc.org/sqlite | SQLite ドライバ（pure Go、cgo 不要） |
| github.com/xuri/excelize/v2 | Excel 読み書き |

---

## 9. テスト実行

```bash
cd /path/to/kyuushoku-system/backend_go
go test -v
```

※ go.mod は backend_go/ 内にあるため、backend_go ディレクトリで実行する。

テスト対象: 栄養価計算、発注集計、参照整合性チェック（Phase1 と同等）

---

## 10. トラブルシューティング

| エラー | 対処 |
|--------|------|
| `go: command not found` | Go が未インストール。`backend_go/README.md` の「Go のインストール」を参照 |
| `no such file or directory: ./menu-system` | `go build` が失敗しているため実行ファイルが未作成。上記を解消してから `go build` を再実行 |
| `データベースが見つかりません` | 先に `python3 tools/migrate_csv_to_sqlite.py` を実行 |

---

## 11. 関連ドキュメント

| ドキュメント | 内容 |
|--------------|------|
| Phase1_実装書 | 栄養計算・発注集計のロジック詳細 |
| データ設計書 | スキーマ詳細 |
| アーキテクチャ設計書 | フェーズ構成 |

---

## 12. 改訂履歴

| 版数 | 日付 | 内容 |
|------|------|------|
| 1.0 | 2026年3月3日 | 初版作成。移行ツール、Go CLI、SQLite スキーマを記載 |
| 1.1 | 2026年3月3日 | Go 単体テスト追加（main_test.go）。Phase2 完了条件を満たす |
