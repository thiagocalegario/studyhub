package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/thiagocalegario/studyhub/internal/database"
	"github.com/thiagocalegario/studyhub/internal/models"
	"github.com/thiagocalegario/studyhub/internal/session"
	"github.com/thiagocalegario/studyhub/internal/ws"
)

type cardWithReplies struct {
	models.CommunityPost
	Replies []models.ForumReply
}

type wsMessage struct {
	Type    string `json:"type"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content"`
	CardID  int    `json:"card_id,omitempty"`
}

var communityTmpl = template.Must(template.ParseFiles(
	"web/templates/layout.html",
	"web/templates/community.html",
))

func CommunityPage(w http.ResponseWriter, r *http.Request) {
	sess, _ := session.Get(r)

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/disciplines/"), "/")
	disciplineID, err := strconv.Atoi(parts[0])
	if err != nil {
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

	cardRows, err := database.DB.Query(`
		SELECT cp.id, cp.user_id, cp.discipline_id, cp.title, cp.content, cp.created_at, u.name
		FROM community_posts cp
		INNER JOIN users u ON u.id = cp.user_id
		WHERE cp.discipline_id = $1
		ORDER BY cp.created_at ASC
	`, disciplineID)
	if err != nil {
		http.Error(w, "Erro ao buscar cards", http.StatusInternalServerError)
		return
	}
	defer cardRows.Close()

	type cardRow struct {
		ID           int
		UserID       int
		DisciplineID int
		Title        string
		Content      string
		CreatedAt    time.Time
		UserName     string
	}

	var scanned []cardRow
	for cardRows.Next() {
		var c cardRow
		if err := cardRows.Scan(&c.ID, &c.UserID, &c.DisciplineID, &c.Title, &c.Content, &c.CreatedAt, &c.UserName); err != nil {
			continue
		}
		scanned = append(scanned, c)
	}

	replyRows, err := database.DB.Query(`
		SELECT fr.id, fr.card_id, fr.user_id, fr.content, fr.created_at, u.name
		FROM forum_replies fr
		INNER JOIN users u ON u.id = fr.user_id
		INNER JOIN community_posts cp ON cp.id = fr.card_id
		WHERE cp.discipline_id = $1
		ORDER BY fr.created_at ASC
	`, disciplineID)
	if err == nil {
		defer replyRows.Close()
	}

	replyMap := make(map[int][]models.ForumReply)
	if err == nil {
		for replyRows.Next() {
			var r models.ForumReply
			if err := replyRows.Scan(&r.ID, &r.CardID, &r.UserID, &r.Content, &r.CreatedAt, &r.UserName); err != nil {
				continue
			}
			r.IsOwner = r.UserID == sess.UserID
			replyMap[r.CardID] = append(replyMap[r.CardID], r)
		}
	}

	var cards []cardWithReplies
	for _, c := range scanned {
		card := cardWithReplies{
			CommunityPost: models.CommunityPost{
				ID:           c.ID,
				UserID:       c.UserID,
				DisciplineID: c.DisciplineID,
				Title:        c.Title,
				Content:      c.Content,
				CreatedAt:    c.CreatedAt,
				UserName:     c.UserName,
				IsOwner:      c.UserID == sess.UserID,
			},
			Replies: replyMap[c.ID],
		}
		if card.Replies == nil {
			card.Replies = []models.ForumReply{}
		}
		cards = append(cards, card)
	}

	errorMsg := ""
	if r.URL.Query().Get("error") == "vazio" {
		errorMsg = "Escreva algo antes de publicar."
	}

	communityTmpl.ExecuteTemplate(w, "layout", map[string]interface{}{
		"Title":      "Fórum — " + d.Name,
		"User":       sess,
		"Discipline": d,
		"Cards":      cards,
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

	title := strings.TrimSpace(r.FormValue("title"))
	content := strings.TrimSpace(r.FormValue("content"))
	if title == "" || content == "" {
		http.Redirect(w, r, "/disciplines/"+disciplineIDStr+"/community?error=vazio", http.StatusSeeOther)
		return
	}

	database.DB.Exec(
		"INSERT INTO community_posts (user_id, discipline_id, title, content) VALUES ($1, $2, $3, $4)",
		sess.UserID, disciplineID, title, content,
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

	database.DB.Exec(
		"DELETE FROM community_posts WHERE id = $1 AND user_id = $2",
		postID, sess.UserID,
	)

	disciplineID, _ := strconv.Atoi(disciplineIDStr)
	hub := ws.GetHub(disciplineID)
	hub.BroadcastJSON(map[string]interface{}{
		"type":    "delete_card",
		"payload": map[string]interface{}{"id": postID},
	})

	http.Redirect(w, r, "/disciplines/"+disciplineIDStr+"/community", http.StatusSeeOther)
}

func AddForumReply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
		return
	}

	sess, _ := session.Get(r)

	cardIDStr := r.FormValue("card_id")
	disciplineIDStr := r.FormValue("discipline_id")
	content := strings.TrimSpace(r.FormValue("content"))

	cardID, err := strconv.Atoi(cardIDStr)
	if err != nil || content == "" {
		http.Redirect(w, r, "/disciplines/"+disciplineIDStr+"/community", http.StatusSeeOther)
		return
	}

	database.DB.Exec(
		"INSERT INTO forum_replies (card_id, user_id, content) VALUES ($1, $2, $3)",
		cardID, sess.UserID, content,
	)

	http.Redirect(w, r, "/disciplines/"+disciplineIDStr+"/community", http.StatusSeeOther)
}

func ServeForumWS(w http.ResponseWriter, r *http.Request) {
	sess, ok := session.Get(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/ws/forum/"), "/")
	disciplineID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid discipline", http.StatusBadRequest)
		return
	}

	hub := ws.GetHub(disciplineID)

	conn, err := ws.Upgrade(w, r)
	if err != nil {
		return
	}

	client := &ws.Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		UserID:   sess.UserID,
		UserName: sess.Name,
		OnMessage: func(c *ws.Client, data []byte) {
			var msg wsMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				return
			}

			switch msg.Type {
			case "new_card":
				title := strings.TrimSpace(msg.Title)
				content := strings.TrimSpace(msg.Content)
				if title == "" || content == "" {
					return
				}

				var id int
				err := database.DB.QueryRow(
					"INSERT INTO community_posts (user_id, discipline_id, title, content) VALUES ($1, $2, $3, $4) RETURNING id",
					c.UserID, disciplineID, title, content,
				).Scan(&id)
				if err != nil {
					return
				}

				var card models.CommunityPost
				database.DB.QueryRow(`
					SELECT cp.id, cp.user_id, cp.discipline_id, cp.title, cp.content, cp.created_at, u.name
					FROM community_posts cp
					INNER JOIN users u ON u.id = cp.user_id
					WHERE cp.id = $1
				`, id).Scan(&card.ID, &card.UserID, &card.DisciplineID, &card.Title, &card.Content, &card.CreatedAt, &card.UserName)
				card.IsOwner = card.UserID == c.UserID

				hub.BroadcastJSON(map[string]interface{}{
					"type":    "new_card",
					"payload": card,
				})

			case "new_reply":
				content := strings.TrimSpace(msg.Content)
				if content == "" || msg.CardID == 0 {
					return
				}

				var id int
				err := database.DB.QueryRow(
					"INSERT INTO forum_replies (card_id, user_id, content) VALUES ($1, $2, $3) RETURNING id",
					msg.CardID, c.UserID, content,
				).Scan(&id)
				if err != nil {
					return
				}

				var reply models.ForumReply
				database.DB.QueryRow(`
					SELECT fr.id, fr.card_id, fr.user_id, fr.content, fr.created_at, u.name
					FROM forum_replies fr
					INNER JOIN users u ON u.id = fr.user_id
					WHERE fr.id = $1
				`, id).Scan(&reply.ID, &reply.CardID, &reply.UserID, &reply.Content, &reply.CreatedAt, &reply.UserName)
				reply.IsOwner = reply.UserID == c.UserID

				hub.BroadcastJSON(map[string]interface{}{
					"type":    "new_reply",
					"payload": reply,
				})

			case "delete_card":
				var payload struct {
					ID int `json:"id"`
				}
				if err := json.Unmarshal(data, &payload); err != nil {
					return
				}
				database.DB.Exec(
					"DELETE FROM community_posts WHERE id = $1 AND user_id = $2",
					payload.ID, c.UserID,
				)
				hub.BroadcastJSON(map[string]interface{}{
					"type":    "delete_card",
					"payload": map[string]interface{}{"id": payload.ID},
				})
			}
		},
	}

	hub.Register(client)
	go client.WritePump()
	go client.ReadPump()
}
