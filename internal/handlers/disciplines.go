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

var validTags = []string{"Geral", "Resumo", "Dúvida", "Exercício", "Prova"}
var validStatuses = []string{"Em andamento", "Concluído", "Revisar"}

var tmplFuncs = template.FuncMap{
	"tagClass": func(tag string) string {
		switch tag {
		case "Resumo":
			return "resumo"
		case "Dúvida":
			return "duvida"
		case "Exercício":
			return "exercicio"
		case "Prova":
			return "prova"
		default:
			return "geral"
		}
	},
	"statusClass": func(status string) string {
		switch status {
		case "Concluído":
			return "concluido"
		case "Revisar":
			return "revisar"
		default:
			return "andamento"
		}
	},
}

func CatalogPage(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.Get(r)

	rows, err := database.DB.Query(`
        SELECT id, name, university FROM courses ORDER BY name
    `)
	if err != nil {
		http.Error(w, "Erro ao buscar cursos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var courses []models.Course
	for rows.Next() {
		var c models.Course
		if err := rows.Scan(&c.ID, &c.Name, &c.University); err != nil {
			continue
		}
		courses = append(courses, c)
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/catalog.html",
	))
	tmpl.ExecuteTemplate(w, "layout", map[string]interface{}{
		"Title":   "Catálogo de Cursos",
		"User":    sess,
		"Courses": courses,
	})
}

func CatalogCoursePage(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.Get(r)

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/catalog/course/"), "/")
	courseID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Redirect(w, r, "/catalog", http.StatusSeeOther)
		return
	}

	var course models.Course
	err = database.DB.QueryRow(
		"SELECT id, name, university FROM courses WHERE id = $1", courseID,
	).Scan(&course.ID, &course.Name, &course.University)
	if err != nil {
		http.Redirect(w, r, "/catalog", http.StatusSeeOther)
		return
	}

	rows, err := database.DB.Query(`
        SELECT d.id, d.code, d.name, d.semester, d.workload,
               EXISTS (
                   SELECT 1 FROM user_disciplines ud
                   WHERE ud.user_id = $1 AND ud.discipline_id = d.id
               ) AS added
        FROM disciplines d
        WHERE d.course_id = $2
        ORDER BY d.semester, d.name
    `, sess.UserID, courseID)
	if err != nil {
		http.Error(w, "Erro ao buscar disciplinas", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	semesterMap := make(map[int][]models.Discipline)
	semesterOrder := []int{}

	for rows.Next() {
		var d models.Discipline
		if err := rows.Scan(&d.ID, &d.Code, &d.Name, &d.Semester, &d.Workload, &d.Added); err != nil {
			continue
		}
		if _, exists := semesterMap[d.Semester]; !exists {
			semesterOrder = append(semesterOrder, d.Semester)
		}
		semesterMap[d.Semester] = append(semesterMap[d.Semester], d)
	}

	type SemesterGroup struct {
		Number      int
		Disciplines []models.Discipline
	}

	var groups []SemesterGroup
	for _, s := range semesterOrder {
		groups = append(groups, SemesterGroup{
			Number:      s,
			Disciplines: semesterMap[s],
		})
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/catalog_course.html",
	))
	tmpl.ExecuteTemplate(w, "layout", map[string]interface{}{
		"Title":     course.Name,
		"User":      sess,
		"Course":    course,
		"Semesters": groups,
	})
}

func AddDiscipline(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.Get(r)

	disciplineIDStr := r.FormValue("discipline_id")
	courseIDStr := r.FormValue("course_id")

	disciplineID, err := strconv.Atoi(disciplineIDStr)
	if err != nil {
		http.Redirect(w, r, "/catalog", http.StatusSeeOther)
		return
	}

	database.DB.Exec(
		"INSERT INTO user_disciplines (user_id, discipline_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		sess.UserID, disciplineID,
	)

	if courseIDStr != "" {
		http.Redirect(w, r, "/catalog/course/"+courseIDStr, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/catalog", http.StatusSeeOther)
}

func RemoveDiscipline(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.Get(r)

	disciplineIDStr := r.FormValue("discipline_id")
	redirectTo := r.FormValue("redirect_to")

	disciplineID, err := strconv.Atoi(disciplineIDStr)
	if err != nil {
		http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
		return
	}

	database.DB.Exec(
		"DELETE FROM user_disciplines WHERE user_id = $1 AND discipline_id = $2",
		sess.UserID, disciplineID,
	)

	if redirectTo != "" {
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
}

func MyDisciplinesPage(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.Get(r)

	rows, err := database.DB.Query(`
        SELECT d.id, d.code, d.name, d.semester, d.workload
        FROM disciplines d
        INNER JOIN user_disciplines ud ON ud.discipline_id = d.id
        WHERE ud.user_id = $1
        ORDER BY d.semester, d.name
    `, sess.UserID)
	if err != nil {
		http.Error(w, "Erro ao buscar disciplinas", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	topicCounts := make(map[int]int)
	tcRows, err := database.DB.Query(`
        SELECT discipline_id, COUNT(*) FROM topics WHERE user_id = $1 GROUP BY discipline_id
    `, sess.UserID)
	if err == nil {
		defer tcRows.Close()
		for tcRows.Next() {
			var disciplineID, count int
			if err := tcRows.Scan(&disciplineID, &count); err == nil {
				topicCounts[disciplineID] = count
			}
		}
	}

	semesterMap := make(map[int][]map[string]interface{})
	semesterOrder := []int{}

	for rows.Next() {
		var d models.Discipline
		if err := rows.Scan(&d.ID, &d.Code, &d.Name, &d.Semester, &d.Workload); err != nil {
			continue
		}
		if _, exists := semesterMap[d.Semester]; !exists {
			semesterOrder = append(semesterOrder, d.Semester)
		}
		semesterMap[d.Semester] = append(semesterMap[d.Semester], map[string]interface{}{
			"ID":         d.ID,
			"Code":       d.Code,
			"Name":       d.Name,
			"Semester":   d.Semester,
			"Workload":   d.Workload,
			"TopicCount": topicCounts[d.ID],
		})
	}

	type SemesterGroup struct {
		Number      int
		Disciplines []map[string]interface{}
	}

	var groups []SemesterGroup
	for _, s := range semesterOrder {
		groups = append(groups, SemesterGroup{
			Number:      s,
			Disciplines: semesterMap[s],
		})
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/my_disciplines.html",
	))
	tmpl.ExecuteTemplate(w, "layout", map[string]interface{}{
		"Title":     "Minhas Disciplinas",
		"User":      sess,
		"Semesters": groups,
	})
}

func DisciplineDetailPage(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.Get(r)

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/disciplines/"), "/")
	disciplineID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
		return
	}

	var count int
	err = database.DB.QueryRow(
		"SELECT COUNT(*) FROM user_disciplines WHERE user_id = $1 AND discipline_id = $2",
		sess.UserID, disciplineID,
	).Scan(&count)
	if err != nil || count == 0 {
		http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
		return
	}

	var d models.Discipline
	err = database.DB.QueryRow(
		"SELECT id, code, name, semester, workload FROM disciplines WHERE id = $1",
		disciplineID,
	).Scan(&d.ID, &d.Code, &d.Name, &d.Semester, &d.Workload)
	if err != nil {
		http.Error(w, "Disciplina não encontrada", http.StatusNotFound)
		return
	}

	filterTag := r.URL.Query().Get("tag")
	filterStatus := r.URL.Query().Get("status")

	query := `
        SELECT t.id, t.user_id, t.discipline_id, t.title, t.content, t.tag, t.status, t.created_at, u.name
        FROM topics t
        INNER JOIN users u ON u.id = t.user_id
        WHERE t.discipline_id = $1 AND t.user_id = $2
    `
	args := []interface{}{disciplineID, sess.UserID}

	if filterTag != "" {
		args = append(args, filterTag)
		query += " AND t.tag = $" + strconv.Itoa(len(args))
	}
	if filterStatus != "" {
		args = append(args, filterStatus)
		query += " AND t.status = $" + strconv.Itoa(len(args))
	}

	query += " ORDER BY t.created_at DESC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Erro ao buscar tópicos: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var topics []models.Topic
	for rows.Next() {
		var t models.Topic
		if err := rows.Scan(&t.ID, &t.UserID, &t.DisciplineID, &t.Title, &t.Content, &t.Tag, &t.Status, &t.CreatedAt, &t.UserName); err != nil {
			continue
		}
		topics = append(topics, t)
	}

	errorMsg := ""
	if r.URL.Query().Get("error") == "campos_obrigatorios" {
		errorMsg = "Preencha todos os campos antes de salvar."
	}

	// Usa template.New com nome neutro para evitar conflito com o define "layout"
	tmpl := template.Must(
		template.New("base").Funcs(tmplFuncs).ParseFiles(
			"web/templates/layout.html",
			"web/templates/discipline_detail.html",
		),
	)
	tmpl.ExecuteTemplate(w, "layout", map[string]interface{}{
		"Title":        d.Name,
		"User":         sess,
		"Discipline":   d,
		"Topics":       topics,
		"Tags":         validTags,
		"Statuses":     validStatuses,
		"FilterTag":    filterTag,
		"FilterStatus": filterStatus,
		"Error":        errorMsg,
	})
}

func EditTopicPage(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.Get(r)

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/disciplines/topic/edit/"), "/")
	topicID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
		return
	}

	var t models.Topic
	err = database.DB.QueryRow(`
        SELECT id, user_id, discipline_id, title, content, tag, status, created_at
        FROM topics WHERE id = $1 AND user_id = $2
    `, topicID, sess.UserID).Scan(
		&t.ID, &t.UserID, &t.DisciplineID, &t.Title, &t.Content, &t.Tag, &t.Status, &t.CreatedAt,
	)
	if err != nil {
		// Loga o erro em vez de redirecionar silenciosamente
		http.Error(w, "Tópico não encontrado: "+err.Error(), http.StatusNotFound)
		return
	}

	var d models.Discipline
	database.DB.QueryRow(
		"SELECT id, code, name, semester, workload FROM disciplines WHERE id = $1",
		t.DisciplineID,
	).Scan(&d.ID, &d.Code, &d.Name, &d.Semester, &d.Workload)

	errorMsg := ""
	if r.URL.Query().Get("error") == "campos_obrigatorios" {
		errorMsg = "Preencha todos os campos antes de salvar."
	}

	// Usa template.New com nome neutro para evitar conflito com o define "layout"
	tmpl := template.Must(
		template.New("base").Funcs(tmplFuncs).ParseFiles(
			"web/templates/layout.html",
			"web/templates/topic_edit.html",
		),
	)
	tmpl.ExecuteTemplate(w, "layout", map[string]interface{}{
		"Title":      "Editar Tópico",
		"User":       sess,
		"Topic":      t,
		"Discipline": d,
		"Tags":       validTags,
		"Statuses":   validStatuses,
		"Error":      errorMsg,
	})
}

func EditTopicPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
		return
	}

	sess, _ := session.Get(r)

	topicIDStr := r.FormValue("topic_id")
	disciplineIDStr := r.FormValue("discipline_id")
	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))
	tag := r.FormValue("tag")
	status := r.FormValue("status")

	topicID, err := strconv.Atoi(topicIDStr)
	if err != nil {
		http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
		return
	}

	if title == "" || content == "" {
		http.Redirect(w, r, "/disciplines/topic/edit/"+topicIDStr+"?error=campos_obrigatorios", http.StatusSeeOther)
		return
	}

	if !isValidOption(tag, validTags) {
		tag = "Geral"
	}
	if !isValidOption(status, validStatuses) {
		status = "Em andamento"
	}

	database.DB.Exec(`
        UPDATE topics SET title = $1, content = $2, tag = $3, status = $4
        WHERE id = $5 AND user_id = $6
    `, title, content, tag, status, topicID, sess.UserID)

	http.Redirect(w, r, "/disciplines/"+disciplineIDStr, http.StatusSeeOther)
}

