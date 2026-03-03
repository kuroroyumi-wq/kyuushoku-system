# Phase1：Python + CSV（MVP）

開発計画書v2.0 の Phase1 実装。

## 成果物

- 献立登録CLI
- 栄養価計算機能
- 献立表Excel出力
- 発注集計機能

## 実行方法

```bash
python main.py create-menu --date 2026-04-01
python main.py calc --date 2026-04-01
python main.py order --start 2026-04-01 --end 2026-04-30 --people 120
python main.py export --month 2026-04 --output 献立表.xlsx
```

## テスト実行

```bash
cd /Applications/cursorフォルダ/idea
source venv/bin/activate
python -m pytest phase1_python/test_main.py -v
```

## 開発手順

`docs/開発手順_初心者向け.md` を参照。
