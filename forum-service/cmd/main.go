package main

import (
	"forum-service/internal/grpc"
	"forum-service/internal/handler"
	"forum-service/internal/repository"
	"forum-service/internal/service"
	"net/http"
	"time"

	"github.com/luckermt/shared/pkg/config"
	"github.com/luckermt/shared/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Инициализация логгера
	if err := logger.Init(); err != nil {
		panic(err)
	}
	defer logger.Log.Sync()

	// Загрузка конфигурации
	cfg := config.Load()

	// Инициализация репозитория
	repo, err := repository.NewPostgresRepository(cfg.Postgres)
	if err != nil {
		logger.Log.Fatal("Failed to initialize repository", zap.Error(err))
	}

	// Инициализация gRPC клиента для auth-service
	authClient, err := grpc.NewAuthClient(cfg.GRPC.AuthServicePort)
	if err != nil {
		logger.Log.Fatal("Failed to initialize auth client", zap.Error(err))
	}

	// Инициализация сервиса
	forumService := service.NewForumService(repo, authClient)

	// Инициализация HTTP обработчиков
	forumHandler := handler.NewForumHandler(forumService)
	chatHandler := handler.NewChatHandler(forumService)

	// Настройка маршрутов
	http.HandleFunc("/topics", forumHandler.HandleTopics)
	http.HandleFunc("/topics/create", forumHandler.CreateTopic)
	http.HandleFunc("/topics/delete", forumHandler.DeleteTopic)
	http.HandleFunc("/messages", forumHandler.GetMessages)
	http.HandleFunc("/ws", chatHandler.HandleConnections)

	// Запуск очистки старых сообщений
	go forumService.CleanOldMessages(24 * time.Hour)

	// Запуск HTTP сервера
	logger.Log.Info("Starting forum service", zap.String("port", cfg.Server.Port))
	if err := http.ListenAndServe(":"+cfg.Server.Port, nil); err != nil {
		logger.Log.Fatal("Failed to start forum service", zap.Error(err))
	}
}
