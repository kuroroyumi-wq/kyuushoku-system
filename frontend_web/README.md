# Phase3：TypeScript + React

献立入力・栄養計算・発注集計の Web UI。

## Node.js のインストール（npm: command not found の場合）

`npm: command not found` が出る場合は、Node.js をインストールしてください。

**公式インストーラ（推奨）**
1. https://nodejs.org/ja を開く
2. LTS 版の .pkg をダウンロードして実行
3. 新しいターミナルを開く

**Homebrew**
```bash
brew install node
```

**確認**
```bash
node -v   # v18.x 以上
npm -v    # 9.x 以上
```

## 前提

1. **API サーバ**が起動していること  
   ```bash
   cd backend_go && ./menu-system serve
   ```
2. **Node.js** がインストールされていること（上記参照）

## セットアップ・実行

```bash
cd /Applications/cursorフォルダ/idea/frontend_web
npm install
npm run dev
```

ブラウザで http://localhost:5173 を開く。

## 画面

- **献立一覧**：月別の献立表示
- **献立入力**：日付を指定して5カテゴリを選択して登録
- **栄養計算**：日付を指定して栄養価を表示
- **発注集計**：期間・人数を指定して集計し、Excel ダウンロード

## 参照

`docs/API仕様書.md`、`docs/アーキテクチャ設計書.md`
