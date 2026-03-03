-- 管理栄養士業務システム SQLite スキーマ
-- schema_version: 1.0
-- データ設計書・schema.json に準拠

-- schema_version 管理（開発計画書v2.0）
CREATE TABLE IF NOT EXISTS version (
    schema_version TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- 食材マスタ
CREATE TABLE IF NOT EXISTS ingredients (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    ingredient_category TEXT NOT NULL,
    energy REAL NOT NULL,
    protein REAL NOT NULL,
    fat REAL NOT NULL,
    carbohydrate REAL NOT NULL,
    salt REAL NOT NULL,
    unit TEXT NOT NULL,
    waste_rate REAL NOT NULL DEFAULT 0,
    note TEXT
);

-- 献立候補（料理マスタ）
CREATE TABLE IF NOT EXISTS dishes (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    menu_category TEXT NOT NULL,
    serving_size INTEGER NOT NULL,
    note TEXT
);

-- レシピ（料理と食材の対応）
CREATE TABLE IF NOT EXISTS recipe (
    dish_id INTEGER NOT NULL REFERENCES dishes(id),
    ingredient_id INTEGER NOT NULL REFERENCES ingredients(id),
    amount REAL NOT NULL,
    PRIMARY KEY (dish_id, ingredient_id)
);

-- 献立
CREATE TABLE IF NOT EXISTS menus (
    date TEXT PRIMARY KEY,
    staple_id INTEGER NOT NULL REFERENCES dishes(id),
    main_id INTEGER NOT NULL REFERENCES dishes(id),
    side_id INTEGER NOT NULL REFERENCES dishes(id),
    soup_id INTEGER NOT NULL REFERENCES dishes(id),
    dessert_id INTEGER NOT NULL REFERENCES dishes(id),
    note TEXT
);

-- インデックス
CREATE INDEX IF NOT EXISTS idx_recipe_dish ON recipe(dish_id);
CREATE INDEX IF NOT EXISTS idx_recipe_ingredient ON recipe(ingredient_id);
CREATE INDEX IF NOT EXISTS idx_menus_date ON menus(date);
