package domain

type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date,omitempty"`
	Title   string `json:"title,omitempty"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

type TaskRepository interface {
	GetTask(id int) (*Task, error)
	GetTasks(limit int) (*[]Task, error)
	SearchTask(word string, limit int) (*[]Task, error)
	SearchForDate(date string, limit int) (*[]Task, error)
	CreateTask(task *Task) (int64, error)
	UpdateTask(task *Task) error
	DeleteTask(id int) error
}
