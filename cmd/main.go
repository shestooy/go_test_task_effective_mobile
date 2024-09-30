package main

import (
	"context"
	"go_test_effective_mobile/internal/config"
	"go_test_effective_mobile/internal/server"
	"os"
	"os/signal"
	"syscall"
)

// @title Реализация онлайн библиотеки песен
// @version 1.0
// @description Тестовое задание для Effective mobile
// @host localhost:8080
// @BasePath /
func main() {
	cfg := config.NewConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverApp, err := server.New(cfg.LogLevel, cfg.ServerEndPoint, cfg.DataBaseEndPoint,
		cfg.DefaultLimit, cfg.DefaultPage, cfg.DefaultVerse)
	if err != nil {
		panic(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- serverApp.Start()
	}()

	select {
	case err = <-serverErr:
		if err != nil {
			panic(err)
		}
	case <-stop:
		if err = serverApp.Stop(ctx); err != nil {
			panic(err)
		}

	}
}
