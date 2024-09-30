package server

import (
	"context"
	"go_test_effective_mobile/internal/handlers"

	echoSwagger "github.com/swaggo/echo-swagger"

	_ "go_test_effective_mobile/docs"
	"go_test_effective_mobile/internal/logger"
	"go_test_effective_mobile/internal/middlewares"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type Server struct {
	server         *echo.Echo
	logger         *zap.SugaredLogger
	endPointServer string
	handler        *handlers.Handler
}

func New(logLvl, endPointServer, endPointDB string, limitParam, pageParam, verseParam int) (*Server, error) {
	ZapLog, err := logger.InitLogger(logLvl)
	if err != nil {
		return nil, err
	}

	h, err := handlers.NewHandler(ZapLog, limitParam, pageParam, verseParam, endPointDB)
	if err != nil {
		return nil, err
	}

	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(middlewares.GetLogg(*ZapLog))
	e.Use(middleware.Gzip())

	songsGroup := e.Group("/songs")
	songsGroup.GET("", h.GetSongs)
	songsGroup.POST("", h.AddSong)
	songsGroup.GET("/:id", h.GetSongByID)
	songsGroup.PUT("/:id", h.UpdateSong)
	songsGroup.DELETE("/:id", h.DeleteSong)
	songsGroup.GET("/:id/verses", h.GetSongVerseByID)

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return &Server{server: e, logger: ZapLog, endPointServer: endPointServer, handler: h}, nil
}

func (s *Server) Start() error {
	s.logger.Info("Server starting on: ", s.endPointServer)
	return s.server.Start(s.endPointServer)
}
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Server shutting down")
	if err := s.server.Server.Shutdown(ctx); err != nil {
		return err
	}
	return s.handler.DB.Close()
}
