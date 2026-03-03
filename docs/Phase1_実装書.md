# 管理栄養士業務システム Phase1 実装書

**版数**：1.0  
**作成日**：2026年3月3日  
**対象**：Phase1（Python + CSV）の実装仕様

---

## 1. 概要

Phase1 は保育園3〜5歳児向け給食管理の MVP 機能を、Python と CSV で実装した CLI アプリケーションである。

| 項目 | 内容 |
|------|------|
| 言語 | Python 3.8以上 |
| データ | CSV（data/ 配下） |
| 実行形態 | CLI（コマンドライン） |
| 参照 | 要件定義書v2.3、開発計画書v2.0、アーキテクチャ設計書v2.0 |

---

## 2. ファイル構成

```
phase1_python/
├── main.py           # メインエントリポイント（約530行）
├── test_main.py      # 単体テスト（pytest）
├── requirements.txt # 依存パッケージ
└── README.md         # 実行方法・テスト手順
```

### 2.1 依存パッケージ（requirements.txt）

```text
openpyxl>=3.1.0
pytest>=7.0.0
```

- **openpyxl**：Excel（.xlsx）の読み書き（export コマンド用）
- **pytest**：単体テスト実行

---

## 3. コマンド一覧

| コマンド | 説明 | 例 |
|----------|------|-----|
| （引数なし） | ステップ1動作確認（CSV読み込み・参照整合性） | `python main.py` |
| `calc` | 献立の栄養価計算 | `python main.py calc --date 2026-04-01` |
| `create-menu` | 献立登録（対話式） | `python main.py create-menu --date 2026-04-01` |
| `order` | 発注集計 | `python main.py order --start 2026-04-01 --end 2026-04-30 --people 120` |
| `export` | 献立表Excel出力 | `python main.py export --month 2026-04 --output 献立表.xlsx` |

### 3.1 実行コマンド

```bash
cd /path/to/kyuushoku-system
source venv/bin/activate
python phase1_python/main.py <コマンド> [オプション]
```

---

## 4. 関数・API一覧

### 4.1 データ読み込み

| 関数 | 戻り値 | 説明 |
|------|--------|------|
| `load_csv(filename)` | `list[dict]` | data/ 配下の CSV を読み込み、辞書のリストを返す。空行は除外。 |

### 4.2 参照整合性

| 関数 | 戻り値 | 説明 |
|------|--------|------|
| `validate_reference_integrity(ingredients, dishes, recipe)` | `tuple[bool, list[str]]` | 参照整合性チェック。`(OKか, エラーメッセージのリスト)` を返す。テスト用純粋関数。 |
| `check_reference_integrity(ingredients, dishes, recipe)` | `None` | 上記を呼び、違反があればエラー表示して `sys.exit(1)`。 |

### 4.3 栄養価計算

| 関数 | 戻り値 | 説明 |
|------|--------|------|
| `calc_dish_nutrition(dish_id, recipe, ingredients)` | `dict[str, float]` | 1料理の栄養価。`energy|protein|fat|carbohydrate|salt` |
| `calc_menu_nutrition(menu_row, recipe, ingredients)` | `dict[str, float]` | 献立全体（5料理）の栄養価合計。 |

### 4.4 発注集計

| 関数 | 戻り値 | 説明 |
|------|--------|------|
| `aggregate_order(menus, recipe, start_date, end_date, people)` | `dict[int, float]` | 期間内の献立から食材使用量を集計。`ingredient_id -> 総使用量(g)`。 |

### 4.5 ヘルパー

| 関数 | 戻り値 | 説明 |
|------|--------|------|
| `get_ingredient_by_id(ingredients, ing_id)` | `Optional[dict]` | 食材IDから食材データを取得 |
| `get_dish_by_id(dishes, dish_id)` | `Optional[dict]` | 料理IDから料理データを取得 |
| `get_dishes_by_category(dishes)` | `dict[str, list[dict]]` | 献立候補をカテゴリ別に分類 |

---

## 5. 栄養計算ロジック

### 5.1 計算式

```
料理の栄養価 = Σ（食材の栄養価 × 使用量(g) ÷ 100）
献立の栄養価 = Σ（各料理の栄養価）
```

### 5.2 丸めルール（要件定義書準拠）

| 項目 | 丸め |
|------|------|
| エネルギー、たんぱく質、脂質、炭水化物 | 小数1位（四捨五入） |
| 食塩相当量 | 小数2位（四捨五入） |

