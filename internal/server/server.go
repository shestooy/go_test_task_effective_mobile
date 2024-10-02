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

	ZapLog.Debugw("Initializing Echo framework", "endPointServer", endPointServer)

	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	ZapLog.Debug("Applying middlewares")

	e.Use(middlewares.GetLogg(ZapLog))
	e.Use(middleware.Gzip())

	ZapLog.Debug("Defining routes")

	e.GET("/info", h.GetInfo)

	songsGroup := e.Group("/songs")

	songsGroup.GET("", h.GetSongs)
	songsGroup.GET("/:id", h.GetSongByID)
	songsGroup.GET("/:id/verse", h.GetSongVerseByID)

	songsGroup.POST("", h.AddSong)

	songsGroup.PUT("/:id", h.UpdateSong)

	songsGroup.DELETE("/:id", h.DeleteSong)

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
	s.logger.Debug("Closing database connection")
	return s.handler.DB.Close()
}
