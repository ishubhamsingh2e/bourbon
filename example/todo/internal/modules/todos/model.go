package todos

import "time"

type Todo struct {
	ID        string    `json:"id" db:"id,primary"`
	UserID    string    `json:"user_id" db:"user_id,not_null"`
	Title     string    `json:"title" db:"title,not_null"`
	Completed bool      `json:"completed" db:"completed" bourbon:"default:false"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (t *Todo) TableName() string {
	return "todos"
}
