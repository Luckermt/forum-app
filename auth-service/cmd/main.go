package main

import (
	"net/http"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/luckermt/forum-app/auth-service/docs"
	"github.com/luckermt/forum-app/auth-service/internal/grpc"
	"github.com/luckermt/forum-app/auth-service/internal/handler"
	"github.com/luckermt/forum-app/auth-service/internal/repository"
	"github.com/luckermt/forum-app/auth-service/internal/service"
	"github.com/luckermt/forum-app/shared/pkg/config"
	"github.com/luckermt/forum-app/shared/pkg/logger"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// @title Auth Service API
// @version 1.0
// @description API для аутентификации и управления пользователями
// @host localhost:8080
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
	if err := godotenv.Load(); err != nil {
		logger.Log.Fatal("Error loading .env file", zap.Error(err))
	}

	cfg := config.Load()
	logger.Log.Info("Starting auth service",
		zap.String("port", cfg.Server.Port),
		zap.String("grpc_port", cfg.GRPC.AuthServicePort))

	// Инициализация репозитория
	repo, err := repository.NewPostgresRepository(cfg.Postgres)
	if err != nil {
		logger.Log.Fatal("Failed to initialize repository", zap.Error(err))
	}

	// Создание сервиса
	authService := service.NewAuthService(repo, cfg.JWT.SecretKey)

	// Запуск gRPC сервера
	grpcServer := grpc.NewAuthServer(authService)
	go func() {
		if err := grpcServer.Start(cfg.GRPC.AuthServicePort); err != nil {
			logger.Log.Fatal("Failed to start gRPC server", zap.Error(err))
		}
	}()
	defer grpcServer.Stop()

	// Создание HTTP обработчиков
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(authService)

	// Настройка маршрутов
	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("POST /api/register", authHandler.Register)
	mux.HandleFunc("POST /api/login", authHandler.Login)
	mux.HandleFunc("GET /api/me", authHandler.GetCurrentUser)

	// User routes
	mux.HandleFunc("GET /api/users/{id}", userHandler.GetUser)
	mux.HandleFunc("PUT /api/users/{id}", userHandler.UpdateUser)

	// Swagger
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Настройка CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080", "http://127.0.0.1:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            true,
	})

	// Запуск HTTP сервера
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      corsHandler.Handler(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Log.Info("Starting HTTP server on port " + cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil {
		logger.Log.Fatal("Failed to start HTTP server", zap.Error(err))
	}
}
