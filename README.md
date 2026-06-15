# 📚 StudyHub

Plataforma de apoio acadêmico desenvolvida em Go para alunos da Universidade de Brasília.
Permite organizar disciplinas, criar anotações privadas e participar de fóruns por disciplina.

---

## Funcionalidades

- Cadastro e autenticação de usuários
- Catálogo de disciplinas organizado por curso e semestre
- Adição de disciplinas de qualquer curso à área pessoal
- Tópicos privados por disciplina com categorias e status
- Fórum comunitário por disciplina (posts públicos entre alunos)
- Contagem de tópicos e posts por disciplina

---

## Tecnologias

- **Backend:** Go (net/http, html/template)
- **Banco de dados:** PostgreSQL
- **Containerização:** Docker e Docker Compose
- **Frontend:** HTML, CSS puro (sem frameworks)

---

## Pré-requisitos

- [Go 1.21+](https://golang.org/dl/)
- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Git](https://git-scm.com/)

---

## Como rodar o projeto

### 1. Clone o repositório

```bash
git clone https://github.com/thiagocalegario/studyhub.git
cd studyhub
```

### 2. Configure as variáveis de ambiente

Copie o arquivo de exemplo e preencha com suas configurações:

```bash
cp .env.example .env
```

Edite o `.env` com suas informações:

```env
DATABASE_URL=postgres://studyhub:studyhub@localhost:5432/studyhub?sslmode=disable
PORT=8080
SESSION_SECRET=escolha_uma_chave_secreta_qualquer
```

### 3. Suba o banco de dados com Docker

```bash
docker-compose up -d
```

Aguarde alguns segundos para o PostgreSQL inicializar completamente.

### 4. Instale as dependências Go

```bash
go mod tidy
```

### 5. Rode o servidor

```bash
go run ./cmd/server/main.go
```

Acesse no navegador: [http://localhost:8080](http://localhost:8080)

---