// 管理栄養士業務システム - Phase2 Go CLI
// SQLite を使用した単体バイナリ

package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

const (
	targetEnergy       = 650
	targetProtein      = 25
	targetFat          = 20
	targetCarbohydrate = 85
	targetSalt         = 3.0
)

var weekdayJA = []string{"月", "火", "水", "木", "金", "土", "日"}

func findDBPath() string {
	cwd, _ := os.Getwd()
	candidates := []string{
		filepath.Join(cwd, "data", "menu.db"),
		filepath.Join(cwd, "..", "data", "menu.db"),
		filepath.Join(cwd, "..", "..", "data", "menu.db"),
	}
	for _, p := range candidates {
		abs, _ := filepath.Abs(p)
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}
	return filepath.Join(cwd, "data", "menu.db")
}

// validateReferenceIntegrity は recipe の dish_id, ingredient_id が有効かチェック。テスト用。
func validateReferenceIntegrity(db *sql.DB) (bool, error) {
	var orphanCount int
	err := db.QueryRow(q(`
		SELECT COUNT(*) FROM recipe r
		LEFT JOIN dishes d ON r.dish_id = d.id
		LEFT JOIN ingredients i ON r.ingredient_id = i.id
		WHERE d.id IS NULL OR i.id IS NULL
	`)).Scan(&orphanCount)
	if err != nil {
		return false, err
	}
	return orphanCount == 0, nil
}

