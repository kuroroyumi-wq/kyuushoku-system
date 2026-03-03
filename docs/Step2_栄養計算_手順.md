# Step2：栄養計算ロジック実装 詳細手順

---

## 1. やること一覧

| No. | 内容 |
|-----|------|
| 1 | `calc_dish_nutrition()` を実装（1料理の栄養価計算） |
| 2 | `calc_menu_nutrition()` を実装（献立全体の栄養価計算） |
| 3 | `argparse` で `calc --date` コマンドを追加 |
| 4 | 目標値との比較表示 |
| 5 | 動作確認用のテスト献立を登録 |

---

## 2. 計算式（要件定義書準拠）

```
料理栄養価 ＝ Σ（食材栄養価 × 使用量 ÷ 100）
献立栄養価 ＝ 料理栄養価の合計
```

### 丸めルール

| 項目 | 桁 | 方法 |
|------|-----|------|
| エネルギー | 小数1位 | `round(value, 1)` |
| たんぱく質 | 小数1位 | `round(value, 1)` |
| 脂質 | 小数1位 | `round(value, 1)` |
| 炭水化物 | 小数1位 | `round(value, 1)` |
| 食塩相当量 | 小数2位 | `round(value, 2)` |

---

## 3. 実装手順

### 手順①：ingredients を ID で検索できるようにする

CSV の `id` は文字列なので、検索用の辞書を作成する。

```python
def get_ingredient_by_id(ingredients: list[dict], ing_id: int) -> dict | None:
    """食材IDから食材データを取得"""
    for i in ingredients:
        if int(i["id"]) == ing_id:
            return i
    return None
```

### 手順②：calc_dish_nutrition を実装

```python
def calc_dish_nutrition(
    dish_id: int,
    recipe: list[dict],
    ingredients: list[dict],
) -> dict[str, float]:
    """
    1料理の栄養価を計算
    戻り値: {"energy": 0, "protein": 0, "fat": 0, "carbohydrate": 0, "salt": 0}
    """
    result = {"energy": 0.0, "protein": 0.0, "fat": 0.0, "carbohydrate": 0.0, "salt": 0.0}
    nut_keys = ["energy", "protein", "fat", "carbohydrate", "salt"]

    for row in recipe:
        if int(row["dish_id"]) != dish_id:
            continue
        ing_id = int(row["ingredient_id"])
        amount = float(row["amount"])
        ing = get_ingredient_by_id(ingredients, ing_id)
        if ing is None:
            continue
        for key in nut_keys:
            result[key] += float(ing[key]) * amount / 100

    # 丸め適用
    result["energy"] = round(result["energy"], 1)
    result["protein"] = round(result["protein"], 1)
    result["fat"] = round(result["fat"], 1)
    result["carbohydrate"] = round(result["carbohydrate"], 1)
    result["salt"] = round(result["salt"], 2)
    return result
```

### 手順③：calc_menu_nutrition を実装

```python
def calc_menu_nutrition(
    menu_row: dict,
    dishes: list[dict],
    recipe: list[dict],
    ingredients: list[dict],
) -> dict[str, float]:
    """献立全体の栄養価を計算"""
    dish_ids = [
        int(menu_row["staple_id"]),
        int(menu_row["main_id"]),
        int(menu_row["side_id"]),
        int(menu_row["soup_id"]),
        int(menu_row["dessert_id"]),
    ]
    result = {"energy": 0.0, "protein": 0.0, "fat": 0.0, "carbohydrate": 0.0, "salt": 0.0}
    nut_keys = ["energy", "protein", "fat", "carbohydrate", "salt"]

    for dish_id in dish_ids:
        dish_nut = calc_dish_nutrition(dish_id, recipe, ingredients)
        for key in nut_keys:
            result[key] += dish_nut[key]

    # 丸め適用（合計後）
    result["energy"] = round(result["energy"], 1)
    result["protein"] = round(result["protein"], 1)
    result["fat"] = round(result["fat"], 1)
    result["carbohydrate"] = round(result["carbohydrate"], 1)
    result["salt"] = round(result["salt"], 2)
    return result
```

