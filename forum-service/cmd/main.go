package main

import (
	"net/http"
	"time"

	"github.com/joho/godotenv"

	_ "github.com/luckermt/forum-app/forum-service/docs"                 // Важно!
	authGrpc "github.com/luckermt/forum-app/forum-service/internal/grpc" // Переименованный импорт
	"github.com/luckermt/forum-app/forum-service/internal/handler"
	"github.com/luckermt/forum-app/forum-service/internal/repository"
	"github.com/luckermt/forum-app/forum-service/internal/service"
	"github.com/luckermt/forum-app/shared/pkg/config"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	httpSwagger "github.com/swaggo/http-swagger" // пакет для UI
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @title Forum Service API
// @version 1.0
// @description API для управления темами и сообщениями форума
// @host localhost:8081
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {

	if err := logger.Init(); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Log.Sync()

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

	forumService := service.NewForumService(repo, authClient)

	forumHandler := handler.NewForumHandler(forumService)
	chatHandler := handler.NewChatHandler(forumService)

	// Настройка маршрутов
	http.HandleFunc("POST /topics", forumHandler.CreateTopic)
	http.HandleFunc("DELETE /topics/{id}", forumHandler.DeleteTopic)
	http.HandleFunc("GET /topics", forumHandler.GetTopics)
	http.HandleFunc("GET /messages", forumHandler.GetMessages)
	http.HandleFunc("/ws", chatHandler.HandleConnections)
	http.Handle("/swagger/", httpSwagger.WrapHandler)

	go forumService.CleanOldMessages(24 * time.Hour)

	logger.Log.Info("Starting forum service", zap.String("port", cfg.Server.Port))
	if err := http.ListenAndServe(":"+cfg.Server.Port, nil); err != nil {
		logger.Log.Fatal("Failed to start forum service", zap.Error(err))
	}

}
