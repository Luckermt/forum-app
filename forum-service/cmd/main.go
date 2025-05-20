package main

import (
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/cors"

	_ "github.com/luckermt/forum-app/forum-service/docs"
	authGrpc "github.com/luckermt/forum-app/forum-service/internal/grpc"
	"github.com/luckermt/forum-app/forum-service/internal/handler"
	"github.com/luckermt/forum-app/forum-service/internal/repository"
	"github.com/luckermt/forum-app/forum-service/internal/service"
	"github.com/luckermt/forum-app/shared/pkg/config"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @title Forum Service API
// @version 1.0
// @description API для управления темами и сообщениями форума
// @host localhost:8081
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Инициализация логгера
	if err := logger.Init(); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Log.Sync()

	// Загрузка конфигурации
	envPath := "D:/Programming_Code/VisualStudioCode/forum-app/forum-service/.env"
	if err := godotenv.Load(envPath); err != nil {
		logger.Log.Fatal("Error loading .env file",
			zap.String("path", envPath),
			zap.Error(err))
	}

	cfg := config.Load()
	logger.Log.Info("Starting forum service with config",
		zap.String("port", cfg.Server.Port),
		zap.String("db_host", cfg.Postgres.Host),
		zap.String("db_name", cfg.Postgres.DBName))

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

	// Создание сервиса
	forumService := service.NewForumService(repo, authClient)

	// Создание обработчиков
	forumHandler := handler.NewForumHandler(forumService)
	chatHandler := handler.NewChatHandler(forumService)
	adminHandler := handler.NewAdminHandler(forumService)

	// Настройка маршрутизатора
	mux := http.NewServeMux()

	// Статические файлы UI
	fs := http.FileServer(http.Dir("../forum-ui/public"))
	mux.Handle("/", fs)

	// API маршруты
	mux.HandleFunc("GET /api/topics", forumHandler.GetTopics)
	mux.HandleFunc("POST /api/topics", forumHandler.CreateTopic)
	mux.HandleFunc("DELETE /api/topics/{id}", forumHandler.DeleteTopic)
	mux.HandleFunc("GET /api/messages", forumHandler.GetMessages)
	mux.HandleFunc("POST /api/messages", forumHandler.CreateMessage)
	mux.HandleFunc("/api/ws", chatHandler.HandleConnections)

	// Админ маршруты
	mux.HandleFunc("GET /api/admin/users", adminHandler.GetUsers)
	mux.HandleFunc("POST /api/admin/users/{id}/block", adminHandler.BlockUser)
	mux.HandleFunc("POST /api/admin/users/{id}/unblock", adminHandler.UnblockUser)

	// Документация Swagger
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Настройка CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8081", "http://127.0.0.1:8081"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		Debug:            true,
	})

	// Запуск очистки старых сообщений
	go forumService.CleanOldMessages(24 * time.Hour)

	// Запуск сервера
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      corsHandler.Handler(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Log.Info("Starting server on port " + cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil {
		logger.Log.Fatal("Failed to start server", zap.Error(err))
	}
}
