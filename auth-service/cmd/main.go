package main

import (
	"net/http"

	"github.com/joho/godotenv"
	_ "github.com/luckermt/forum-app/auth-service/docs"
	"github.com/luckermt/forum-app/auth-service/internal/grpc"
	"github.com/luckermt/forum-app/auth-service/internal/handler"
	"github.com/luckermt/forum-app/auth-service/internal/repository"
	"github.com/luckermt/forum-app/auth-service/internal/service"
	"github.com/luckermt/forum-app/shared/pkg/config"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// @Summary Регистрация пользователя
// @Description Создание нового аккаунта
// @Tags auth
// @Accept json
// @Produce json
// @Param input body models.User true "Данные пользователя"
// @Success 201 {object} models.User
// @Router /register [post]
func main() {
	// 1. Инициализация логгера (самое первое!)
	if err := logger.Init(); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Log.Sync()

	envPath := "D:/Programming_Code/VisualStudioCode/forum-app/.env"
	if err := godotenv.Load(envPath); err != nil {
		logger.Log.Fatal("Error loading .env file",
			zap.String("path", envPath),
			zap.Error(err))
	}

	cfg := config.Load()
	logger.Log.Info("DB connection config",
		zap.String("host", cfg.Postgres.Host),
		zap.String("port", cfg.Postgres.Port),
		zap.String("user", cfg.Postgres.User),
		zap.String("dbname", cfg.Postgres.DBName))

	repo, err := repository.NewPostgresRepository(cfg.Postgres)
	if err != nil {
		logger.Log.Fatal("Failed to initialize repository", zap.Error(err))
	}

	authService := service.NewAuthService(repo, cfg.JWT.SecretKey)
	grpcServer := grpc.NewAuthServer(authService)

	go func() {
		if err := grpcServer.Start(cfg.GRPC.AuthServicePort); err != nil {
			logger.Log.Fatal("Failed to start gRPC server", zap.Error(err))
		}
	}()
	defer grpcServer.Stop()

	authHandler := handler.NewAuthHandler(authService)
	http.HandleFunc("/register", authHandler.Register)
	http.HandleFunc("/login", authHandler.Login)
	http.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	logger.Log.Info("Starting auth service", zap.String("port", cfg.Server.Port))
	if err := http.ListenAndServe(":"+cfg.Server.Port, nil); err != nil {
		logger.Log.Fatal("Failed to start auth service", zap.Error(err))
	}

}
