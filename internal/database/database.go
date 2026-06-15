package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL não definida")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Erro ao abrir conexão com banco: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Erro ao conectar ao banco: %v", err)
	}

	DB = db
	fmt.Println("Banco de dados conectado com sucesso.")
}

func RunMigrations() {
	migrations := []string{
		"db/migrations/001_create_users.sql",
		"db/migrations/002_create_disciplines.sql",
		"db/migrations/003_create_user_disciplines.sql",
		"db/migrations/004_seed_disciplines.sql",
	}

	for _, file := range migrations {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Erro ao ler migração %s: %v", file, err)
		}

		if _, err := DB.Exec(string(content)); err != nil {
			log.Fatalf("Erro ao executar migração %s: %v", file, err)
		}

		fmt.Printf("Migração executada: %s\n", file)
	}
}
