package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/agidelle/todo_web/internal/domain"
	"github.com/agidelle/todo_web/internal/service"
)

var errorMap = map[error]int{
	domain.ErrID:             http.StatusBadRequest,
	domain.ErrBadTitle:       http.StatusBadRequest,
	domain.ErrDate:           http.StatusBadRequest,
	domain.ErrInternalServer: http.StatusInternalServerError,
}

type TaskHandler struct {
	service *service.TaskService
}

func NewHandler(service *service.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func sendJSONError(w http.ResponseWriter, customErr *domain.CustomError) {
	w.WriteHeader(customErr.Code)
	var errorMessage string
	if customErr.ErrStorage != nil {
		errorMessage = fmt.Sprintf("%v: %v", customErr.Err, customErr.ErrStorage)
	} else {
		errorMessage = fmt.Sprintf("%v", customErr.Err)
	}

	err := json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: errorMessage,
	})
	if err != nil {
		log.Println(err)
	}
}

func sendJSONTasks(w http.ResponseWriter, tasks []*domain.Task) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(struct {
		Tasks []*domain.Task `json:"tasks"`
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
		sendJSONError(w, domain.NewCustomError(http.StatusBadRequest, errors.New("ошибка десериализации JSON"), nil))
		return
	}

	//Добавление задачи
	id, cErr := h.service.Create(&task)
	if cErr != nil {
		if code, ok := errorMap[cErr.Err]; ok {
			cErr.Code = code
		} else {
			cErr.Code = http.StatusInternalServerError
		}
		sendJSONError(w, cErr)
		return
	}

	task.ID = strconv.Itoa(int(id))

	w.WriteHeader(http.StatusCreated)
	err := json.NewEncoder(w).Encode(map[string]int64{"id": id})
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	var filter domain.Filter
	queryValues := r.URL.Query()
	_, searchParamExists := queryValues["search"]

	filter.SearchTerm = r.URL.Query().Get("search")

	if !searchParamExists {
		res, cErr := h.service.GetTasks(&filter)
		if cErr != nil {
			if code, ok := errorMap[cErr.Err]; ok {
				cErr.Code = code
			} else {
				cErr.Code = http.StatusInternalServerError
			}
			sendJSONError(w, cErr)
			return
		}
		sendJSONTasks(w, res)
	} else {
		res, cErr := h.service.Search(&filter)
		if cErr != nil {
			if code, ok := errorMap[cErr.Err]; ok {
				cErr.Code = code
			} else {
				cErr.Code = http.StatusInternalServerError
			}
			sendJSONError(w, cErr)
			return
		}
		sendJSONTasks(w, res)
	}
}

// Для использования FindAll нужна маленькая корректировка
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	var filter domain.Filter
	searchID := r.URL.Query().Get("id")
	if searchID == "" {
		sendJSONError(w, domain.NewCustomError(http.StatusBadRequest, domain.ErrID, nil))
		return
	}
	id, err := strconv.Atoi(searchID)
	if err != nil {
		sendJSONError(w, domain.NewCustomError(http.StatusBadRequest, domain.ErrID, err))
		return
	}
	filter.ID = &id
	task, cErr := h.service.GetTask(&filter)
	if cErr != nil {
		if code, ok := errorMap[cErr.Err]; ok {
			cErr.Code = code
		} else {
			cErr.Code = http.StatusInternalServerError
		}
		sendJSONError(w, cErr)
		return
	}
	//Корректировка: task[0].ID = searchID, проверка на len есть в FindAll
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
		sendJSONError(w, domain.NewCustomError(http.StatusBadRequest, errors.New("ошибка десериализации JSON"), err))
		return
	}
	cErr := h.service.Update(&task)
	if cErr != nil {
		if code, ok := errorMap[cErr.Err]; ok {
			cErr.Code = code
		} else {
			cErr.Code = http.StatusInternalServerError
		}
		sendJSONError(w, cErr)
		return
	}
	err := json.NewEncoder(w).Encode(domain.Task{})
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (h *TaskHandler) Done(w http.ResponseWriter, r *http.Request) {
	var filter domain.Filter
	searchID := r.URL.Query().Get("id")
	if searchID == "" {
		sendJSONError(w, domain.NewCustomError(http.StatusBadRequest, domain.ErrID, nil))
		return
	}
	id, err := strconv.Atoi(searchID)
	if err != nil {
		sendJSONError(w, domain.NewCustomError(http.StatusBadRequest, domain.ErrID, err))
		return
	}
	filter.ID = &id
	cErr := h.service.Done(&filter)
	if cErr != nil {
		if code, ok := errorMap[cErr.Err]; ok {
			cErr.Code = code
		} else {
			cErr.Code = http.StatusInternalServerError
		}
		sendJSONError(w, cErr)
		return
	}

	err = json.NewEncoder(w).Encode(domain.Task{})
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	searchID := r.URL.Query().Get("id")
	if searchID == "" {
		sendJSONError(w, domain.NewCustomError(http.StatusBadRequest, domain.ErrID, nil))
		return
	}
	id, err := strconv.Atoi(searchID)
	if err != nil {
		sendJSONError(w, domain.NewCustomError(http.StatusBadRequest, domain.ErrID, err))
		return
	}
	cErr := h.service.Delete(id)
	if cErr != nil {
		if code, ok := errorMap[cErr.Err]; ok {
			cErr.Code = code
		} else {
			cErr.Code = http.StatusInternalServerError
		}
		sendJSONError(w, cErr)
		return
	}

	err = json.NewEncoder(w).Encode(domain.Task{})
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}

}

func (h *TaskHandler) NextDateHandler(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, domain.ErrDate.Error(), http.StatusBadRequest)
			return
		}
	}
	if dateStr == "" {
		http.Error(w, domain.ErrDate.Error(), http.StatusBadRequest)
		return
	}
	_, err := time.Parse("20060102", dateStr)
	if err != nil {
		http.Error(w, domain.ErrDate.Error(), http.StatusBadRequest)
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
