package models

import "time"

type User struct {
	ID           int
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type Course struct {
	ID         int
	Name       string
	University string
}

type Discipline struct {
	ID       int
	Code     string
	Name     string
	Semester int
	Workload int
	CourseID int
	Added    bool
}

type UserDiscipline struct {
	ID           int
	UserID       int
	DisciplineID int
	AddedAt      time.Time
}

type Topic struct {
	ID           int
	UserID       int
	DisciplineID int
	Title        string
	Content      string
	Tag          string
	Status       string
	CreatedAt    time.Time
	UserName     string
}

type CommunityPost struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	DisciplineID int       `json:"discipline_id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
	UserName     string    `json:"user_name"`
	IsOwner      bool      `json:"is_owner"`
}

type ForumReply struct {
	ID        int       `json:"id"`
	CardID    int       `json:"card_id"`
	UserID    int       `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UserName  string    `json:"user_name"`
	IsOwner   bool      `json:"is_owner"`
}
