// Phase4 Task3: JWT 認証
package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const jwtExpireHours = 24

type claims struct {
	UserID     int    `json:"user_id"`
	Email      string `json:"email"`
	FacilityID int    `json:"facility_id"`
	Role       string `json:"role"`
	jwt.RegisteredClaims
}

func jwtSecret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "kyuushoku-dev-secret-change-in-production"
	}
	return []byte(s)
}

func generateToken(userID int, email string, facilityID int, role string) (string, error) {
	claims := claims{
		UserID:     userID,
		Email:      email,
		FacilityID: facilityID,
		Role:       role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtExpireHours * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret())
}

func parseToken(tokenString string) (*claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	if c, ok := token.Claims.(*claims); ok && token.Valid {
		return c, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

// POST /api/auth/login
func handleLogin(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if body.Email == "" || body.Password == "" {
			writeError(w, "email と password は必須です", http.StatusBadRequest)
			return
		}

		var userID int
		var pwHash string
		var facilityID sql.NullInt64
		var role string
		err := db.QueryRow(q(`
			SELECT id, password_hash, facility_id, role FROM users WHERE email = ?
		`), body.Email).Scan(&userID, &pwHash, &facilityID, &role)
		if err == sql.ErrNoRows {
			writeError(w, "メールアドレスまたはパスワードが正しくありません", http.StatusUnauthorized)
			return
		}
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(pwHash), []byte(body.Password)); err != nil {
			writeError(w, "メールアドレスまたはパスワードが正しくありません", http.StatusUnauthorized)
			return
		}

		fid := 1
		if facilityID.Valid {
			fid = int(facilityID.Int64)
		}
		token, err := generateToken(userID, body.Email, fid, role)
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, map[string]interface{}{
			"token": token,
			"user": map[string]interface{}{
				"id":          userID,
				"email":      body.Email,
				"facility_id": fid,
				"role":       role,
			},
		}, http.StatusOK)
	}
}

// POST /api/auth/register（初回ユーザー作成のみ）
func handleRegister(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if body.Email == "" || body.Password == "" {
			writeError(w, "email と password は必須です", http.StatusBadRequest)
			return
		}
		if len(body.Password) < 6 {
			writeError(w, "パスワードは6文字以上にしてください", http.StatusBadRequest)
			return
		}

		var count int
		db.QueryRow(q("SELECT COUNT(*) FROM users")).Scan(&count)
		if count > 0 {
			writeError(w, "既にユーザーが登録されています。ログインしてください。", http.StatusConflict)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(q(`
			INSERT INTO users (email, password_hash, facility_id, role) VALUES (?, ?, 1, 'admin')
		`), body.Email, string(hash))
		if err != nil {
			writeError(w, dbErrorToJapanese(err), http.StatusInternalServerError)
			return
		}

		writeJSON(w, map[string]string{"message": "ユーザーを登録しました。ログインしてください。"}, http.StatusCreated)
	}
}

// AuthMiddleware は Authorization ヘッダーから JWT を検証（オプション）
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			token := strings.TrimPrefix(auth, "Bearer ")
			if _, err := parseToken(token); err == nil {
				// 認証成功（将来は r.Context() にユーザー情報をセット）
			}
		}
		next.ServeHTTP(w, r)
	})
}
