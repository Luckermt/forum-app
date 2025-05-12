package main

import (
	"github.com/luckermt/shared/pkg/config"
	"github.com/luckermt/shared/pkg/logger"
	"auth-service/internal/handler"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"auth-service/internal/grpc"
	"net/http"
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

	// Инициализация сервиса
	authService := service.NewAuthService(repo, cfg.JWT)

	// Инициализация gRPC сервера
	grpcServer := grpc.NewAuthServer(authService)
	go grpcServer.Start(cfg.GRPC.AuthServicePort)

	// Инициализация HTTP обработчиков
	authHandler := handler.NewAuthHandler(authService)

	// Настройка маршрутов
	http.HandleFunc("/register", authHandler.Register)
	http.HandleFunc("/login", authHandler.Login)

	// Запуск HTTP сервера
	logger.Log.Info("Starting auth service", zap.String("port", cfg.Server.Port))
	if err := http.ListenAndServe(":"+cfg.Server.Port, nil); err != nil {
		logger.Log.Fatal("Failed to start auth service", zap.Error(err))
	}
}