package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/agidelle/todo_web/internal/config"
	"github.com/agidelle/todo_web/internal/domain"
	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func NewConn(cfg *config.Config) *Storage {
	//Проверка на наличие БД и миграции
	CheckDB(cfg)

	//Инициализация БД
	db, err := sql.Open(cfg.DBdriver, cfg.DBPath)
	if err != nil {
		log.Fatalf("Ошибка открытия БД: %v", err)
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		log.Fatalf("Ошибка проверки соединения с БД: %v", err)
	}
	return &Storage{db: db}
}

func CheckDB(cfg *config.Config) {
	_, err := os.Stat(cfg.DBPath)
	dbExists := !os.IsNotExist(err)

	//Если файл БД отсутствует выполняем миграции
	if !dbExists {
		if err = RunMigrations(cfg); err != nil {
			os.Remove(cfg.DBPath)
			log.Fatalf("миграции не удались: %v", err)
		}
		log.Println("База данных успешно создана")
	}
}

func RunMigrations(cfg *config.Config) error {
	db, err := sql.Open(cfg.DBdriver, cfg.DBPath)
	if err != nil {
		log.Printf("не удалось открыть БД: %v", err)
	}

	//Миграция для SQLite
	schema := []string{
		`CREATE TABLE IF NOT EXISTS scheduler (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date CHAR(8) NOT NULL DEFAULT '',
			title VARCHAR(128) NOT NULL DEFAULT '',
			comment TEXT NOT NULL DEFAULT '',
			repeat VARCHAR(128) NOT NULL DEFAULT ''
		);`,
		`CREATE INDEX IF NOT EXISTS date_index ON scheduler (date);`,
	}

	for _, query := range schema {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("ошибка выполнения миграции: %v\nЗапрос: %s", err, query)
		}
	}
	return nil
}

func (s *Storage) Close() error {
	err := s.db.Close()
	if err != nil {
		return err
	}
	return nil
}
func (s *Storage) FindTask(filter *domain.Filter) ([]*domain.Task, error) {
	tasks := make([]*domain.Task, 0)
	query := "SELECT id, date, title, comment, repeat FROM scheduler"
	args := []interface{}{}
	conditions := []string{}

	//Добавление условий в зависимости от фильтра
	if filter.ID != nil {
		conditions = append(conditions, "id = ?")
		args = append(args, *filter.ID)
	}
	if filter.SearchTerm != "" {
		searchPattern := "%" + filter.SearchTerm + "%"
		conditions = append(conditions, "(title LIKE ? OR comment LIKE ?)")
		args = append(args, searchPattern, searchPattern)
	}
	if filter.Date != "" {
		conditions = append(conditions, "date = ?")
		args = append(args, filter.Date)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY date"
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()
	for rows.Next() {
		var t domain.Task
		err = rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *Storage) CreateTask(task *domain.Task) (int64, error) {
	res, err := s.db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)", task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *Storage) UpdateTask(task *domain.Task) error {
	query, err := s.db.Exec("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?", task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	count, err := query.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("id задачи не найден в БД")
	}
	return nil
}

func (s *Storage) DeleteTask(id *int) error {
	query, err := s.db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		return err
	}
	count, err := query.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("id задачи не найден в БД")
	}
	return nil
}
