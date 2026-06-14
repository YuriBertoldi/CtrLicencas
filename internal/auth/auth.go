package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"delphilic/internal/models"
)

const sessionCookie = "dlicsession"
const sessionDuration = 24 * time.Hour

func NewToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func CreateSession(db *sql.DB, userID int) (string, error) {
	token := NewToken()
	expira := time.Now().Add(sessionDuration)
	_, err := db.Exec(
		`INSERT INTO sessions (token, user_id, expira_em) VALUES ($1, $2, $3)`,
		token, userID, expira,
	)
	return token, err
}

func DeleteSession(db *sql.DB, token string) {
	db.Exec(`DELETE FROM sessions WHERE token = $1`, token)
}

func GetUserFromSession(db *sql.DB, r *http.Request) *models.Usuario {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return nil
	}
	var u models.Usuario
	err = db.QueryRow(`
		SELECT u.id, u.nome, u.email, u.admin, u.ativo
		FROM sessions s
		JOIN usuarios u ON u.id = s.user_id
		WHERE s.token = $1 AND s.expira_em > NOW() AND u.ativo = true
	`, c.Value).Scan(&u.ID, &u.Nome, &u.Email, &u.Admin, &u.Ativo)
	if err != nil {
		return nil
	}
	return &u
}

func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionDuration.Seconds()),
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookie,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

// Protected exige login. Redireciona para /login se não autenticado.
func Protected(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := GetUserFromSession(db, r)
		if u == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// AdminOnly exige login + admin.
func AdminOnly(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := GetUserFromSession(db, r)
		if u == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if !u.Admin {
			http.Error(w, "Acesso negado", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
