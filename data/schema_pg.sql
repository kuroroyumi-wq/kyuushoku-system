-- 管理栄養士業務システム PostgreSQL スキーマ
-- schema_version: 2.0
-- Phase4: facilities 追加

-- 施設マスタ（Phase4）
CREATE TABLE IF NOT EXISTS facilities (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    code TEXT UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- schema_version 管理
CREATE TABLE IF NOT EXISTS version (
    schema_version TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 食材マスタ（facility_id NULL = 共通マスタ）
CREATE TABLE IF NOT EXISTS ingredients (
    id SERIAL PRIMARY KEY,
    facility_id INTEGER,
    name TEXT NOT NULL,
    ingredient_category TEXT NOT NULL,
    energy DOUBLE PRECISION NOT NULL,
    protein DOUBLE PRECISION NOT NULL,
    fat DOUBLE PRECISION NOT NULL,
    carbohydrate DOUBLE PRECISION NOT NULL,
    salt DOUBLE PRECISION NOT NULL,
    unit TEXT NOT NULL,
    waste_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
    note TEXT
);

-- 献立候補（料理マスタ、facility_id NULL = 共通）
CREATE TABLE IF NOT EXISTS dishes (
    id SERIAL PRIMARY KEY,
    facility_id INTEGER,
    name TEXT NOT NULL,
    menu_category TEXT NOT NULL,
    serving_size INTEGER NOT NULL,
    note TEXT
);

-- レシピ（料理と食材の対応）
CREATE TABLE IF NOT EXISTS recipe (
    dish_id INTEGER NOT NULL REFERENCES dishes(id),
    ingredient_id INTEGER NOT NULL REFERENCES ingredients(id),
    amount DOUBLE PRECISION NOT NULL,
    PRIMARY KEY (dish_id, ingredient_id)
);

-- 献立（facility_id: 施設別）
CREATE TABLE IF NOT EXISTS menus (
    date DATE NOT NULL,
    facility_id INTEGER NOT NULL DEFAULT 1 REFERENCES facilities(id),
    staple_id INTEGER NOT NULL REFERENCES dishes(id),
    main_id INTEGER NOT NULL REFERENCES dishes(id),
    side_id INTEGER NOT NULL REFERENCES dishes(id),
    soup_id INTEGER NOT NULL REFERENCES dishes(id),
    dessert_id INTEGER NOT NULL REFERENCES dishes(id),
    note TEXT,
    PRIMARY KEY (date, facility_id)
);

-- ユーザー（Phase4 Task3）
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    facility_id INTEGER REFERENCES facilities(id),
    role TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- まとめ買いガイド（発注単位・分類）
CREATE TABLE IF NOT EXISTS bulk_purchase_guide (
    ingredient_id INTEGER PRIMARY KEY REFERENCES ingredients(id),
    order_unit_g DOUBLE PRECISION NOT NULL,
    order_unit_name TEXT NOT NULL,
    bulk_category TEXT NOT NULL
);

-- インデックス
CREATE INDEX IF NOT EXISTS idx_recipe_dish ON recipe(dish_id);
CREATE INDEX IF NOT EXISTS idx_recipe_ingredient ON recipe(ingredient_id);
CREATE INDEX IF NOT EXISTS idx_menus_date ON menus(date);
CREATE INDEX IF NOT EXISTS idx_menus_facility_date ON menus(facility_id, date);
