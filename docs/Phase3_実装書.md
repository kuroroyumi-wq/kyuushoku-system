# 管理栄養士業務システム Phase3 実装書

**版数**：1.0  
**作成日**：2026年3月3日  
**対象**：Phase3（Go REST API + TypeScript + React）

---

## 1. 概要

| 項目 | 内容 |
|------|------|
| バックエンド | Go REST API（Phase2 ロジック再利用） |
| フロントエンド | React + TypeScript + Vite |
| データ | SQLite（Phase2 と同じ） |

---

## 2. 起動手順

### 2.1 API サーバ

```bash
cd /Applications/cursorフォルダ/idea
python3 tools/migrate_csv_to_sqlite.py   # 初回のみ
cd backend_go
go build -o menu-system .
./menu-system serve
```

→ http://localhost:8080 で API が起動

### 2.2 フロントエンド

**前提**: Node.js がインストールされていること。`npm: command not found` の場合は https://nodejs.org/ja から LTS 版をインストール。

```bash
cd /Applications/cursorフォルダ/idea/frontend_web
npm install
npm run dev
```

→ http://localhost:5173 で Web UI が起動

※ Vite のプロキシで /api は localhost:8080 に転送される

---

## 3. 画面一覧

| パス | 画面 | 機能 |
|------|------|------|
| / | 献立一覧 | 月別の献立表示 |
| /menu-input | 献立入力 | 日付・5カテゴリ選択で登録 |
| /nutrition | 栄養計算 | 日付指定で栄養価表示 |
| /order | 発注集計 | 日別/期間・人数で集計、Excel ダウンロード |
| /bulk-order | まとめ買い | 米・調味料・乾物などのまとめ買い目安 |

---

## 4. API エンドポイント

| メソッド | パス | 説明 |
|----------|------|------|
| GET | /api/dishes | 献立候補一覧 |
| GET | /api/menus | 献立一覧（month）または単件（date） |
| POST | /api/menus | 献立登録 |
| GET | /api/calc | 栄養価計算 |
| GET | /api/order | 発注集計 |
| GET | /api/order/bulk | まとめ買い目安 |
| GET | /api/export | 献立表Excelダウンロード |

詳細は `docs/API仕様書.md` を参照。

---

## 5. ファイル構成

```
backend_go/
├── main.go      # CLI + serve コマンド
├── server.go    # REST API ハンドラ
└── main_test.go

frontend_web/
├── src/
│   ├── App.tsx
│   ├── api.ts       # API クライアント
│   ├── main.tsx
│   └── pages/
│       ├── MenuList.tsx
│       ├── MenuInput.tsx
│       ├── Nutrition.tsx
│       └── Order.tsx
├── package.json
└── vite.config.ts
```

---

## 6. トラブルシューティング

| エラー | 対処 |
|--------|------|
| `npm: command not found` | Node.js が未インストール。https://nodejs.org/ja から LTS 版をインストール |
| `address already in use` (8080) | ポート8080が使用中。`lsof -i :8080` でプロセス確認後 `kill <PID>` で終了。または `./menu-system serve --port 8081` で別ポート起動（その場合 vite.config.ts の proxy target を 8081 に変更） |
| API 接続エラー | `./menu-system serve` で API サーバが起動しているか確認 |

---

## 7. 改訂履歴

| 版数 | 日付 | 内容 |
|------|------|------|
| 1.0 | 2026年3月3日 | 初版作成 |
| 1.1 | 2026年3月3日 | Node.js インストール手順・トラブルシューティング追加 |
