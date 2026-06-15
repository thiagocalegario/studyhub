package handlers

import (
	"html/template"
	"net/http"

	"github.com/thiagocalegario/studyhub/internal/database"
	"github.com/thiagocalegario/studyhub/internal/session"
	"golang.org/x/crypto/bcrypt"
)

func RegisterPage(w http.ResponseWriter, r *http.Request) {
	errorMsg := ""
	successMsg := ""

	switch r.URL.Query().Get("error") {
	case "campos_obrigatorios":
		errorMsg = "Preencha todos os campos obrigatórios."
	case "email_em_uso":
		errorMsg = "Este e-mail já está cadastrado. Tente fazer login."
	case "erro_interno":
		errorMsg = "Ocorreu um erro interno. Tente novamente."
	}

	if r.URL.Query().Get("success") == "cadastrado" {
		successMsg = "Conta criada com sucesso! Faça login para continuar."
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/register.html",
	))
	tmpl.ExecuteTemplate(w, "layout", map[string]interface{}{
		"Title":   "Cadastro",
		"Error":   errorMsg,
		"Success": successMsg,
	})
}

func RegisterPost(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if name == "" || email == "" || password == "" {
		http.Redirect(w, r, "/register?error=campos_obrigatorios", http.StatusSeeOther)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Redirect(w, r, "/register?error=erro_interno", http.StatusSeeOther)
		return
	}

	var id int
	err = database.DB.QueryRow(
		"INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id",
		name, email, string(hash),
	).Scan(&id)

	if err != nil {
		http.Redirect(w, r, "/register?error=email_em_uso", http.StatusSeeOther)
		return
	}

	session.Set(w, session.SessionData{
		UserID: id,
		Name:   name,
		Email:  email,
	})

	http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	errorMsg := ""
	successMsg := ""

	switch r.URL.Query().Get("error") {
	case "credenciais_invalidas":
		errorMsg = "E-mail ou senha incorretos. Tente novamente."
	case "campos_obrigatorios":
		errorMsg = "Preencha e-mail e senha antes de continuar."
	}

	if r.URL.Query().Get("success") == "logout" {
		successMsg = "Você saiu com sucesso."
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/login.html",
	))
	tmpl.ExecuteTemplate(w, "layout", map[string]interface{}{
		"Title":   "Login",
		"Error":   errorMsg,
		"Success": successMsg,
	})
}

func LoginPost(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		http.Redirect(w, r, "/login?error=campos_obrigatorios", http.StatusSeeOther)
		return
	}

	var id int
	var name, hash string

	err := database.DB.QueryRow(
		"SELECT id, name, password_hash FROM users WHERE email = $1",
		email,
	).Scan(&id, &name, &hash)

	if err != nil {
		http.Redirect(w, r, "/login?error=credenciais_invalidas", http.StatusSeeOther)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		http.Redirect(w, r, "/login?error=credenciais_invalidas", http.StatusSeeOther)
		return
	}

	session.Set(w, session.SessionData{
		UserID: id,
		Name:   name,
		Email:  email,
	})

	http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session.Clear(w)
	http.Redirect(w, r, "/login?success=logout", http.StatusSeeOther)
}
