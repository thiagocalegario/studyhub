package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/thiagocalegario/studyhub/internal/database"
	"github.com/thiagocalegario/studyhub/internal/models"
	"github.com/thiagocalegario/studyhub/internal/session"
)

func CommunityPage(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.Get(r)

	// Extrai o discipline_id da URL: /disciplines/{id}/community
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/disciplines/"), "/")
	disciplineID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
		return
	}

	// Busca dados da disciplina — qualquer usuário logado pode ver o fórum
	// mesmo sem ter adicionado a disciplina
	var d models.Discipline
	err = database.DB.QueryRow(
		"SELECT id, code, name, semester, workload FROM disciplines WHERE id = $1",
		disciplineID,
	).Scan(&d.ID, &d.Code, &d.Name, &d.Semester, &d.Workload)
	if err != nil {
		http.Error(w, "Disciplina não encontrada", http.StatusNotFound)
		return
	}

	// Busca todos os posts públicos da disciplina, mais antigos primeiro
	rows, err := database.DB.Query(`
        SELECT cp.id, cp.user_id, cp.discipline_id, cp.content, cp.created_at, u.name
        FROM community_posts cp
        INNER JOIN users u ON u.id = cp.user_id
        WHERE cp.discipline_id = $1
        ORDER BY cp.created_at ASC
    `, disciplineID)
	if err != nil {
		http.Error(w, "Erro ao buscar posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.CommunityPost
	for rows.Next() {
		var p models.CommunityPost
		if err := rows.Scan(&p.ID, &p.UserID, &p.DisciplineID, &p.Content, &p.CreatedAt, &p.UserName); err != nil {
			continue
		}
		p.IsOwner = p.UserID == sess.UserID
		posts = append(posts, p)
	}

	errorMsg := ""
	if r.URL.Query().Get("error") == "vazio" {
		errorMsg = "Escreva algo antes de publicar."
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/community.html",
	))
	tmpl.ExecuteTemplate(w, "layout", map[string]interface{}{
		"Title":      "Fórum — " + d.Name,
		"User":       sess,
		"Discipline": d,
		"Posts":      posts,
		"Error":      errorMsg,
	})
}

func AddCommunityPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
		return
	}

	sess, _ := session.Get(r)

	disciplineIDStr := r.FormValue("discipline_id")
	disciplineID, err := strconv.Atoi(disciplineIDStr)
	if err != nil {
		http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
		return
	}

	content := strings.TrimSpace(r.FormValue("content"))
	if content == "" {
		http.Redirect(w, r, "/disciplines/"+disciplineIDStr+"/community?error=vazio", http.StatusSeeOther)
		return
	}

	database.DB.Exec(
		"INSERT INTO community_posts (user_id, discipline_id, content) VALUES ($1, $2, $3)",
		sess.UserID, disciplineID, content,
	)

	http.Redirect(w, r, "/disciplines/"+disciplineIDStr+"/community", http.StatusSeeOther)
}

func DeleteCommunityPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
		return
	}

	sess, _ := session.Get(r)

	postIDStr := r.FormValue("post_id")
	disciplineIDStr := r.FormValue("discipline_id")

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
		return
	}

	// Só deleta se o post pertencer ao próprio usuário
	database.DB.Exec(
		"DELETE FROM community_posts WHERE id = $1 AND user_id = $2",
		postID, sess.UserID,
	)

	http.Redirect(w, r, "/disciplines/"+disciplineIDStr+"/community", http.StatusSeeOther)
}
