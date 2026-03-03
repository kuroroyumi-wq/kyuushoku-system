# 管理栄養士業務システム API 仕様書

**版数**：1.0  
**作成日**：2026年3月3日  
**対象**：Phase3 REST API

---

## 1. 概要

| 項目 | 内容 |
|------|------|
| ベースURL | `http://localhost:8080` |
| 形式 | JSON（export は Excel バイナリ） |
| CORS | 全オリジン許可 |

---

## 2. エンドポイント一覧

| メソッド | パス | 説明 |
|----------|------|------|
| GET | /api/dishes | 献立候補一覧 |
| GET | /api/menus | 献立一覧（月）または単件（日付） |
| POST | /api/menus | 献立登録 |
| PUT | /api/menus | 献立更新 |
| GET | /api/calc | 栄養価計算 |
| GET | /api/order | 発注集計 |
| GET | /api/order/bulk | まとめ買い目安 |
| GET | /api/export | 献立表Excelダウンロード |

---

## 3. 詳細

### GET /api/dishes

献立候補（料理）の一覧を取得。

**クエリ**
| パラメータ | 型 | 必須 | 説明 |
|------------|-----|------|------|
| category | string | - | 主食/主菜/副菜/汁物/デザート で絞り込み |

**レスポンス例**
```json
[
  {"id": 1, "name": "ご飯", "menu_category": "主食", "serving_size": 150, "note": "1人分茶碗1杯"},
  {"id": 4, "name": "豚の生姜焼き", "menu_category": "主菜", "serving_size": 100, "note": "豚肉80g＋調味料"}
]
```

---

### GET /api/menus

**クエリ（いずれか必須）**
| パラメータ | 型 | 説明 |
|------------|-----|------|
| month | string | YYYY-MM 形式。該当月の献立一覧 |
| date | string | YYYY-MM-DD 形式。該当日の献立1件 |

**month 指定時のレスポンス例**
```json
[
  {
    "date": "2026-04-01",
    "weekday": "水",
    "staple": "ご飯",
    "main": "豚の生姜焼き",
    "side": "ほうれん草のお浸し",
    "soup": "味噌汁（豆腐・わかめ）",
    "dessert": "りんご",
    "energy": 615.7,
    "staple_id": 1,
    "main_id": 4,
    "side_id": 7,
    "soup_id": 9,
    "dessert_id": 11
  }
]
```

**date 指定時のレスポンス例**
```json
{
  "date": "2026-04-01",
  "staple_id": 1,
  "main_id": 4,
  "side_id": 7,
  "soup_id": 9,
  "dessert_id": 11,
  "note": ""
}
```

---

### POST /api/menus

献立を登録。

**リクエストボディ**
```json
{
  "date": "2026-04-01",
  "staple_id": 1,
  "main_id": 4,
  "side_id": 7,
  "soup_id": 9,
  "dessert_id": 11
}
```

**成功時（201）**
```json
{"message": "献立を登録しました", "date": "2026-04-01"}
```

**エラー（409）** 同日登録済み
```json
{"error": "2026-04-01 は既に登録されています"}
```

---

### PUT /api/menus

献立を更新。指定日の献立が存在する場合のみ更新可能。

**リクエストボディ**（POST と同じ）
```json
{
  "date": "2026-04-01",
  "staple_id": 1,
  "main_id": 4,
  "side_id": 7,
  "soup_id": 9,
  "dessert_id": 11
}
```

**成功時（200）**
```json
{"message": "献立を更新しました", "date": "2026-04-01"}
```

**エラー（404）** 献立未登録
```json
{"error": "2026-04-01 の献立が登録されていません"}
```

---

### GET /api/calc

指定日の献立の栄養価を計算。

**クエリ**
| パラメータ | 型 | 必須 | 説明 |
|------------|-----|------|------|
| date | string | ○ | YYYY-MM-DD |

**レスポンス例**
```json
{
  "date": "2026-04-01",
  "energy": 615.7,
  "protein": 27.4,
  "fat": 17.8,
  "carbohydrate": 85.4,
  "salt": 4.3,
  "targets": {
    "energy": 650,
    "protein": 25,
    "fat": 20,
    "carbohydrate": 85,
    "salt": 3
  }
}
```

---

### GET /api/order

発注集計を取得。start=end で1日分の集計が可能。

**クエリ**
| パラメータ | 型 | 必須 | 説明 |
|------------|-----|------|------|
| start | string | ○ | 開始日 YYYY-MM-DD |
| end | string | ○ | 終了日 YYYY-MM-DD |
| people | int | ○ | 人数 |
| exclude_condiments | string | - | 1 のとき調味料を除外 |

**レスポンス例**
```json
{
  "start": "2026-04-01",
  "end": "2026-04-30",
  "people": 120,
  "items": [
    {"ingredient_id": 1, "name": "精白米（炊飯）", "total_g": 54000},
    {"ingredient_id": 4, "name": "豚もも肉（生）", "total_g": 19200}
  ]
}
```

---

### GET /api/order/bulk

まとめ買い目安を取得。米・調味料・乾物など、まとめて発注する食材の使用量と発注目安を返す。

**クエリ**
| パラメータ | 型 | 必須 | 説明 |
|------------|-----|------|------|
| start | string | ○ | 開始日 YYYY-MM-DD |
| end | string | ○ | 終了日 YYYY-MM-DD |
| people | int | ○ | 人数 |

**レスポンス例**
```json
{
  "start": "2026-04-01",
  "end": "2026-04-30",
  "people": 120,
  "items": [
    {
      "ingredient_id": 1,
      "name": "精白米（炊飯）",
      "total_g": 54000,
      "order_unit_g": 5000,
      "order_unit_name": "5kg袋",
      "bulk_category": "米・主食",
      "order_qty": 11
    }
  ]
}
```

---

### GET /api/export

献立表Excelをダウンロード。

**クエリ**
| パラメータ | 型 | 必須 | 説明 |
|------------|-----|------|------|
| month | string | ○ | YYYY-MM |

**レスポンス**
- Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
- Content-Disposition: attachment; filename="献立表_2026年04月.xlsx"
- バイナリ（Excelファイル）

---

## 4. エラー形式

```json
{"error": "エラーメッセージ"}
```

---

## 5. 改訂履歴

| 版数 | 日付 | 内容 |
|------|------|------|
| 1.0 | 2026年3月3日 | 初版作成 |
