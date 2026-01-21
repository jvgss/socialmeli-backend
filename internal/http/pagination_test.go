package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPaginateSlice_Basic(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	out, meta := paginateSlice(items, 1, 3)
	require.Equal(t, []int{1, 2, 3}, out)
	require.Equal(t, 1, meta.Page)
	require.Equal(t, 3, meta.Limit)
	require.Equal(t, 10, meta.Total)
	require.Equal(t, 4, meta.TotalPages)

	out, meta = paginateSlice(items, 4, 3)
	require.Equal(t, []int{10}, out)
	require.Equal(t, 4, meta.Page)
}

func TestPaginateSlice_LimitZeroAndOutOfRange(t *testing.T) {
	items := []string{"a", "b", "c"}
	out, meta := paginateSlice(items, 1, 0)
	require.Equal(t, items, out)
	require.Equal(t, 3, meta.Total)

	out, _ = paginateSlice(items, 10, 2)
	require.Empty(t, out)
}

func TestParsePageLimit_ValidAndInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		page, limit, ok := parsePageLimit(c, 20, 50)
		if !ok {
			return
		}
		c.JSON(200, gin.H{"page": page, "limit": limit})
	})

	// ok + cap
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?page=2&limit=999", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), "\"page\":2")
	require.Contains(t, w.Body.String(), "\"limit\":50")

	// page inválida
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/?page=0", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)

	// limit inválido
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/?limit=-1", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}
