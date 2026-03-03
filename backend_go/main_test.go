// Phase2 単体テスト: 栄養価計算・発注集計・参照整合性
package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	// テスト時はプロジェクトルートの data/menu.db を参照
	cwd, _ := os.Getwd()
	candidates := []string{
		filepath.Join(cwd, "data", "menu.db"),
		filepath.Join(cwd, "..", "data", "menu.db"),
		filepath.Join(cwd, "..", "..", "data", "menu.db"),
	}
	for _, p := range candidates {
		abs, _ := filepath.Abs(p)
		if _, err := os.Stat(abs); err == nil {
			db, err := sql.Open("sqlite", abs)
			if err != nil {
				t.Fatalf("DB open failed: %v", err)
			}
			return db
		}
	}
	t.Skip("data/menu.db が見つかりません。tools/migrate_csv_to_sqlite.py を実行してください")
	return nil
}

// TestCalcDishNutrition 1料理の栄養価計算
func TestCalcDishNutrition(t *testing.T) {
	db := openTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	nut, err := calcDishNutrition(db, 1)
	if err != nil {
		t.Fatalf("calcDishNutrition: %v", err)
	}
	// dish_id=1: ご飯（精白米150g）
	if nut.Energy != 252.0 {
		t.Errorf("energy: got %.1f, want 252.0", nut.Energy)
	}
	if nut.Protein != 3.8 {
		t.Errorf("protein: got %.1f, want 3.8", nut.Protein)
	}
	if nut.Carbohydrate != 55.6 {
		t.Errorf("carbohydrate: got %.1f, want 55.6", nut.Carbohydrate)
	}
}

// TestCalcDishNutrition_Rounding 丸めルール
func TestCalcDishNutrition_Rounding(t *testing.T) {
	db := openTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	nut, err := calcDishNutrition(db, 1)
	if err != nil {
		t.Fatalf("calcDishNutrition: %v", err)
	}
	if nut.Energy != round1(nut.Energy) {
		t.Errorf("energy should be rounded to 1 decimal")
	}
	if nut.Salt != round2(nut.Salt) {
		t.Errorf("salt should be rounded to 2 decimals")
	}
}

// TestCalcMenuNutrition 献立全体の栄養価（F2: ±0.1）
func TestCalcMenuNutrition(t *testing.T) {
	db := openTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	// テスト献立: ご飯, 豚の生姜焼き, ほうれん草のお浸し, 味噌汁, りんご
	nut, err := calcMenuNutrition(db, 1, 4, 7, 9, 11)
	if err != nil {
		t.Fatalf("calcMenuNutrition: %v", err)
	}
	if nut.Energy < 615.5 || nut.Energy > 615.7 {
		t.Errorf("energy: got %.1f, want 615.5-615.7", nut.Energy)
	}
	if nut.Protein < 27.3 || nut.Protein > 27.5 {
		t.Errorf("protein: got %.1f, want 27.3-27.5", nut.Protein)
	}
	if nut.Fat < 17.7 || nut.Fat > 17.9 {
		t.Errorf("fat: got %.1f, want 17.7-17.9", nut.Fat)
	}
	if nut.Carbohydrate < 85.2 || nut.Carbohydrate > 85.4 {
		t.Errorf("carbohydrate: got %.1f, want 85.2-85.4", nut.Carbohydrate)
	}
	if nut.Salt < 4.31 || nut.Salt > 4.33 {
		t.Errorf("salt: got %.2f, want 4.31-4.33", nut.Salt)
	}
}

// TestAggregateOrder_同一食材合算 同一食材は正しく合算される
func TestAggregateOrder_同一食材合算(t *testing.T) {
	db := openTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	result, err := aggregateOrder(db, "2026-04-01", "2026-04-03", 1, false)
	if err != nil {
		t.Fatalf("aggregateOrder: %v", err)
	}
	// 精白米は3日ともご飯(150g) → 450g
	if result[1] != 450.0 {
		t.Errorf("ingredient 1 (精白米): got %.1f, want 450.0", result[1])
	}
}

// TestAggregateOrder_人数倍率 人数倍率が正しく反映される（F4）
func TestAggregateOrder_人数倍率(t *testing.T) {
	db := openTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	result, err := aggregateOrder(db, "2026-04-01", "2026-04-03", 10, false)
	if err != nil {
		t.Fatalf("aggregateOrder: %v", err)
	}
	if result[1] != 4500.0 {
		t.Errorf("精白米 10人分: got %.1f, want 4500.0", result[1])
	}
	if result[4] != 1600.0 {
		t.Errorf("豚もも肉 10人分: got %.1f, want 1600.0", result[4])
	}
}

// TestAggregateOrder_期間外 指定期間外の献立は含まれない
func TestAggregateOrder_期間外(t *testing.T) {
	db := openTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	result, err := aggregateOrder(db, "2026-04-02", "2026-04-02", 1, false)
	if err != nil {
		t.Fatalf("aggregateOrder: %v", err)
	}
	if result[1] != 150.0 {
		t.Errorf("4/2のみ ご飯150g: got %.1f, want 150.0", result[1])
	}
}

// TestValidateReferenceIntegrity 参照整合性（recipe の dish_id, ingredient_id が有効）
func TestValidateReferenceIntegrity(t *testing.T) {
	db := openTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ok, err := validateReferenceIntegrity(db)
	if err != nil {
		t.Fatalf("validateReferenceIntegrity: %v", err)
	}
	if !ok {
		t.Error("参照整合性: 正常データでは OK であるべき")
	}
}
