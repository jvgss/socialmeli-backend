package http

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PageMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// parsePageLimit lê page/limit da query. Retorna ok=false se valores forem inválidos.
func parsePageLimit(c *gin.Context, defaultLimit, maxLimit int) (page int, limit int, ok bool) {
	page = 1
	limit = defaultLimit

	if p := c.Query("page"); p != "" {
		v, err := strconv.Atoi(p)
		if err != nil || v < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro inválido: page"})
			return 0, 0, false
		}
		page = v
	}

	if l := c.Query("limit"); l != "" {
		v, err := strconv.Atoi(l)
		if err != nil || v < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro inválido: limit"})
			return 0, 0, false
		}
		if v > maxLimit {
			v = maxLimit
		}
		limit = v
	}

	return page, limit, true
}

func paginateSlice[T any](items []T, page, limit int) (out []T, meta PageMeta) {
	total := len(items)
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if limit <= 0 {
		limit = total
	}
	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	out = items[start:end]
	meta = PageMeta{Page: page, Limit: limit, Total: total, TotalPages: totalPages}
	return out, meta
}
