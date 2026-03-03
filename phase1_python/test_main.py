#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
管理栄養士業務システム - 単体テスト
Step6: 栄養価計算・発注集計・参照整合性チェック
"""

import os
import sys
import pytest

# プロジェクトルートをパスに追加（main.py の DATA_DIR は絶対パスなので cwd は不要）
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from main import (
    load_csv,
    validate_reference_integrity,
    calc_dish_nutrition,
    calc_menu_nutrition,
    aggregate_order,
)


# --- フィクスチャ: 実データを読み込む ---
@pytest.fixture
def ingredients():
    return load_csv("ingredients.csv")


@pytest.fixture
def dishes():
    return load_csv("dishes.csv")


@pytest.fixture
def recipe():
    return load_csv("recipe.csv")


@pytest.fixture
def menus():
    return load_csv("menus.csv")


# --- 栄養価計算テスト ---
class TestCalcDishNutrition:
    """1料理の栄養価計算"""

    def test_ご飯(self, recipe, ingredients):
        """dish_id=1: ご飯（精白米150g）"""
        nut = calc_dish_nutrition(1, recipe, ingredients)
        assert nut["energy"] == 252.0  # 168 * 150 / 100
        assert nut["protein"] == 3.8
        assert nut["carbohydrate"] == 55.6  # 37.1*150/100 → round 55.6

    def test_丸めルール(self, recipe, ingredients):
        """エネルギー等は小数1位、食塩は小数2位で丸め"""
        nut = calc_dish_nutrition(1, recipe, ingredients)
        assert isinstance(nut["energy"], float)
        assert isinstance(nut["salt"], float)
        # 小数点以下が正しく丸められていることを確認
        assert nut["energy"] == round(nut["energy"], 1)
        assert nut["salt"] == round(nut["salt"], 2)


class TestCalcMenuNutrition:
    """献立全体の栄養価計算（F2: 受入基準 ±0.1）"""

    def test_テスト献立_2026_04_01(self, recipe, ingredients):
        """テスト献立: ご飯, 豚の生姜焼き, ほうれん草のお浸し, 味噌汁, りんご"""
        menu_row = {
            "date": "2026-04-01",
            "staple_id": "1",
            "main_id": "4",
            "side_id": "7",
            "soup_id": "9",
            "dessert_id": "11",
        }
        nut = calc_menu_nutrition(menu_row, recipe, ingredients)
        # 目標値±0.1以内（要件定義書 F2）
        assert 615.5 <= nut["energy"] <= 615.7
        assert 27.3 <= nut["protein"] <= 27.5
        assert 17.7 <= nut["fat"] <= 17.9
        assert 85.2 <= nut["carbohydrate"] <= 85.4
        assert 4.31 <= nut["salt"] <= 4.33


# --- 発注集計テスト（F4: 同一食材合算・人数倍率） ---
class TestAggregateOrder:
    """発注集計ロジック"""

    def test_同一食材が合算される(self, menus, recipe):
        """同一食材は正しく合算される"""
        # 2026-04-01〜03 の3日分、1人分
        result = aggregate_order(menus, recipe, "2026-04-01", "2026-04-03", 1)
        # 精白米は3日ともご飯(150g) → 450g
        assert 1 in result
        assert result[1] == 450.0

    def test_人数倍率が適用される(self, menus, recipe):
        """人数倍率が正しく反映される（F4）"""
        result = aggregate_order(menus, recipe, "2026-04-01", "2026-04-03", 10)
        # 精白米 450g/人 × 10人 = 4500g
        assert result[1] == 4500.0
        # 豚もも肉: 4日×80g×10人 = 1600g（4/1,4/2は豚肉、4/3は卵のみ）
        assert result[4] == 1600.0

    def test_期間外の献立は含まれない(self, menus, recipe):
        """指定期間外の献立は集計に含まれない"""
        result = aggregate_order(menus, recipe, "2026-04-02", "2026-04-02", 1)
        # 4/2のみ: ご飯150g
        assert result[1] == 150.0


# --- 参照整合性テスト ---
class TestValidateReferenceIntegrity:
    """参照整合性チェック"""

    def test_正常データではOK(self, ingredients, dishes, recipe):
        """整合性が取れたデータでは OK"""
        ok, errors = validate_reference_integrity(ingredients, dishes, recipe)
        assert ok is True
        assert len(errors) == 0

    def test_存在しないdish_idでエラー(self, ingredients, dishes):
        """recipe に存在しない dish_id があるとエラー"""
        invalid_recipe = [{"dish_id": "999", "ingredient_id": "1", "amount": "100"}]
        ok, errors = validate_reference_integrity(ingredients, dishes, invalid_recipe)
        assert ok is False
        assert any("dish_id=999" in e for e in errors)

    def test_存在しないingredient_idでエラー(self, ingredients, dishes):
        """recipe に存在しない ingredient_id があるとエラー"""
        invalid_recipe = [{"dish_id": "1", "ingredient_id": "999", "amount": "100"}]
        ok, errors = validate_reference_integrity(ingredients, dishes, invalid_recipe)
        assert ok is False
        assert any("ingredient_id=999" in e for e in errors)
