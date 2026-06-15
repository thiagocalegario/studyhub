package session

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

const cookieName = "studyhub_session"

type SessionData struct {
	UserID int
	Name   string
	Email  string
}

func Set(w http.ResponseWriter, data SessionData) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}

	encoded := base64.URLEncoding.EncodeToString(payload)

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
	})
}

func Get(r *http.Request) (*SessionData, bool) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, false
	}

	decoded, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, false
	}

	var data SessionData
	if err := json.Unmarshal(decoded, &data); err != nil {
		return nil, false
	}

	if data.UserID == 0 {
		return nil, false
	}

	return &data, true
}

func Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    cookieName,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	})
}

func init() {
	_ = os.Getenv("SESSION_SECRET")
}
