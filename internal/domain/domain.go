package domain

type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date,omitempty"`
	Title   string `json:"title,omitempty"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

type Filter struct {
	ID         *int
	SearchTerm string
	Date       string
	Limit      int
}

type TaskRepository interface {
	FindTask(filter *Filter) ([]*Task, error)
	CreateTask(task *Task) (int64, error)
	UpdateTask(task *Task) error
	DeleteTask(id *int) error
	Close() error
}
