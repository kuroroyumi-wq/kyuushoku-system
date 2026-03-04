// db.go: SQLite / PostgreSQL 両対応
// DATABASE_URL が postgres:// で始まる場合、PostgreSQL を使用

package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

var usePostgres bool

// DefaultFacilityID は未指定時のデフォルト施設（Phase4）
const DefaultFacilityID = 1

// q は PostgreSQL 使用時に ? を $1, $2, ... に変換する
func q(s string) string {
	if !usePostgres {
		return s
	}
	var b strings.Builder
	n := 1
	for _, c := range s {
		if c == '?' {
			fmt.Fprintf(&b, "$%d", n)
			n++
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}

func openDB() (*sql.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" && (strings.HasPrefix(dbURL, "postgres://") || strings.HasPrefix(dbURL, "postgresql://")) {
		usePostgres = true
		db, err := sql.Open("pgx", dbURL)
		if err != nil {
			return nil, fmt.Errorf("PostgreSQL 接続エラー: %w", err)
		}
		if err := db.Ping(); err != nil {
			return nil, fmt.Errorf("PostgreSQL 接続確認失敗: %w", err)
		}
		return db, nil
	}

	usePostgres = false
	dbPath := findDBPath()
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("データベースが見つかりません。先に tools/migrate_csv_to_sqlite.py を実行してください: %s", dbPath)
	}
	return sql.Open("sqlite", dbPath)
}
