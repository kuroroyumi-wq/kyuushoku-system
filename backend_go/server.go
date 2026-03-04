// Phase3: REST API サーバ
package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, v interface{}, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, msg string, status int) {
	writeJSON(w, map[string]string{"error": msg}, status)
}

// dbErrorToJapanese は DB エラーをユーザー向け日本語メッセージに変換
func dbErrorToJapanese(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	if strings.Contains(s, "readonly") || strings.Contains(s, "read only") {
		return "データベースが読み取り専用のため、書き込みできません。data/menu.db のファイル権限を確認してください。"
	}
	return s
}

// GET /api/dishes?category=主食
func handleDishes(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		category := r.URL.Query().Get("category")
		var rows *sql.Rows
		var err error
		if category != "" {
			rows, err = db.Query(q("SELECT id, name, menu_category, serving_size, note FROM dishes WHERE menu_category = ? ORDER BY id"), category)
		} else {
			rows, err = db.Query(q("SELECT id, name, menu_category, serving_size, note FROM dishes ORDER BY menu_category, id"))
		}
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var list []map[string]interface{}
		for rows.Next() {
			var id, servingSize int
			var name, menuCategory string
			var note sql.NullString
			rows.Scan(&id, &name, &menuCategory, &servingSize, &note)
			list = append(list, map[string]interface{}{
				"id":            id,
				"name":          name,
				"menu_category": menuCategory,
				"serving_size":  servingSize,
				"note":          note.String,
			})
		}
		writeJSON(w, list, http.StatusOK)
	}
}

// GET /api/menus?month=2026-04 または ?date=2026-04-01
func handleMenus(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		month := r.URL.Query().Get("month")
		dateRaw := r.URL.Query().Get("date")
		date := strings.ReplaceAll(dateRaw, "/", "-")

		if date != "" {
			var stapleID, mainID, sideID, soupID, dessertID int
			var note sql.NullString
			err := db.QueryRow(q(`
				SELECT staple_id, main_id, side_id, soup_id, dessert_id, note FROM menus WHERE date = ? AND facility_id = ?
			`), date, DefaultFacilityID).Scan(&stapleID, &mainID, &sideID, &soupID, &dessertID, &note)
			if err == sql.ErrNoRows {
				writeError(w, "献立が登録されていません", http.StatusNotFound)
				return
			}
			if err != nil {
				writeError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, map[string]interface{}{
				"date":       date,
				"staple_id":  stapleID,
				"main_id":    mainID,
				"side_id":    sideID,
				"soup_id":    soupID,
				"dessert_id": dessertID,
				"note":       note.String,
			}, http.StatusOK)
			return
		}

		if month == "" {
			writeError(w, "month または date を指定してください", http.StatusBadRequest)
			return
		}

		parts := strings.Split(month, "-")
		if len(parts) != 2 {
			writeError(w, "month は YYYY-MM 形式で指定してください", http.StatusBadRequest)
			return
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
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var list []map[string]interface{}
		for rows.Next() {
			var date string
			var stapleID, mainID, sideID, soupID, dessertID int
			var staple, mainD, side, soup, dessert string
			rows.Scan(&date, &stapleID, &mainID, &sideID, &soupID, &dessertID,
				&staple, &mainD, &side, &soup, &dessert)
			nut, _ := calcMenuNutrition(db, stapleID, mainID, sideID, soupID, dessertID)
			t, _ := time.Parse("2006-01-02", date)
			wd := t.Weekday()
			idx := 0
			if wd == 0 {
				idx = 6
			} else {
				idx = int(wd) - 1
			}
			weekday := weekdayJA[idx]
			list = append(list, map[string]interface{}{
				"date":       date,
				"weekday":    weekday,
				"staple":     staple,
				"main":       mainD,
				"side":       side,
				"soup":       soup,
				"dessert":    dessert,
				"energy":     nut.Energy,
				"staple_id":  stapleID,
				"main_id":    mainID,
				"side_id":    sideID,
				"soup_id":    soupID,
				"dessert_id": dessertID,
			})
		}
		writeJSON(w, list, http.StatusOK)
	}
}

