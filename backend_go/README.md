# Phase2：Go CLI + SQLite

開発計画書v2.0 の Phase2 実装。単体バイナリで配布可能。

## Go のインストール（macOS）

`go: command not found` が出る場合は、以下の手順で Go をインストールしてください。

### 公式インストーラでのインストール（推奨）

**1. ダウンロードページを開く**
- ブラウザで https://go.dev/dl/ を開く

**2. 自分の Mac に合ったファイルを選ぶ**
- **Apple Silicon（M1/M2/M3 など）** → `go1.26.0.darwin-arm64.pkg` をクリック
- **Intel Mac** → `go1.26.0.darwin-amd64.pkg` をクリック

**3. ダウンロードした .pkg を実行**
- ダウンロードフォルダの `.pkg` ファイルをダブルクリック
- 画面の指示に従って「続ける」→「インストール」→ パスワード入力 → 完了

**4. ターミナルを開き直す**
- 既に開いているターミナルは閉じ、**新しいターミナル**を開く  
  （PATH を反映するため）

**5. インストール確認**
```bash
go version
```
`go version go1.26.0` のように表示されれば OK。

### Homebrew を使う場合

```bash
brew install go
```
※ 権限エラーが出る場合は `sudo chown -R $(whoami) /opt/homebrew` を実行してから再試行

## 前提

1. **Go 1.21 以上** がインストールされていること（上記参照）
2. **SQLite データベース** が作成済みであること  
   → 先に `python3 tools/migrate_csv_to_sqlite.py` を実行

## ビルド・実行

```bash
cd /path/to/kyuushoku-system
python3 tools/migrate_csv_to_sqlite.py   # 初回のみ
cd backend_go
go mod tidy
go build -o menu-system .
```

※ `go: command not found` の場合は、上記「Go のインストール」を実行してください。

## テスト実行

**ターミナルで以下を実行**（backend_go ディレクトリで実行）：

```bash
cd /path/to/kyuushoku-system/backend_go
go test -v
```

※ 事前にプロジェクトルートで `python3 tools/migrate_csv_to_sqlite.py` を実行し、data/menu.db を作成してください。

## コマンド

| コマンド | 説明 | 例 |
|----------|------|-----|
| （引数なし） | ステップ1動作確認 | `./menu-system` |
| `calc` | 栄養価計算 | `./menu-system calc --date 2026-04-01` |
| `create-menu` | 献立登録 | `./menu-system create-menu --date 2026-04-01` |
| `order` | 発注集計 | `./menu-system order --start 2026-04-01 --end 2026-04-30 --people 120` |
| `export` | 献立表Excel出力 | `./menu-system export --month 2026-04 --output 献立表.xlsx` |
| `serve` | REST API サーバ起動（Phase3） | `./menu-system serve` または `./menu-system serve --port 8080` |

## 実行場所

`data/menu.db` をカレントディレクトリまたは親ディレクトリから検索する。  
プロジェクトルートで実行することを推奨：

```bash
cd /path/to/kyuushoku-system
./backend_go/menu-system calc --date 2026-04-01
```

## 参照

`docs/アーキテクチャ設計書.md`、`管理栄養士業務システム_開発計画書.md`、`docs/Phase1_実装書.md`
