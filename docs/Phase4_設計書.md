# Phase4 設計書（PostgreSQL + 本格運用）

**版数**：1.0  
**作成日**：2026年3月3日  
**ステータス**：設計検討中

---

## 1. 概要

Phase4 は**複数施設対応**と**本番運用体制**の確立を目的とする。

| 項目 | 内容 |
|------|------|
| **目的** | 複数施設対応・本番運用体制確立 |
| **技術** | Go（原則）、PostgreSQL、React |
| **成果物** | マルチ施設設計、権限管理、監査ログ、バックアップ設計、運用Runbook |
| **完了条件** | 複数施設データ分離。定期バックアップ自動化。運用ドキュメント完成 |

---

## 2. マルチ施設設計

### 2.1 施設（facility）の概念

```
施設（facility）
├── 施設マスタ（施設名、住所、利用者数など）
├── 献立データ（施設ごとに独立）
├── 食材・料理マスタ（共通 or 施設別）
└── ユーザー（施設に紐づく）
```

### 2.2 データ分離の方式

| 方式 | メリット | デメリット |
|------|----------|------------|
| **A. テナントID方式** | 1DBで管理、スキーマ変更が容易 | データ漏洩リスク、クエリに facility_id 必須 |
| **B. スキーマ分離** | 1DB内で論理分離 | PostgreSQL の schema 管理が複雑 |
| **C. DB分離** | 完全分離、セキュリティ高い | 運用・バックアップが施設数分増 |
| **D. ハイブリッド** | マスタは共通、業務データは施設別 | 設計が複雑 |

**推奨**: **A. テナントID方式**（初期は施設数が少ない想定）

### 2.3 スキーマ変更案（テナントID方式）

```sql
-- 施設マスタ（新規）
CREATE TABLE facilities (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    code TEXT UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 既存テーブルに facility_id を追加
ALTER TABLE menus ADD COLUMN facility_id INTEGER REFERENCES facilities(id);
ALTER TABLE ingredients ADD COLUMN facility_id INTEGER;  -- NULL で共通マスタ
ALTER TABLE dishes ADD COLUMN facility_id INTEGER;
-- recipe は dish 経由で facility を継承

-- インデックス
CREATE INDEX idx_menus_facility_date ON menus(facility_id, date);
```

---

## 3. 権限管理

### 3.1 ロール（役割）

| ロール | 権限 |
|--------|------|
| **システム管理者** | 全施設の管理、ユーザー管理 |
| **施設管理者** | 自施設の献立・発注・設定 |
| **栄養士** | 自施設の献立入力・栄養計算・発注 |
| **閲覧者** | 自施設の参照のみ |

### 3.2 ユーザー・認証

| 項目 | 方針 |
|------|------|
| 認証方式 | JWT または セッション |
| ユーザー管理 | users テーブル + facility_id |
| パスワード | bcrypt ハッシュ |

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    facility_id INTEGER REFERENCES facilities(id),
    role TEXT NOT NULL,  -- admin, facility_admin, dietitian, viewer
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

## 4. 監査ログ

### 4.1 記録対象

| 操作 | 記録内容 |
|------|----------|
| 献立登録・更新・削除 | 誰が・いつ・何を |
| 発注集計実行 | 誰が・いつ・期間・人数 |
| ユーザー管理 | 誰が・いつ・何を変更 |
| ログイン | 誰が・いつ・成功/失敗 |

### 4.2 テーブル設計

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    facility_id INTEGER REFERENCES facilities(id),
    user_id INTEGER REFERENCES users(id),
    action TEXT NOT NULL,  -- create_menu, update_menu, login, etc.
    resource_type TEXT,   -- menu, user, order
    resource_id TEXT,
    old_value JSONB,
    new_value JSONB,
    ip_address INET,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_facility_created ON audit_logs(facility_id, created_at);
```

---

## 5. PostgreSQL 移行

### 5.1 移行手順

1. **PostgreSQL スキーマ作成**（data/schema_pg.sql）
2. **SQLite → PostgreSQL 移行ツール**（tools/migrate_sqlite_to_pg.py）
3. **Go の DB ドライバを切り替え**（database/sql + driver: lib/pq または pgx）
4. **環境変数で DB 接続先を切り替え**（SQLite / PostgreSQL）

### 5.2 接続文字列

```
DATABASE_URL=postgres://user:pass@localhost:5432/kyuushoku?sslmode=disable
```

### 5.3 型差異の対応

| SQLite | PostgreSQL |
|--------|------------|
| INTEGER | INTEGER / SERIAL |
| TEXT | TEXT / VARCHAR |
| REAL | REAL / DOUBLE PRECISION |
| date | DATE / TIMESTAMPTZ |

---

## 6. バックアップ設計

### 6.1 バックアップ対象

| 対象 | 頻度 | 保持期間 |
|------|------|----------|
| PostgreSQL データ | 日次 | 30日 |
| 設定ファイル | 変更時 | 90日 |

### 6.2 方式

- **pg_dump** による論理バックアップ
-  cron または systemd timer で定期実行

```bash
# 日次バックアップ例
pg_dump -Fc kyuushoku > backup_$(date +%Y%m%d).dump
```

---

## 7. 運用 Runbook（概要）

| 項目 | 内容 |
|------|------|
| 起動・停止 | systemd サービス定義 |
| 監視 | ヘルスチェック API、ログ |
| 障害対応 | よくあるエラーと対処 |
| ログローテーション | logrotate 設定 |
| バックアップ復元 | 手順書 |

---

## 8. 実装順序（提案）

| 順序 | タスク | 工数目安 |
|------|--------|----------|
| 1 | SQLite と PostgreSQL の両対応（環境変数切り替え） | 1〜2日 |
| 2 | 施設マスタ・facility_id 追加 | 2〜3日 |
| 3 | ユーザー・認証（JWT） | 2〜3日 |
| 4 | 権限チェック（API ミドルウェア） | 1〜2日 |
| 5 | 監査ログ | 1〜2日 |
| 6 | バックアップスクリプト | 0.5日 |
| 7 | 運用 Runbook | 1日 |

---

## 9. 次のアクション

1. **設計の確定**: テナントID方式で進めるか決定
2. **PostgreSQL 環境構築**: ローカルまたは Docker で PostgreSQL を用意
3. **移行ツール作成**: SQLite → PostgreSQL の移行スクリプト作成

---

## 10. 改訂履歴

| 版数 | 日付 | 内容 |
|------|------|------|
| 1.0 | 2026年3月3日 | 初版作成 |