// POST /api/menus
func handleCreateMenu(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Date      string `json:"date"`
			StapleID  int    `json:"staple_id"`
			MainID    int    `json:"main_id"`
			SideID    int    `json:"side_id"`
			SoupID    int    `json:"soup_id"`
			DessertID int    `json:"dessert_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if body.Date == "" {
			writeError(w, "date は必須です", http.StatusBadRequest)
			return
		}
		dateNorm := strings.ReplaceAll(body.Date, "/", "-")

		var exists int
		db.QueryRow(q("SELECT 1 FROM menus WHERE date = ? AND facility_id = ?"), dateNorm, DefaultFacilityID).Scan(&exists)
		if exists == 1 {
			writeError(w, dateNorm+" は既に登録されています", http.StatusConflict)
			return
		}

		_, err := db.Exec(q(`
			INSERT INTO menus (date, facility_id, staple_id, main_id, side_id, soup_id, dessert_id, note)
			VALUES (?, ?, ?, ?, ?, ?, ?, '')
		`), dateNorm, DefaultFacilityID, body.StapleID, body.MainID, body.SideID, body.SoupID, body.DessertID)
		if err != nil {
			writeError(w, dbErrorToJapanese(err), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]string{"message": "献立を登録しました", "date": dateNorm}, http.StatusCreated)
	}
}

// PUT /api/menus
func handleUpdateMenu(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Date      string `json:"date"`
			StapleID  int    `json:"staple_id"`
			MainID    int    `json:"main_id"`
			SideID    int    `json:"side_id"`
			SoupID    int    `json:"soup_id"`
			DessertID int    `json:"dessert_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if body.Date == "" {
			writeError(w, "date は必須です", http.StatusBadRequest)
			return
		}
		dateNorm := strings.ReplaceAll(body.Date, "/", "-")

		result, err := db.Exec(q(`
			UPDATE menus SET staple_id=?, main_id=?, side_id=?, soup_id=?, dessert_id=?
			WHERE date = ? AND facility_id = ?
		`), body.StapleID, body.MainID, body.SideID, body.SoupID, body.DessertID, dateNorm, DefaultFacilityID)
		if err != nil {
			writeError(w, dbErrorToJapanese(err), http.StatusInternalServerError)
			return
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			writeError(w, dateNorm+" の献立が登録されていません", http.StatusNotFound)
			return
		}
		writeJSON(w, map[string]string{"message": "献立を更新しました", "date": dateNorm}, http.StatusOK)
	}
}

// GET /api/calc?date=2026-04-01
func handleCalc(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		date := r.URL.Query().Get("date")
		if date == "" {
			writeError(w, "date を指定してください", http.StatusBadRequest)
			return
		}

		var stapleID, mainID, sideID, soupID, dessertID int
		err := db.QueryRow(q(`
			SELECT staple_id, main_id, side_id, soup_id, dessert_id FROM menus WHERE date = ? AND facility_id = ?
		`), date, DefaultFacilityID).Scan(&stapleID, &mainID, &sideID, &soupID, &dessertID)
		if err == sql.ErrNoRows {
			writeError(w, date+" の献立が登録されていません", http.StatusNotFound)
			return
		}
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		nut, err := calcMenuNutrition(db, stapleID, mainID, sideID, soupID, dessertID)
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, map[string]interface{}{
			"date":         date,
			"energy":       nut.Energy,
			"protein":      nut.Protein,
			"fat":          nut.Fat,
			"carbohydrate": nut.Carbohydrate,
			"salt":         nut.Salt,
			"targets": map[string]float64{
				"energy":       targetEnergy,
				"protein":      targetProtein,
				"fat":          targetFat,
				"carbohydrate": targetCarbohydrate,
				"salt":         targetSalt,
			},
		}, http.StatusOK)
	}
}

// GET /api/order?start=2026-04-01&end=2026-04-30&people=120
func handleOrder(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")
		peopleStr := r.URL.Query().Get("people")
		if start == "" || end == "" || peopleStr == "" {
			writeError(w, "start, end, people を指定してください", http.StatusBadRequest)
			return
		}
		people, err := strconv.Atoi(peopleStr)
		if err != nil || people < 1 {
			writeError(w, "people は正の整数で指定してください", http.StatusBadRequest)
			return
		}

		excludeCondiments := r.URL.Query().Get("exclude_condiments") == "1"
		totalByIng, err := aggregateOrder(db, start, end, people, excludeCondiments)
		if err != nil {
			writeError(w, err.Error(), http.StatusNotFound)
			return
		}

		ingNames := make(map[int]string)
		for ingID := range totalByIng {
			var name string
			db.QueryRow(q("SELECT name FROM ingredients WHERE id = ?"), ingID).Scan(&name)
			ingNames[ingID] = name
		}

		ingIDs := make([]int, 0, len(totalByIng))
		for ingID := range totalByIng {
			ingIDs = append(ingIDs, ingID)
		}
		sort.Ints(ingIDs)

		var items []map[string]interface{}
		for _, ingID := range ingIDs {
			total := totalByIng[ingID]
			name := ingNames[ingID]
			if name == "" {
				name = "不明"
			}
			items = append(items, map[string]interface{}{
				"ingredient_id": ingID,
				"name":          name,
				"total_g":       total,
			})
		}
		writeJSON(w, map[string]interface{}{
			"start":  start,
			"end":    end,
			"people": people,
			"items":  items,
		}, http.StatusOK)
	}
}