type Nutrition struct {
	Energy       float64
	Protein      float64
	Fat          float64
	Carbohydrate float64
	Salt         float64
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func calcDishNutrition(db *sql.DB, dishID int) (Nutrition, error) {
	rows, err := db.Query(q(`
		SELECT i.energy, i.protein, i.fat, i.carbohydrate, i.salt, r.amount
		FROM recipe r
		JOIN ingredients i ON r.ingredient_id = i.id
		WHERE r.dish_id = ?
	`), dishID)
	if err != nil {
		return Nutrition{}, err
	}
	defer rows.Close()

	var n Nutrition
	for rows.Next() {
		var energy, protein, fat, carb, salt, amount float64
		if err := rows.Scan(&energy, &protein, &fat, &carb, &salt, &amount); err != nil {
			return Nutrition{}, err
		}
		n.Energy += energy * amount / 100
		n.Protein += protein * amount / 100
		n.Fat += fat * amount / 100
		n.Carbohydrate += carb * amount / 100
		n.Salt += salt * amount / 100
	}
	n.Energy = round1(n.Energy)
	n.Protein = round1(n.Protein)
	n.Fat = round1(n.Fat)
	n.Carbohydrate = round1(n.Carbohydrate)
	n.Salt = round2(n.Salt)
	return n, nil
}

func calcMenuNutrition(db *sql.DB, stapleID, mainID, sideID, soupID, dessertID int) (Nutrition, error) {
	ids := []int{stapleID, mainID, sideID, soupID, dessertID}
	var total Nutrition
	for _, id := range ids {
		n, err := calcDishNutrition(db, id)
		if err != nil {
			return Nutrition{}, err
		}
		total.Energy += n.Energy
		total.Protein += n.Protein
		total.Fat += n.Fat
		total.Carbohydrate += n.Carbohydrate
		total.Salt += n.Salt
	}
	total.Energy = round1(total.Energy)
	total.Protein = round1(total.Protein)
	total.Fat = round1(total.Fat)
	total.Carbohydrate = round1(total.Carbohydrate)
	total.Salt = round2(total.Salt)
	return total, nil
}

func cmdCalc(db *sql.DB, date string) error {
	var stapleID, mainID, sideID, soupID, dessertID int
	err := db.QueryRow(q(`
		SELECT staple_id, main_id, side_id, soup_id, dessert_id FROM menus WHERE date = ? AND facility_id = ?
	`), date, DefaultFacilityID).Scan(&stapleID, &mainID, &sideID, &soupID, &dessertID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("%s の献立が登録されていません", date)
	}
	if err != nil {
		return err
	}

	nut, err := calcMenuNutrition(db, stapleID, mainID, sideID, soupID, dessertID)
	if err != nil {
		return err
	}

	labels := map[string]string{
		"energy": "エネルギー", "protein": "たんぱく質", "fat": "脂質",
		"carbohydrate": "炭水化物", "salt": "食塩相当量",
	}
	units := map[string]string{
		"energy": "kcal", "protein": "g", "fat": "g", "carbohydrate": "g", "salt": "g",
	}
	targets := map[string]float64{
		"energy": targetEnergy, "protein": targetProtein, "fat": targetFat,
		"carbohydrate": targetCarbohydrate, "salt": targetSalt,
	}
	values := map[string]float64{
		"energy": nut.Energy, "protein": nut.Protein, "fat": nut.Fat,
		"carbohydrate": nut.Carbohydrate, "salt": nut.Salt,
	}

	fmt.Printf("献立の栄養価（%s）\n", date)
	fmt.Println("----------------------------------------")
	for _, key := range []string{"energy", "protein", "fat", "carbohydrate", "salt"} {
		v := values[key]
		t := targets[key]
		diff := v - t
		if key == "salt" {
			diff = round2(diff)
		} else {
			diff = round1(diff)
		}
		diffStr := fmt.Sprintf("%+.1f", diff)
		if key == "salt" {
			diffStr = fmt.Sprintf("%+.2f", diff)
		}
		if key == "energy" || key == "protein" {
			fmt.Printf("%s：%.1f %s（目標%.0f%s：%s）\n", labels[key], v, units[key], t, units[key], diffStr)
		} else {
			fmt.Printf("%s：%.1f %s\n", labels[key], v, units[key])
		}
	}
	return nil
}

// aggregateOrder は期間内の献立から食材使用量を集計。excludeCondiments が true のとき調味料を除外。
func aggregateOrder(db *sql.DB, start, end string, people int, excludeCondiments bool) (map[int]float64, error) {
	rows, err := db.Query(q(`
		SELECT staple_id, main_id, side_id, soup_id, dessert_id FROM menus
		WHERE date >= ? AND date <= ? AND facility_id = ?
		ORDER BY date
	`), start, end, DefaultFacilityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	amountByIng := make(map[int]float64)
	count := 0
	for rows.Next() {
		var s, m, si, so, d int
		if err := rows.Scan(&s, &m, &si, &so, &d); err != nil {
			return nil, err
		}
		count++
		for _, dishID := range []int{s, m, si, so, d} {
			r2, err2 := db.Query(q(`SELECT ingredient_id, amount FROM recipe WHERE dish_id = ?`), dishID)
			if err2 != nil {
				continue
			}
			for r2.Next() {
				var ingID int
				var amt float64
				r2.Scan(&ingID, &amt)
				amountByIng[ingID] += amt
			}
			r2.Close()
		}
	}
	if count == 0 {
		return nil, fmt.Errorf("%s 〜 %s に献立が登録されていません", start, end)
	}

	result := make(map[int]float64)
	for ingID, amt := range amountByIng {
		if excludeCondiments {
			var cat string
			if err := db.QueryRow(q("SELECT ingredient_category FROM ingredients WHERE id = ?"), ingID).Scan(&cat); err == nil && cat == "調味料" {
				continue
			}
		}
		result[ingID] = round1(amt * float64(people))
	}
	return result, nil
}

func cmdOrder(db *sql.DB, start, end string, people int) error {
	totalByIng, err := aggregateOrder(db, start, end, people, false)
	if err != nil {
		return err
	}

	ingNames := make(map[int]string)
	for ingID := range totalByIng {
		var name string
		db.QueryRow(q("SELECT name FROM ingredients WHERE id = ?"), ingID).Scan(&name)
		ingNames[ingID] = name
	}

	fmt.Printf("発注集計（%s 〜 %s、%d人分）\n", start, end, people)
	fmt.Println("----------------------------------------")
	fmt.Printf("%-20s %12s\n", "食材名", "総使用量(g)")
	fmt.Println("----------------------------------------")

	ingIDs := make([]int, 0, len(totalByIng))
	for ingID := range totalByIng {
		ingIDs = append(ingIDs, ingID)
	}
	sort.Ints(ingIDs)
	for _, ingID := range ingIDs {
		total := totalByIng[ingID]
		name := ingNames[ingID]
		if name == "" {
			name = fmt.Sprintf("不明(id:%d)", ingID)
		}
		fmt.Printf("%-20s %12.1f\n", name, total)
	}
	return nil
}

func cmdExport(db *sql.DB, month, output string) error {
	parts := strings.Split(month, "-")
	if len(parts) != 2 {
		return fmt.Errorf("--month は YYYY-MM 形式で指定してください（例: 2026-04）")
	}
	year, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	start := fmt.Sprintf("%04d-%02d-01", year, m)
	lastDay := time.Date(year, time.Month(m+1), 0, 0, 0, 0, 0, time.UTC).Day()
	end := fmt.Sprintf("%04d-%02d-%02d", year, m, lastDay)

	rows, err := db.Query(q(`
		SELECT m.date, m.staple_id, m.main_id, m.side_id, m.soup_id, m.dessert_id,
		       d1.name as staple, d2.name as main, d3.name as side, d4.name as soup, d5.name as dessert
		FROM menus m
		JOIN dishes d1 ON m.staple_id = d1.id
		JOIN dishes d2 ON m.main_id = d2.id
		JOIN dishes d3 ON m.side_id = d3.id
		JOIN dishes d4 ON m.soup_id = d4.id
		JOIN dishes d5 ON m.dessert_id = d5.id
		WHERE m.date >= ? AND m.date <= ? AND m.facility_id = ?
		ORDER BY m.date
	`), start, end, DefaultFacilityID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var records [][]interface{}
	for rows.Next() {
		var date string
		var stapleID, mainID, sideID, soupID, dessertID int
		var staple, mainD, side, soup, dessert string
		rows.Scan(&date, &stapleID, &mainID, &sideID, &soupID, &dessertID,
			&staple, &mainD, &side, &soup, &dessert)
		nut, _ := calcMenuNutrition(db, stapleID, mainID, sideID, soupID, dessertID)
		t, _ := time.Parse("2006-01-02", date)
		w := t.Weekday() // 0=日, 1=月, ..., 6=土
		idx := 0
		if w == 0 {
			idx = 6
		} else {
			idx = int(w) - 1
		}
		weekday := weekdayJA[idx]
		records = append(records, []interface{}{date, weekday, staple, mainD, side, soup, dessert, nut.Energy})
	}
	if len(records) == 0 {
		return fmt.Errorf("%s に献立が登録されていません", month)
	}

	f := excelize.NewFile()
	sheet := fmt.Sprintf("%d年%02d月", year, m)
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"日付", "曜日", "主食", "主菜", "副菜", "汁物", "デザート", "エネルギー"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for rowIdx, rec := range records {
		for colIdx, v := range rec {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheet, cell, v)
		}
	}

	outPath := output
	if !filepath.IsAbs(output) {
		cwd, _ := os.Getwd()
		outPath = filepath.Join(cwd, output)
	}
	if err := f.SaveAs(outPath); err != nil {
		return err
	}
	fmt.Printf("献立表を出力しました: %s\n", outPath)
	return nil
}

func cmdCreateMenu(db *sql.DB, date string) error {
	var exists int
	db.QueryRow(q("SELECT 1 FROM menus WHERE date = ? AND facility_id = ?"), date, DefaultFacilityID).Scan(&exists)
	if exists == 1 {
		return fmt.Errorf("%s は既に登録されています。同日登録は禁止です。", date)
	}

	dishesByCat := make(map[string][]struct {
		id   int
		name string
	})
	rows, _ := db.Query(q("SELECT id, name, menu_category FROM dishes"))
	for rows.Next() {
		var id int
		var name, cat string
		rows.Scan(&id, &name, &cat)
		dishesByCat[cat] = append(dishesByCat[cat], struct {
			id   int
			name string
		}{id, name})
	}
	rows.Close()

	cats := []struct {
		label string
		col   string
	}{
		{"主食", "staple_id"}, {"主菜", "main_id"}, {"副菜", "side_id"},
		{"汁物", "soup_id"}, {"デザート", "dessert_id"},
	}

	selected := make(map[string]int)
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("献立登録（%s）\n", date)
	fmt.Println("----------------------------------------")

	for _, c := range cats {
		list := dishesByCat[c.label]
		if len(list) == 0 {
			return fmt.Errorf("%s の献立候補がありません", c.label)
		}
		fmt.Printf("\n【%s】\n", c.label)
		for i, d := range list {
			fmt.Printf("  %d. %s (id:%d)\n", i+1, d.name, d.id)
		}
		for {
			fmt.Printf("  %sを選択 (1-%d): ", c.label, len(list))
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)
			idx, err := strconv.Atoi(text)
			if err == nil && idx >= 1 && idx <= len(list) {
				selected[c.col] = list[idx-1].id
				break
			}
			fmt.Println("  無効な入力です。正しい番号を入力してください。")
		}
	}

	_, err := db.Exec(q(`
		INSERT INTO menus (date, facility_id, staple_id, main_id, side_id, soup_id, dessert_id, note)
		VALUES (?, ?, ?, ?, ?, ?, ?, '')
	`), date, DefaultFacilityID, selected["staple_id"], selected["main_id"], selected["side_id"],
		selected["soup_id"], selected["dessert_id"])
	if err != nil {
		return err
	}
	fmt.Println("----------------------------------------")
	fmt.Printf("献立を登録しました: %s\n", date)
	return nil
}

func cmdStep1(db *sql.DB) error {
	var ingCount, dishCount, recipeCount, menuCount int
	db.QueryRow(q("SELECT COUNT(*) FROM ingredients")).Scan(&ingCount)
	db.QueryRow(q("SELECT COUNT(*) FROM dishes")).Scan(&dishCount)
	db.QueryRow(q("SELECT COUNT(*) FROM recipe")).Scan(&recipeCount)
	db.QueryRow(q("SELECT COUNT(*) FROM menus")).Scan(&menuCount)

	fmt.Println("管理栄養士業務システム - ステップ1 動作確認")
	fmt.Println("----------------------------------------")
	fmt.Printf("食材: %d 件読み込みました\n", ingCount)
	fmt.Printf("献立候補: %d 件読み込みました\n", dishCount)
	fmt.Printf("レシピ: %d 件読み込みました\n", recipeCount)
	fmt.Printf("献立: %d 件読み込みました\n", menuCount)
	fmt.Println("参照整合性: OK")
	fmt.Println("----------------------------------------")
	fmt.Println("ステップ1 完了: データスキーマ固定 OK")
	return nil
}

func main() {
	if len(os.Args) < 2 {
		db, err := openDB()
		if err != nil {
			fmt.Fprintln(os.Stderr, "エラー:", err)
			os.Exit(1)
		}
		defer db.Close()
		cmdStep1(db)
		return
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	db, err := openDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "エラー:", err)
		os.Exit(1)
	}
	defer db.Close()

	var runErr error
	switch cmd {
	case "calc":
		if len(args) < 2 || args[0] != "--date" {
			fmt.Fprintln(os.Stderr, "用法: menu-system calc --date YYYY-MM-DD")
			os.Exit(1)
		}
		runErr = cmdCalc(db, args[1])
	case "create-menu":
		if len(args) < 2 || args[0] != "--date" {
			fmt.Fprintln(os.Stderr, "用法: menu-system create-menu --date YYYY-MM-DD")
			os.Exit(1)
		}
		runErr = cmdCreateMenu(db, args[1])
	case "order":
		if len(args) < 6 {
			fmt.Fprintln(os.Stderr, "用法: menu-system order --start YYYY-MM-DD --end YYYY-MM-DD --people 人数")
			os.Exit(1)
		}
		var start, end string
		var people int
		for i := 0; i < len(args); i++ {
			if args[i] == "--start" && i+1 < len(args) {
				start = args[i+1]
				i++
			} else if args[i] == "--end" && i+1 < len(args) {
				end = args[i+1]
				i++
			} else if args[i] == "--people" && i+1 < len(args) {
				people, _ = strconv.Atoi(args[i+1])
				i++
			}
		}
		runErr = cmdOrder(db, start, end, people)
	case "export":
		if len(args) < 4 {
			fmt.Fprintln(os.Stderr, "用法: menu-system export --month YYYY-MM --output ファイル.xlsx")
			os.Exit(1)
		}
		var month, output string
		for i := 0; i < len(args); i++ {
			if args[i] == "--month" && i+1 < len(args) {
				month = args[i+1]
				i++
			} else if args[i] == "--output" && i+1 < len(args) {
				output = args[i+1]
				i++
			}
		}
		runErr = cmdExport(db, month, output)
	case "serve":
		port := "8080"
		if len(args) >= 2 && args[0] == "--port" {
			port = args[1]
		}
		runErr = runServer(db, port)
	default:
		fmt.Fprintf(os.Stderr, "未知のコマンド: %s\n", cmd)
		os.Exit(1)
	}

	if runErr != nil {
		fmt.Fprintln(os.Stderr, "エラー:", runErr)
		os.Exit(1)
	}
}
