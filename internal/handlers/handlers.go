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
	return c, c.DB.InitStorage(*log, endPointDB)
}

// GetSongs
// @Summary Получить список песен
// @Description Получение данных библиотеки с возможностью фильтрации по полям (название группы, название песни, дата выпуска) и пагинацией.
// @Tags songs
// @Accept json
// @Produce json
// @Param group query string false "Фильтрация по названию группы"
// @Param song query string false "Фильтрация по названию песни"
// @Param release_date query string false "Фильтрация по дате выпуска (ГГГГ-ММ-ДД)"
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество элементов на странице" default(10)
// @Success 200 {array} model.Song "Список песен"
// @Failure 400 {object} map[string]string "Ошибка при получении списка песен"
// @Router /songs [get]
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

	songs, err := r.DB.GetSongs(c.Request().Context(), group, song, releaseDate, limit, offset)
	if err != nil {
		r.log.Error(err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to fetch songs",
		})
	}
	return c.JSON(http.StatusOK, songs)
}

// AddSong
// @Summary Добавить новую песню
// @Description Добавление новой песни в базу данных.
// @Tags songs
// @Accept json
// @Produce json
// @Param song body model.Song true "Данные новой песни"
// @Success 200 {object} model.Song "Добавленная песня"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 500 {object} map[string]string "Ошибка при добавлении песни"
// @Router /songs [post]
func (r *Handler) AddSong(c echo.Context) error {
	var song model.Song
	if err := c.Bind(&song); err != nil {
		r.log.Error(err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	song, err := r.DB.AddSong(c.Request().Context(), song)
	if err != nil {
		r.log.Error(err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "this song is already exists"})
	}

	return c.JSON(http.StatusOK, song)
}

// GetSongByID
// @Summary Получить информацию о песне по ID
// @Description Извлечение песни по её уникальному идентификатору.
// @Tags songs
// @Accept json
// @Produce json
// @Param id path int true "ID песни"
// @Success 200 {object} model.Song "Информация о песне"
// @Failure 404 {object} map[string]string "Песня не найдена"
// @Failure 500 {object} map[string]string "Ошибка при извлечении песни"
// @Router /songs/{id} [get]
func (r *Handler) GetSongByID(c echo.Context) error {
	id := c.Param("id")

	song, err := r.DB.GetSongByID(c.Request().Context(), id)
	if err != nil {
		r.log.Error(err)
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Song not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to extract the song",
		})
	}

	return c.JSON(http.StatusOK, song)
}

// UpdateSong
// @Summary Обновить песню
// @Description Обновление существующей песни по её ID.
// @Tags songs
// @Accept json
// @Produce json
// @Param id path int true "ID песни"
// @Param song body model.Song true "Обновленные данные песни"
// @Success 200 {object} model.Song "Обновленная песня"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 404 {object} map[string]string "Песня не найдена"
// @Failure 500 {object} map[string]string "Ошибка при обновлении песни"
// @Router /songs/{id} [put]
func (r *Handler) UpdateSong(c echo.Context) error {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		r.log.Error(err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid song ID",
		})
	}

	var song model.Song
	if err = c.Bind(&song); err != nil {
		r.log.Error(err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	song.ID = id

	song, err = r.DB.UpdateSong(c.Request().Context(), song)
	if err != nil {
		r.log.Error(err)
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Song not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update song",
		})
	}

	return c.JSON(http.StatusOK, song)
}

// DeleteSong
// @Summary Удалить песню
// @Description Удаление песни по её ID.
// @Tags songs
// @Accept json
// @Produce json
// @Param id path int true "ID песни"
// @Success 204 {string} string "Песня успешно удалена"
// @Failure 404 {object} map[string]string "Песня не найдена"
// @Failure 500 {object} map[string]string "Ошибка при удалении песни"
// @Router /songs/{id} [delete]
func (r *Handler) DeleteSong(c echo.Context) error {
	id := c.Param("id")

	err := r.DB.DeleteSong(c.Request().Context(), id)
	if err != nil {
		r.log.Error(err)
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

// GetSongVerseByID
// @Summary Получить текст куплета песни по ID
// @Description Извлечение текста определенного куплета песни по её ID. Можно передать номер куплета, по умолчанию извлекается первый куплет.
// @Tags songs
// @Accept json
// @Produce json
// @Param id path int true "ID песни"
// @Param verse query int false "Номер куплета" default(1)
// @Success 200 {object} map[string]string "Текст куплета"
// @Failure 400 {object} map[string]string "Некорректный ID песни"
// @Failure 404 {object} map[string]string "Песня не найдена"
// @Failure 500 {object} map[string]string "Ошибка при получении куплета"
// @Router /songs/{id}/verse [get]
func (r *Handler) GetSongVerseByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		r.log.Error(err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid song ID"})
	}

	verseStr := c.QueryParam("verse")
	verse, err := strconv.Atoi(verseStr)
	if err != nil || verse < 1 {
		verse = r.verseParamDefault
	}

	verseText, err := r.DB.GetSongVerseByID(c.Request().Context(), id, verse)

	if err != nil {
		r.log.Error(err)
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Song not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve verse"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"verse": verseText,
	})
}
