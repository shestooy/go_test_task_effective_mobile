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
	GetInfo(ctx context.Context, group, song string) (model.Song, error)
	Close() error
}

type Storage struct {
	db     *sql.DB
	logger zap.SugaredLogger
}

func (s *Storage) InitStorage(logger zap.SugaredLogger, EndPointDB string) error {
	var err error
	logger.Debug("Initializing storage with DB endpoint:", EndPointDB)
	s.db, err = sql.Open("pgx", EndPointDB)
	if err != nil {
		logger.Info(zap.Error(err))
		return err
	}
	s.logger = logger
	return s.initMigrations()
}

func (s *Storage) initMigrations() error {
	s.logger.Debug("Initializing migrations...")
	driver, err := pgx.WithInstance(s.db, &pgx.Config{})
	if err != nil {
		s.logger.Info(zap.Error(err))
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	if err != nil {
		s.logger.Info(zap.Error(err))
		return err
	}
	s.logger.Debug("Running migrations...")
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		s.logger.Info(zap.Error(err))
		return err
	}
	s.logger.Debug("Migrations completed successfully")
	return nil
}

func (s *Storage) Ping(ctx context.Context) error {
	s.logger.Debug("Pinging database...")
	err := s.db.PingContext(ctx)
	if err != nil {
		s.logger.Info(zap.Error(err))
	}
	return err
}

func (s *Storage) GetSongs(ctx context.Context, group, song, releaseDate string, limit, offset int) ([]model.Song, error) {
	s.logger.Debugw("Fetching songs with filters", "group", group, "song", song, "releaseDate", releaseDate, "limit", limit, "offset", offset)

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
	s.logger.Debug("Generated SQL:", sqlString, "args:", args)

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
		s.logger.Debug("Scanned song:", song)
		songs = append(songs, song)
	}

	return songs, nil
}

func (s *Storage) AddSong(ctx context.Context, song model.Song) (model.Song, error) {
	s.logger.Debugw("Adding new song", "song", song)

	query := squirrel.Insert("songs").Columns("group_name", "song", "release_date", "text", "link").
		Values(song.Group, song.Song, song.ReleaseDate, song.Text, song.Link).
		Suffix("ON CONFLICT (group_name, song) DO NOTHING RETURNING id, group_name, song, release_date, text, link;")

	sqlString, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		s.logger.Info(zap.Error(err))
		return song, err
	}
	s.logger.Debug("Generated SQL:", sqlString, "args:", args)

	row := s.db.QueryRowContext(ctx, sqlString, args...)
	var addedSong model.Song
	if err := row.Scan(&addedSong.ID, &addedSong.Group, &addedSong.Song, &addedSong.ReleaseDate, &addedSong.Text, &addedSong.Link); err != nil {
		s.logger.Info(zap.Error(err))
		return song, err
	}
	s.logger.Debug("Added song:", addedSong)

	return addedSong, nil
}

func (s *Storage) GetSongByID(ctx context.Context, id string) (model.Song, error) {
	s.logger.Debug("Fetching song by ID:", id)

	query := squirrel.Select("*").From("songs").Where(squirrel.Eq{"id": id})
	sqlString, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		s.logger.Info(zap.Error(err))
		return model.Song{}, err
	}
	s.logger.Debug("Generated SQL:", sqlString, "args:", args)

	row := s.db.QueryRowContext(ctx, sqlString, args...)
	var song model.Song
	if err := row.Scan(&song.ID, &song.Group, &song.Song, &song.ReleaseDate, &song.Text, &song.Link); err != nil {
		s.logger.Info(zap.Error(err))
		return song, err
	}
	s.logger.Debug("Fetched song:", song)

	return song, nil
}

func (s *Storage) DeleteSong(ctx context.Context, id string) error {
	s.logger.Debug("Deleting song by ID:", id)

	query := squirrel.Delete("songs").Where(squirrel.Eq{"id": id})
	sqlString, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		s.logger.Info(zap.Error(err))
		return err
	}
	s.logger.Debug("Generated SQL:", sqlString, "args:", args)

	_, err = s.db.ExecContext(ctx, sqlString, args...)
	if err != nil {
		s.logger.Info(zap.Error(err))
	}
	return err
}

func (s *Storage) UpdateSong(ctx context.Context, song model.Song) (model.Song, error) {
	s.logger.Debugw("Updating song", "song", song)

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
	s.logger.Debug("Generated SQL:", sqlString, "args:", args)

	row := s.db.QueryRowContext(ctx, sqlString, args...)
	var updatedSong model.Song
	if err := row.Scan(&updatedSong.ID, &updatedSong.Group, &updatedSong.Song, &updatedSong.ReleaseDate, &updatedSong.Text, &updatedSong.Link); err != nil {
		s.logger.Info(zap.Error(err))
		return song, err
	}
	s.logger.Debug("Updated song:", updatedSong)

	return updatedSong, nil
}

func (s *Storage) GetSongVerseByID(ctx context.Context, id, verse int) (string, error) {
	s.logger.Debug("Fetching song verse by ID:", id, "verse:", verse)

	query := squirrel.Select("text").From("songs").Where(squirrel.Eq{"id": id})
	sqlString, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		s.logger.Info(zap.Error(err))
		return "", err
	}
	s.logger.Debug("Generated SQL:", sqlString, "args:", args)

	row := s.db.QueryRowContext(ctx, sqlString, args...)
	var text string
	if err := row.Scan(&text); err != nil {
		s.logger.Info(zap.Error(err))
		return "", err
	}
	s.logger.Debug("Fetched song text:", text)

	verses := strings.Split(text, "\n\n")

	if len(verses) < verse {
		err = errors.New("verse not found")
		s.logger.Info(zap.Error(err))
		return "", err
	}

	s.logger.Debug("Fetched verse:", verses[verse-1])
	return verses[verse-1], nil
}

func (s *Storage) GetInfo(ctx context.Context, group, song string) (model.Song, error) {
	s.logger.Debug("Fetching song info", "group", group, "song", song)

	query := squirrel.Select("release_date", "text", "link").From("songs").Where(squirrel.Eq{"group_name": group, "song": song})
	sqlString, args, err := query.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		s.logger.Info(zap.Error(err))
		return model.Song{}, err
	}
	s.logger.Debug("Generated SQL:", sqlString, "args:", args)

	row := s.db.QueryRowContext(ctx, sqlString, args...)
	var res model.Song
	err = row.Scan(&res.ReleaseDate, &res.Text, &res.Link)
	if err != nil {
		s.logger.Info(zap.Error(err))
	}
	s.logger.Debug("Fetched song info:", res)

	return res, err
}

func (s *Storage) Close() error {
	s.logger.Debug("Closing database connection...")
	err := s.db.Close()
	if err != nil {
		s.logger.Info(zap.Error(err))
	}
	return err
}