### 5.3 目標値（保育園3〜5歳児）

| 項目 | 目標値 |
|------|--------|
| エネルギー | 650 kcal |
| たんぱく質 | 25 g |
| 脂質 | 20 g |
| 炭水化物 | 85 g |
| 食塩相当量 | 3.0 g |

---

## 6. 発注集計ロジック

### 6.1 処理フロー

1. 期間内の献立を取得（`start_date <= date <= end_date`）
2. 各献立の5料理（主食・主菜・副菜・汁物・デザート）のレシピを集計
3. 同一 `ingredient_id` は合算
4. 人数倍率を適用：`総使用量 = 1人分合計 × 人数`
5. 丸め：小数1位

### 6.2 出力形式

```
食材名                       総使用量(g)
----------------------------------------
精白米（炊飯）                    4500.0
豚もも肉（生）                    1600.0
...
```

---

## 7. 献立登録フロー
### 7.1 制約

- 同日登録は禁止（既存日付の場合はエラー終了）
- 5カテゴリ未選択の場合は登録不可

### 7.2 選択順序

主食 → 主菜 → 副菜 → 汁物 → デザート

### 7.3 出力先

`data/menus.csv` に追記。カラム：`date`,`staple_id`,`main_id`,`side_id`,`soup_id`,`dessert_id`,`note`

---

## 8. Excel出力仕様

### 8.1 必須列

| 列 | 内容 |
|----|------|
| 日付 | YYYY-MM-DD |
| 曜日 | 月〜日 |
| 主食 | 料理名 |
| 主菜 | 料理名 |
| 副菜 | 料理名 |
| 汁物 | 料理名 |
| デザート | 料理名 |
| エネルギー | 献立の合計エネルギー（kcal） |

### 8.2 シート名

`YYYY年MM月`（例：2026年04月）

### 8.3 出力パス

相対パス指定の場合はプロジェクトルート基準で解決。

---

## 9. テスト構成

### 9.1 実行方法

```bash
cd /path/to/kyuushoku-system
source venv/bin/activate
python -m pytest phase1_python/test_main.py -v
```

### 9.2 テスト対象

| クラス | テスト内容 |
|--------|------------|
| `TestCalcDishNutrition` | 1料理の栄養価計算、丸めルール |
| `TestCalcMenuNutrition` | 献立全体の栄養価（F2: ±0.1以内） |
| `TestAggregateOrder` | 同一食材合算、人数倍率、期間フィルタ（F4） |
| `TestValidateReferenceIntegrity` | 正常データ、不正 dish_id、不正 ingredient_id |

### 9.3 受入基準との対応

| 基準 | テスト |
|------|--------|
| F2: テスト献立で期待値±0.1以内 | `TestCalcMenuNutrition` |
| F4: 同一食材合算・人数倍率 | `TestAggregateOrder` |

---

## 10. 定数・設定値

| 定数 | 値 | 説明 |
|------|-----|------|
| `PROJECT_ROOT` | phase1_python の親ディレクトリ | データ・出力パスの基準 |
| `DATA_DIR` | `PROJECT_ROOT/data` | CSV の読み込み先 |
| `TARGET_ENERGY` | 650 | 目標エネルギー（kcal） |
| `TARGET_PROTEIN` | 25 | 目標たんぱく質（g） |
| `TARGET_FAT` | 20 | 目標脂質（g） |
| `TARGET_CARBOHYDRATE` | 85 | 目標炭水化物（g） |
| `TARGET_SALT` | 3.0 | 目標食塩相当量（g） |
| `MENU_CATEGORIES` | 主食,主菜,副菜,汁物,デザート | 献立カテゴリ（選択順） |
| `WEEKDAY_JA` | 月〜日 | 曜日表示用 |

---

## 11. 関連ドキュメント

| ドキュメント | 内容 |
|--------------|------|
| 要件定義書 | 機能要件、受入基準 |
| アーキテクチャ設計書 | フェーズ構成、技術スタック |
| データ設計書 | CSV スキーマ詳細 |
| 開発手順_初心者向け | ステップ別実装手順 |

---

## 12. 改訂履歴

| 版数 | 日付 | 内容 |
|------|------|------|
| 1.0 | 2026年3月3日 | 初版作成。Phase1 全機能（calc, create-menu, order, export）およびテストの実装仕様を記載 |
