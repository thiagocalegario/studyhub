package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/thiagocalegario/studyhub/internal/database"
	"github.com/thiagocalegario/studyhub/internal/handlers"
	"github.com/thiagocalegario/studyhub/internal/middleware"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	database.Connect()
	database.RunMigrations()

	mux := http.NewServeMux()

	// Arquivos estáticos
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Rota raiz
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/my-disciplines", http.StatusSeeOther)
	})

	// Auth
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.RegisterPost(w, r)
		} else {
			handlers.RegisterPage(w, r)
		}
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.LoginPost(w, r)
		} else {
			handlers.LoginPage(w, r)
		}
	})

	mux.HandleFunc("/logout", handlers.Logout)

	// Catálogo
	mux.HandleFunc("/catalog", middleware.RequireAuth(handlers.CatalogPage))
	mux.HandleFunc("/catalog/course/", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		handlers.CatalogCoursePage(w, r)
	}))

	// Disciplinas pessoais
	mux.HandleFunc("/my-disciplines", middleware.RequireAuth(handlers.MyDisciplinesPage))
	mux.HandleFunc("/disciplines/add", middleware.RequireAuth(handlers.AddDiscipline))
	mux.HandleFunc("/disciplines/remove", middleware.RequireAuth(handlers.RemoveDiscipline))

	// Tópicos privados
	mux.HandleFunc("/disciplines/topic/add", middleware.RequireAuth(handlers.AddTopic))
	mux.HandleFunc("/disciplines/topic/delete", middleware.RequireAuth(handlers.DeleteTopic))
	mux.HandleFunc("/disciplines/topic/edit/", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.EditTopicPost(w, r)
		} else {
			handlers.EditTopicPage(w, r)
		}
	}))

	// Fórum comunitário
	mux.HandleFunc("/disciplines/community/post/add", middleware.RequireAuth(handlers.AddCommunityPost))
	mux.HandleFunc("/disciplines/community/post/delete", middleware.RequireAuth(handlers.DeleteCommunityPost))

	// Rotas dinâmicas de disciplina — ordem importa: mais específicas primeiro
	mux.HandleFunc("/disciplines/", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/disciplines/")

		// Rotas fixas já registradas — ignora
		if path == "add" || path == "remove" ||
			strings.HasPrefix(path, "topic/") ||
			strings.HasPrefix(path, "community/") {
			http.NotFound(w, r)
			return
		}

		// /disciplines/{id}/community
		if strings.HasSuffix(path, "/community") {
			handlers.CommunityPage(w, r)
			return
		}

		// /disciplines/{id}
		handlers.DisciplineDetailPage(w, r)
	}))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Servidor rodando em http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