func AddTopic(w http.ResponseWriter, r *http.Request) {
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

	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))
	tag := r.FormValue("tag")
	status := r.FormValue("status")

	if title == "" || content == "" {
		http.Redirect(w, r, "/disciplines/"+disciplineIDStr+"?error=campos_obrigatorios", http.StatusSeeOther)
		return
	}

	if !isValidOption(tag, validTags) {
		tag = "Geral"
	}
	if !isValidOption(status, validStatuses) {
		status = "Em andamento"
	}

	var count int
	err = database.DB.QueryRow(
		"SELECT COUNT(*) FROM user_disciplines WHERE user_id = $1 AND discipline_id = $2",
		sess.UserID, disciplineID,
	).Scan(&count)
	if err != nil || count == 0 {
		http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
		return
	}

	database.DB.Exec(
		"INSERT INTO topics (user_id, discipline_id, title, content, tag, status) VALUES ($1, $2, $3, $4, $5, $6)",
		sess.UserID, disciplineID, title, content, tag, status,
	)

	http.Redirect(w, r, "/disciplines/"+disciplineIDStr, http.StatusSeeOther)
}

func DeleteTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
		return
	}

	sess, _ := session.Get(r)

	topicIDStr := r.FormValue("topic_id")
	disciplineIDStr := r.FormValue("discipline_id")

	topicID, err := strconv.Atoi(topicIDStr)
	if err != nil {
		http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
		return
	}

	database.DB.Exec(
		"DELETE FROM topics WHERE id = $1 AND user_id = $2",
		topicID, sess.UserID,
	)

	http.Redirect(w, r, "/disciplines/"+disciplineIDStr, http.StatusSeeOther)
}

func isValidOption(value string, options []string) bool {
	for _, o := range options {
		if o == value {
			return true
		}
	}
	return false
}
