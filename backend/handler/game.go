package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/irham/topup-backend/model"
	"github.com/irham/topup-backend/repository"
)

type GameHandler struct {
	games *repository.GameRepo
}

func NewGameHandler(games *repository.GameRepo) *GameHandler {
	return &GameHandler{games: games}
}

func (h *GameHandler) ListGames(c *gin.Context) {
	ctx := c.Request.Context()
	categorySlug := strings.TrimSpace(c.Query("category"))

	var (
		games []model.Game
		err   error
	)
	if categorySlug != "" {
		games, err = h.games.ListGamesByCategory(ctx, categorySlug)
	} else {
		games, err = h.games.ListGames(ctx)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Gagal mengambil daftar game"))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(games))
}

func (h *GameHandler) GetGameBySlug(c *gin.Context) {
	slug := c.Param("slug")
	game, err := h.games.GetGameBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Gagal mengambil game"))
		return
	}
	if game == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse("Game tidak ditemukan"))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(game))
}

func (h *GameHandler) SearchGames(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if len(q) < 2 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse("Query minimal 2 karakter"))
		return
	}
	games, err := h.games.SearchGames(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Gagal mencari game"))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(games))
}

func (h *GameHandler) ListCategories(c *gin.Context) {
	cats, err := h.games.ListCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Gagal mengambil kategori"))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(cats))
}