// GET /api/order/bulk?start=2026-04-01&end=2026-04-30&people=120
func handleBulkOrder(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")
		peopleStr := r.URL.Query().Get("people")
		if start == "" || end == "" || peopleStr == "" {
			writeError(w, "start, end, people を指定してください", http.StatusBadRequest)
			return
		}
		people, err := strconv.Atoi(peopleStr)
		if err != nil || people < 1 {
			writeError(w, "people は正の整数で指定してください", http.StatusBadRequest)
			return
		}

		totalByIng, err := aggregateOrder(db, start, end, people, false)
		if err != nil {
			writeError(w, err.Error(), http.StatusNotFound)
			return
		}

		rows, err := db.Query(q(`
			SELECT b.ingredient_id, i.name, b.order_unit_g, b.order_unit_name, b.bulk_category
			FROM bulk_purchase_guide b
			JOIN ingredients i ON b.ingredient_id = i.id
			ORDER BY b.bulk_category, b.ingredient_id
		`))
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var items []map[string]interface{}
		for rows.Next() {
			var ingID int
			var name string
			var orderUnitG float64
			var orderUnitName, bulkCategory string
			if err := rows.Scan(&ingID, &name, &orderUnitG, &orderUnitName, &bulkCategory); err != nil {
				continue
			}
			totalG, ok := totalByIng[ingID]
			if !ok || totalG <= 0 {
				continue
			}
			orderQty := int(math.Ceil(totalG / orderUnitG))
			items = append(items, map[string]interface{}{
				"ingredient_id":   ingID,
				"name":           name,
				"total_g":        totalG,
				"order_unit_g":   orderUnitG,
				"order_unit_name": orderUnitName,
				"bulk_category":  bulkCategory,
				"order_qty":      orderQty,
			})
		}

		writeJSON(w, map[string]interface{}{
			"start":  start,
			"end":    end,
			"people": people,
			"items":  items,
		}, http.StatusOK)
	}
}

// GET /api/export?month=2026-04
func handleExport(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		month := r.URL.Query().Get("month")
		if month == "" {
			writeError(w, "month を指定してください (YYYY-MM)", http.StatusBadRequest)
			return
		}

		parts := strings.Split(month, "-")
		if len(parts) != 2 {
			writeError(w, "month は YYYY-MM 形式で指定してください", http.StatusBadRequest)
			return
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
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
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
			wd := t.Weekday()
			idx := 0
			if wd == 0 {
				idx = 6
			} else {
				idx = int(wd) - 1
			}
			weekday := weekdayJA[idx]
			records = append(records, []interface{}{date, weekday, staple, mainD, side, soup, dessert, nut.Energy})
		}
		if len(records) == 0 {
			writeError(w, month+" に献立が登録されていません", http.StatusNotFound)
			return
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

		var buf bytes.Buffer
		if err := f.Write(&buf); err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		filename := fmt.Sprintf("献立表_%d年%02d月.xlsx", year, m)
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
	}
}

func runServer(db *sql.DB, port string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/dishes", handleDishes(db))
	mux.HandleFunc("/api/menus", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreateMenu(db)(w, r)
		case http.MethodPut:
			handleUpdateMenu(db)(w, r)
		default:
			handleMenus(db)(w, r)
		}
	})
	mux.HandleFunc("/api/calc", handleCalc(db))
	mux.HandleFunc("/api/order", handleOrder(db))
	mux.HandleFunc("/api/order/bulk", handleBulkOrder(db))
	mux.HandleFunc("/api/export", handleExport(db))
	mux.HandleFunc("/api/auth/login", handleLogin(db))
	mux.HandleFunc("/api/auth/register", handleRegister(db))

	handler := corsMiddleware(mux)
	addr := ":" + port
	fmt.Printf("API サーバ起動: http://localhost%s\n", addr)
	return http.ListenAndServe(addr, handler)
}
