package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/irham/topup-backend/model"
	"github.com/irham/topup-backend/repository"
)

type ProductHandler struct {
	games *repository.GameRepo
}

func NewProductHandler(games *repository.GameRepo) *ProductHandler {
	return &ProductHandler{games: games}
}

func (h *ProductHandler) ListProductsByGame(c *gin.Context) {
	gameID, err := uuid.Parse(c.Param("game_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse("game_id tidak valid"))
		return
	}
	products, err := h.games.ListProductsByGame(c.Request.Context(), gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Gagal mengambil produk"))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(products))
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse("id tidak valid"))
		return
	}
	product, err := h.games.GetProductByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse("Gagal mengambil produk"))
		return
	}
	if product == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse("Produk tidak ditemukan"))
		return
	}
	c.JSON(http.StatusOK, model.SuccessResponse(product))
}
