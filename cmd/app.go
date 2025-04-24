package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/agidelle/todo_web/internal/api"
	"github.com/agidelle/todo_web/internal/config"
	"github.com/agidelle/todo_web/internal/service"
	"github.com/agidelle/todo_web/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite"
)

type App struct {
	cfg     *config.Config
	handler *api.TaskHandler
}

func Initialize() (*App, *storage.Storage) {
	//Загружаем конфиг из .env или переменных окружения
	cfg, err := config.LoadCfg()
	if err != nil {
		log.Fatalf("Ошибка загрузки файла конфигурации: %v", err)
	}

	db := storage.NewConn(cfg)
	svc := service.NewService(db)
	handler := api.NewHandler(svc)

	return &App{
		cfg:     cfg,
		handler: handler,
	}, db
}

func (a *App) Start() *http.Server {
	fmt.Println("Starting server...")
	authEnabled := a.cfg.Password != ""

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)

	r.Handle("/*", http.FileServer(http.Dir("web")))
	r.Get("/api/nextdate", a.handler.NextDateHandler)
	r.Post("/api/signin", a.handler.Login(a.cfg.Password, a.cfg.JWTKey))

	r.Group(func(r chi.Router) {
		if authEnabled {
			r.Use(a.handler.JWTMiddleware(a.cfg.Password, a.cfg.JWTKey))
		}
		r.Get("/api/tasks", a.handler.GetTasks)
		r.Get("/api/task", a.handler.GetTask)
		r.Post("/api/task", a.handler.AddTask)
		r.Put("/api/task", a.handler.UpdateTask)
		r.Delete("/api/task", a.handler.DeleteTask)
		r.Post("/api/task/done", a.handler.Done)
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", a.cfg.Port),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	go func() {
		fmt.Printf("Listening on port %d.\n", a.cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Error starting server: %s\n", err)
		}
	}()
	return server
}

func (a *App) Stop(server *http.Server, db *storage.Storage) {
	fmt.Println("\nShutting down server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v\n", err)
	} else {
		fmt.Println("Server stopped gracefully.")
	}
	if err := db.Close(); err != nil {
		log.Printf("Error closing database connection: %v\n", err)
	} else {
		fmt.Println("Database closed.")
	}
}
