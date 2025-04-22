package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/agidelle/todo_web/internal/domain"
	"github.com/agidelle/todo_web/internal/service"
)

type TaskHandler struct {
	service *service.TaskService
}

func NewHandler(service *service.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: message,
	})
	if err != nil {
		log.Println(err)
	}
}

func sendJSONTasks(w http.ResponseWriter, tasks []domain.Task) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(struct {
		Tasks []domain.Task `json:"tasks"`
	}{
		Tasks: tasks,
	})
	if err != nil {
		log.Println(err)
	}
}

func (h *TaskHandler) AddTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var task domain.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		sendJSONError(w, "ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	//Добавление задачи
	id, code, err := h.service.Create(&task)
	if err != nil {
		sendJSONError(w, err.Error(), code)
		return
	}
	task.ID = strconv.Itoa(int(id))

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]int64{"id": id})
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	_, searchParamExists := queryValues["search"]

	search := r.URL.Query().Get("search")

	if !searchParamExists {
		res, code, err := h.service.GetTasks()
		if err != nil {
			sendJSONError(w, err.Error(), code)
			return
		}
		sendJSONTasks(w, *res)
	} else {
		res, code, err := h.service.Search(search)
		if err != nil {
			sendJSONError(w, err.Error(), code)
			return
		}
		sendJSONTasks(w, *res)
	}
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	searchID := r.URL.Query().Get("id")
	if searchID == "" {
		sendJSONError(w, "не указан идентификатор", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(searchID)
	if err != nil {
		sendJSONError(w, "неправильный формат идентификатора", http.StatusBadRequest)
		return
	}
	task, code, err := h.service.GetTask(id)
	if err != nil {
		sendJSONError(w, err.Error(), code)
		return
	}
	task.ID = searchID
	err = json.NewEncoder(w).Encode(&task)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var task domain.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		sendJSONError(w, "ошибка десериализации JSON", http.StatusBadRequest)
		return
	}
	code, err := h.service.Update(&task)
	if err != nil {
		sendJSONError(w, err.Error(), code)
		return
	}
	err = json.NewEncoder(w).Encode(domain.Task{})
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (h *TaskHandler) Done(w http.ResponseWriter, r *http.Request) {
	searchID := r.URL.Query().Get("id")
	if searchID == "" {
		sendJSONError(w, "не указан идентификатор", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(searchID)
	if err != nil {
		sendJSONError(w, "неправильный формат идентификатора", http.StatusBadRequest)
		return
	}
	code, err := h.service.Done(id)
	if err != nil {
		sendJSONError(w, err.Error(), code)
	}
	err = json.NewEncoder(w).Encode(domain.Task{})
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	searchID := r.URL.Query().Get("id")
	if searchID == "" {
		sendJSONError(w, "не указан идентификатор", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(searchID)
	if err != nil {
		sendJSONError(w, "неправильный формат идентификатора", http.StatusBadRequest)
		return
	}
	code, err := h.service.Delete(id)
	if err != nil {
		sendJSONError(w, fmt.Sprintf("ошибка удаления: %v", err), code)
	}
	err = json.NewEncoder(w).Encode(domain.Task{})
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}

}

func (h *TaskHandler) NextDateHandler(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "application/json")
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	var now time.Time
	if nowStr == "" {
		now = time.Now()
	} else {
		var err error
		now, err = time.Parse("20060102", nowStr)
		if err != nil {
			http.Error(w, "неправильный формат даты", http.StatusBadRequest)
			return
		}
	}
	if dateStr == "" {
		http.Error(w, "отсутствует обязательный параметр date", http.StatusBadRequest)
		return
	}
	_, err := time.Parse("20060102", dateStr)
	if err != nil {
		http.Error(w, "неправильный формат даты в параметре date", http.StatusBadRequest)
		return
	}
	nextDate, err := h.service.NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, fmt.Sprintf("ошибка вычисления следующей даты: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	if nextDate == "" {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.Write([]byte(nextDate))
	}
}
