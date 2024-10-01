package handlers

import (
	"database/sql"
	"errors"
	"go_test_effective_mobile/internal/model"
	"go_test_effective_mobile/internal/storage"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Handler struct {
	log               *zap.SugaredLogger
	DB                storage.IStorage
	limitParamDefault int
	pageParamDefault  int
	verseParamDefault int
}

func NewHandler(log *zap.SugaredLogger, limitParam, pageParam, verseParam int, endPointDB string) (*Handler, error) {
	db := &storage.Storage{}
	c := &Handler{log: log, DB: db, limitParamDefault: limitParam, pageParamDefault: pageParam, verseParamDefault: verseParam}
	log.Debug("Initializing new handler with DB endpoint:", endPointDB)
	return c, c.DB.InitStorage(*log, endPointDB)
}

func (r *Handler) GetSongs(c echo.Context) error {
	group := c.QueryParam("group")
	song := c.QueryParam("song")
	releaseDate := c.QueryParam("release_date")

	pageParam := c.QueryParam("page")
	limitParam := c.QueryParam("limit")

	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = r.limitParamDefault
	}
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit < 1 {
		limit = r.pageParamDefault
	}
	offset := (page - 1) * limit

	r.log.Debugw("Fetching songs", "group", group, "song", song, "releaseDate", releaseDate, "limit", limit, "offset", offset)
	songs, err := r.DB.GetSongs(c.Request().Context(), group, song, releaseDate, limit, offset)
	if err != nil {
		r.log.Errorw("Failed to fetch songs", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to fetch songs",
		})
	}
	return c.JSON(http.StatusOK, songs)
}

func (r *Handler) AddSong(c echo.Context) error {
	var song model.Song
	if err := c.Bind(&song); err != nil {
		r.log.Errorw("Failed to bind song", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	r.log.Debugw("Adding new song", "song", song)
	song, err := r.DB.AddSong(c.Request().Context(), song)
	if err != nil {
		r.log.Errorw("Failed to add song", "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "this song already exists"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	r.log.Debug("Song added successfully", "song", song)
	return c.JSON(http.StatusOK, song)
}

func (r *Handler) GetSongByID(c echo.Context) error {
	id := c.Param("id")
	r.log.Debug("Fetching song by ID", "id", id)

	song, err := r.DB.GetSongByID(c.Request().Context(), id)
	if err != nil {
		r.log.Errorw("Failed to fetch song by ID", "id", id, "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Song not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to extract the song",
		})
	}

	r.log.Debug("Song fetched successfully", "song", song)
	return c.JSON(http.StatusOK, song)
}

func (r *Handler) UpdateSong(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		r.log.Errorw("Invalid song ID", "id", idStr, "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid song ID",
		})
	}

	var song model.Song
	if err = c.Bind(&song); err != nil {
		r.log.Errorw("Failed to bind song", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	song.ID = id

	r.log.Debugw("Updating song", "song", song)
	song, err = r.DB.UpdateSong(c.Request().Context(), song)
	if err != nil {
		r.log.Errorw("Failed to update song", "id", id, "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Song not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update song",
		})
	}

	r.log.Debug("Song updated successfully", "song", song)
	return c.JSON(http.StatusOK, song)
}

func (r *Handler) DeleteSong(c echo.Context) error {
	id := c.Param("id")
	r.log.Debug("Deleting song by ID", "id", id)

	err := r.DB.DeleteSong(c.Request().Context(), id)
	if err != nil {
		r.log.Errorw("Failed to delete song", "id", id, "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Song not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete song",
		})
	}
	return c.NoContent(http.StatusNoContent)
}

func (r *Handler) GetSongVerseByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		r.log.Errorw("Invalid song ID", "id", idStr, "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid song ID",
		})
	}

	verseStr := c.QueryParam("verse")
	verse, err := strconv.Atoi(verseStr)
	if err != nil || verse < 1 {
		verse = r.verseParamDefault
	}

	r.log.Debugw("Fetching song verse", "id", id, "verse", verse)
	verseText, err := r.DB.GetSongVerseByID(c.Request().Context(), id, verse)
	if err != nil {
		r.log.Errorw("Failed to fetch song verse", "id", id, "verse", verse, "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Song not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve verse",
		})
	}

	r.log.Debug("Verse fetched successfully", "verseText", verseText)
	return c.JSON(http.StatusOK, map[string]string{
		"verse": verseText,
	})
}

func (r *Handler) GetInfo(c echo.Context) error {
	group := c.QueryParam("group")
	song := c.QueryParam("song")
	if group == "" || song == "" {
		r.log.Errorw("Group or song parameters are missing")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Group or song are required"})
	}

	r.log.Debugw("Fetching song info", "group", group, "song", song)
	info, err := r.DB.GetInfo(c.Request().Context(), group, song)
	if err != nil {
		r.log.Errorw("Failed to fetch song info", "group", group, "song", song, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch song info",
		})
	}
	r.log.Debug("Song info fetched successfully", "info", info)
	return c.JSON(http.StatusOK, info)
}