### 手順④：argparse で calc コマンドを追加

`main.py` の先頭に `import argparse` を追加。

```python
def cmd_calc(args, ingredients, dishes, recipe, menus) -> None:
    """calc --date の処理"""
    date_str = args.date
    menu_row = None
    for m in menus:
        if m["date"] == date_str:
            menu_row = m
            break
    if menu_row is None:
        print(f"エラー: {date_str} の献立が登録されていません")
        sys.exit(1)
    nut = calc_menu_nutrition(menu_row, dishes, recipe, ingredients)
    # 目標値
    targets = {"energy": 650, "protein": 25, "fat": 20, "carbohydrate": 85, "salt": 3.0}
    labels = {"energy": "エネルギー", "protein": "たんぱく質", "fat": "脂質",
              "carbohydrate": "炭水化物", "salt": "食塩相当量"}
    units = {"energy": "kcal", "protein": "g", "fat": "g", "carbohydrate": "g", "salt": "g"}
    for key in ["energy", "protein", "fat", "carbohydrate", "salt"]:
        v = nut[key]
        t = targets[key]
        diff = v - t
        diff_str = f"+{diff}" if diff >= 0 else str(diff)
        if key in ["energy", "protein"]:
            print(f"{labels[key]}：{v} {units[key]}（目標{t}{units[key]}：{diff_str}）")
        else:
            print(f"{labels[key]}：{v} {units[key]}")
```

### 手順⑤：main を argparse 対応に変更

```python
def main() -> None:
    parser = argparse.ArgumentParser(description="管理栄養士業務システム")
    subparsers = parser.add_subparsers(dest="command", help="コマンド")

    # calc コマンド
    calc_parser = subparsers.add_parser("calc", help="栄養価計算")
    calc_parser.add_argument("--date", required=True, help="日付 (YYYY-MM-DD)")
    calc_parser.set_defaults(func=lambda a: run_calc(a))

    args = parser.parse_args()

    if args.command == "calc":
        ingredients = load_csv("ingredients.csv")
        dishes = load_csv("dishes.csv")
        recipe = load_csv("recipe.csv")
        menus = load_csv("menus.csv")
        check_reference_integrity(ingredients, dishes, recipe)
        cmd_calc(args, ingredients, dishes, recipe, menus)
    else:
        # 引数なしの場合はステップ1の動作確認
        run_step1_check()
```

---

## 4. テスト献立の準備

`calc --date` を試すには、menus.csv に献立が必要です。

**方法A：手動で1行追加**

`data/menus.csv` に以下を追加（既存の空行を削除して1行にする）：

```
date,staple_id,main_id,side_id,soup_id,dessert_id,note
2026-04-01,1,4,7,9,11,テスト用
```

- 1=ご飯、4=豚の生姜焼き、7=ほうれん草のお浸し、9=味噌汁、11=りんご

**方法B：Step3（献立登録）を先に実装**

対話式で献立を登録してから calc を実行する。

---

## 5. 動作確認

```bash
python3 phase1_python/main.py calc --date 2026-04-01
```

**期待される出力例：**

```
エネルギー：612.4 kcal（目標650kcal：-37.6）
たんぱく質：24.3 g（目標25g：-0.7）
脂質：19.8 g
炭水化物：87.1 g
食塩相当量：2.8 g
```

※実際の値はレシピ・食材データにより異なります。

---

## 6. 完了の目安

- [ ] `calc_dish_nutrition` が正しく計算する
- [ ] `calc_menu_nutrition` が5料理分を合算する
- [ ] 丸めルールが適用されている（エネルギー等は小数1位、食塩は小数2位）
- [ ] `calc --date 2026-04-01` がエラーなく実行できる
- [ ] 献立が存在しない日付でエラーメッセージが表示される
