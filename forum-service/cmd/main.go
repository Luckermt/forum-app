package main

import (
	"net/http"
	"time"

	"github.com/joho/godotenv"
	authGrpc "github.com/luckermt/forum-app/forum-service/internal/grpc" // Переименованный импорт
	"github.com/luckermt/forum-app/forum-service/internal/handler"
	"github.com/luckermt/forum-app/forum-service/internal/repository"
	"github.com/luckermt/forum-app/forum-service/internal/service"
	"github.com/luckermt/forum-app/shared/pkg/config"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Инициализация логгера
	if err := logger.Init(); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Log.Sync()

	// Загрузка .env файла
	envPath := "D:/Programming_Code/VisualStudioCode/forum-app/forum-service/.env"
	if err := godotenv.Load(envPath); err != nil {
		logger.Log.Fatal("Error loading .env file",
			zap.String("path", envPath),
			zap.Error(err))
	}

	// Загрузка конфигурации
	cfg := config.Load()
	logger.Log.Info("DB Config",
		zap.String("host", cfg.Postgres.Host),
		zap.String("port", cfg.Postgres.Port),
		zap.String("user", cfg.Postgres.User),
		zap.String("dbname", cfg.Postgres.DBName))

	// Инициализация репозитория
	repo, err := repository.NewPostgresRepository(cfg.Postgres)
	if err != nil {
		logger.Log.Fatal("Failed to initialize repository", zap.Error(err))
	}

	// Инициализация gRPC клиента для auth-service
	grpcConn, err := grpc.Dial(
		"localhost:"+cfg.GRPC.AuthServicePort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		logger.Log.Fatal("Failed to connect to auth service", zap.Error(err))
	}
	defer grpcConn.Close()

	authClient, err := authGrpc.NewAuthClient("localhost:" + cfg.GRPC.AuthServicePort)
	if err != nil {
		logger.Log.Fatal("Failed to initialize auth client", zap.Error(err))
	}
	// Инициализация сервиса
	forumService := service.NewForumService(repo, authClient)

	// Инициализация HTTP обработчиков
	forumHandler := handler.NewForumHandler(forumService)
	chatHandler := handler.NewChatHandler(forumService)

	// Настройка маршрутов
	http.HandleFunc("POST /topics", forumHandler.CreateTopic)
	http.HandleFunc("DELETE /topics/{id}", forumHandler.DeleteTopic)
	http.HandleFunc("GET /topics", forumHandler.GetTopics)
	http.HandleFunc("GET /messages", forumHandler.GetMessages)
	http.HandleFunc("/ws", chatHandler.HandleConnections)

	// Запуск очистки старых сообщений
	go forumService.CleanOldMessages(24 * time.Hour)

	// Запуск HTTP сервера
	logger.Log.Info("Starting forum service", zap.String("port", cfg.Server.Port))
	if err := http.ListenAndServe(":"+cfg.Server.Port, nil); err != nil {
		logger.Log.Fatal("Failed to start forum service", zap.Error(err))
	}
}
