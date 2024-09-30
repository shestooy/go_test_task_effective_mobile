package storage

import (
	"context"
	"database/sql"
	"errors"
	"go_test_effective_mobile/internal/model"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	"go.uber.org/zap"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type IStorage interface {
	InitStorage(logger zap.SugaredLogger, EndPointDB string) error
	initMigrations() error
	Ping(ctx context.Context) error
	GetSongs(ctx context.Context, group, song, releaseDate string, limit, offset int) ([]model.Song, error)
	AddSong(ctx context.Context, song model.Song) (model.Song, error)
	GetSongByID(ctx context.Context, id string) (model.Song, error)
	DeleteSong(ctx context.Context, id string) error
	UpdateSong(ctx context.Context, song model.Song) (model.Song, error)
	GetSongVerseByID(ctx context.Context, id, verse int) (string, error)
	Close() error
}

type Storage struct {
	db     *sql.DB
	logger zap.SugaredLogger
}

func (s *Storage) InitStorage(logger zap.SugaredLogger, EndPointDB string) error {
	var err error
	s.db, err = sql.Open("pgx", EndPointDB)
	if err != nil {
		s.logger.Info(zap.Error(err))
		return err
	}
	s.logger = logger
	return s.initMigrations()
}

func (s *Storage) initMigrations() error {
	driver, err := pgx.WithInstance(s.db, &pgx.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	if err != nil {
		s.logger.Info(zap.Error(err))
		return err
	}
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		s.logger.Info(zap.Error(err))
		return err
	}
	return nil
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Storage) GetSongs(ctx context.Context, group, song, releaseDate string, limit, offset int) ([]model.Song, error) {
	query := squirrel.Select("*").From("songs").OrderBy("id").Limit(uint64(limit)).Offset(uint64(offset))

	if group != "" {
		query = query.Where(squirrel.Eq{"group_name": group})
	}

	if song != "" {
		query = query.Where(squirrel.Eq{"song": song})
	}

	if releaseDate != "" {
		query = query.Where(squirrel.Eq{"release_date": releaseDate})
	}

	sqlString, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		s.logger.Info(zap.Error(err))
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, sqlString, args...)
	if err != nil {
		s.logger.Info(zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	songs := make([]model.Song, 0)
	for rows.Next() {
		var song model.Song
		if err = rows.Scan(&song.ID, &song.Group, &song.Song, &song.ReleaseDate, &song.Text, &song.Link); err != nil {
			return nil, err
		}
		songs = append(songs, song)
	}

	return songs, nil
}

func (s *Storage) AddSong(ctx context.Context, song model.Song) (model.Song, error) {
	query := squirrel.Insert("songs").Columns("group_name", "song", "release_date", "text", "link").
		Values(song.Group, song.Song, song.ReleaseDate, song.Text, song.Link).
		Suffix("ON CONFLICT DO NOTHING RETURNING id, group_name, song, release_date, text, link;")

	sqlString, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		s.logger.Info(zap.Error(err))
		return song, err
	}

	row := s.db.QueryRowContext(ctx, sqlString, args...)
	var addedSong model.Song
	if err := row.Scan(&addedSong.ID, &addedSong.Group, &addedSong.Song, &addedSong.ReleaseDate, &addedSong.Text, &addedSong.Link); err != nil {
		s.logger.Info(zap.Error(err))
		return song, err
	}

	return addedSong, nil
}

func (s *Storage) GetSongByID(ctx context.Context, id string) (model.Song, error) {
	query := squirrel.Select("*").From("songs").Where(squirrel.Eq{"id": id})
	sqlString, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		s.logger.Info(zap.Error(err))
		return model.Song{}, err
	}

	row := s.db.QueryRowContext(ctx, sqlString, args...)
	var song model.Song
	if err := row.Scan(&song.ID, &song.Group, &song.Song, &song.ReleaseDate, &song.Text, &song.Link); err != nil {
		s.logger.Info(zap.Error(err))
		return song, err
	}

	return song, nil
}

func (s *Storage) DeleteSong(ctx context.Context, id string) error {
	query := squirrel.Delete("songs").Where(squirrel.Eq{"id": id})
	sqlString, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		s.logger.Info(zap.Error(err))
		return err
	}

	_, err = s.db.ExecContext(ctx, sqlString, args...)
	return err
}

func (s *Storage) UpdateSong(ctx context.Context, song model.Song) (model.Song, error) {
	query := squirrel.Update("songs").
		Set("group_name", song.Group).
		Set("song", song.Song).
		Set("release_date", song.ReleaseDate).
		Set("text", song.Text).
		Set("link", song.Link).
		Where(squirrel.Eq{"id": song.ID}).
		Suffix("RETURNING id, group_name, song, release_date, text, link;")

	sqlString, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		s.logger.Info(zap.Error(err))
		return song, err
	}

	row := s.db.QueryRowContext(ctx, sqlString, args...)
	var updatedSong model.Song
	if err := row.Scan(&updatedSong.ID, &updatedSong.Group, &updatedSong.Song, &updatedSong.ReleaseDate, &updatedSong.Text, &updatedSong.Link); err != nil {
		s.logger.Info(zap.Error(err))
		return song, err
	}

	return updatedSong, nil
}

func (s *Storage) GetSongVerseByID(ctx context.Context, id, verse int) (string, error) {
	query := squirrel.Select("text").From("songs").Where(squirrel.Eq{"id": id})
	sqlString, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		s.logger.Info(zap.Error(err))
		return "", err
	}

	row := s.db.QueryRowContext(ctx, sqlString, args...)
	var text string
	if err := row.Scan(&text); err != nil {
		s.logger.Info(zap.Error(err))
		return "", err
	}

	verses := strings.Split(text, "\n\n")

	if len(verses) < verse {
		err = errors.New("verse not found")
		s.logger.Info(zap.Error(err))
		return "", err
	}

	return verses[verse-1], nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
