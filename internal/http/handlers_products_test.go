package http

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"socialmeli/internal/domain"
	"socialmeli/internal/service"

	"github.com/gin-gonic/gin"
)

// Mock que implementa a interface ProductService
type productServiceMock struct {
	PublishFn              func(p service.PublishPayload) (int, error)
	FollowedLastTwoWeeksFn func(userID int, order string) ([]domain.Post, error)
	PromoCountFn           func(userID int) (domain.User, int, error)
	PromoListFn            func(userID int) (domain.User, []domain.Post, error)
	DeleteMyPostFn         func(userID, postID int) error
}

func (m *productServiceMock) Publish(p service.PublishPayload) (int, error) {
	if m.PublishFn == nil {
		return 0, nil
	}
	return m.PublishFn(p)
}

func (m *productServiceMock) FollowedLastTwoWeeks(userID int, order string) ([]domain.Post, error) {
	if m.FollowedLastTwoWeeksFn == nil {
		return nil, nil
	}
	return m.FollowedLastTwoWeeksFn(userID, order)
}

func (m *productServiceMock) PromoCount(userID int) (domain.User, int, error) {
	if m.PromoCountFn == nil {
		return domain.User{}, 0, nil
	}
	return m.PromoCountFn(userID)
}

func (m *productServiceMock) PromoList(userID int) (domain.User, []domain.Post, error) {
	if m.PromoListFn == nil {
		return domain.User{}, nil, nil
	}
	return m.PromoListFn(userID)
}

func (m *productServiceMock) DeleteMyPost(userID, postID int) error {
	if m.DeleteMyPostFn == nil {
		return nil
	}
	return m.DeleteMyPostFn(userID, postID)
}

func TestNewProductHandlers(t *testing.T) {
	ps := &productServiceMock{}
	h := NewProductHandlers(ps)
	if h == nil {
		t.Fatalf("NewProductHandlers retornou nil")
	}
	if h.ps != ps {
		t.Fatalf("NewProductHandlers não setou o service corretamente")
	}
}

func TestProductHandlers_Publish_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewProductHandlers(&productServiceMock{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := bytes.NewBufferString(`{invalid json`)
	req, _ := http.NewRequest(http.MethodPost, "/products/publish", body)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Publish(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	want := `{"error":"JSON inválido"}`
	if w.Body.String() != want {
		t.Fatalf("body = %q, want %q", w.Body.String(), want)
	}
}

func validPublishBody() string {
	// campos mínimos para passar nas validações do service:
	// user_id, date, product.product_id, product_name, type, brand, color, category, price
	return `{
		"user_id": 1,
		"date": "01-01-2024",
		"product": {
			"product_id": 1,
			"product_name": "Produto",
			"type": "Tipo",
			"brand": "Marca",
			"color": "Azul"
		},
		"category": 1,
		"price": 10
	}`
}

func TestProductHandlers_Publish_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotPayload service.PublishPayload
	psMock := &productServiceMock{
		PublishFn: func(p service.PublishPayload) (int, error) {
			gotPayload = p
			return 123, nil
		},
	}

	h := NewProductHandlers(psMock)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := bytes.NewBufferString(validPublishBody())
	req, _ := http.NewRequest(http.MethodPost, "/products/publish", body)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Publish(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	// garante que o handler setou HasPromo=false e Discount=0
	if gotPayload.HasPromo != false || gotPayload.Discount != 0 {
		t.Fatalf("payload após Publish: HasPromo=%v Discount=%v, esperado false/0", gotPayload.HasPromo, gotPayload.Discount)
	}
	wantBody := `{"post_id":123}`
	if w.Body.String() != wantBody {
		t.Fatalf("body = %q, want %q", w.Body.String(), wantBody)
	}
}

func TestProductHandlers_Publish_ServiceErro(t *testing.T) {
	gin.SetMode(gin.TestMode)

	psMock := &productServiceMock{
		PublishFn: func(p service.PublishPayload) (int, error) {
			return 0, errors.New("erro no service")
		},
	}

	h := NewProductHandlers(psMock)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := bytes.NewBufferString(validPublishBody())
	req, _ := http.NewRequest(http.MethodPost, "/products/publish", body)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Publish(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	wantBody := `{"error":"erro no service"}`
	if w.Body.String() != wantBody {
		t.Fatalf("body = %q, want %q", w.Body.String(), wantBody)
	}
}

func TestProductHandlers_PromoPublish_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewProductHandlers(&productServiceMock{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := bytes.NewBufferString(`{invalid json`)
	req, _ := http.NewRequest(http.MethodPost, "/products/promo-pub", body)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.PromoPublish(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	want := `{"error":"JSON inválido"}`
	if w.Body.String() != want {
		t.Fatalf("body = %q, want %q", w.Body.String(), want)
	}
}

func TestProductHandlers_PromoPublish_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)

	psMock := &productServiceMock{
		PublishFn: func(p service.PublishPayload) (int, error) {
			return 456, nil
		},
	}

	h := NewProductHandlers(psMock)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := bytes.NewBufferString(validPublishBody())
	req, _ := http.NewRequest(http.MethodPost, "/products/promo-pub", body)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.PromoPublish(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	wantBody := `{"post_id":456}`
	if w.Body.String() != wantBody {
		t.Fatalf("body = %q, want %q", w.Body.String(), wantBody)
	}
}

func TestProductHandlers_FollowedLastTwoWeeks_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewProductHandlers(&productServiceMock{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/products/followed/abc/list", nil)
	c.Request = req
	c.Params = gin.Params{{Key: "userId", Value: "abc"}}

	h.FollowedLastTwoWeeks(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	want := `{"error":"Parâmetro inválido: userId"}`
	if w.Body.String() != want {
		t.Fatalf("body = %q, want %q", w.Body.String(), want)
	}
}

func TestProductHandlers_FollowedLastTwoWeeks_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)

	psMock := &productServiceMock{
		FollowedLastTwoWeeksFn: func(userID int, order string) ([]domain.Post, error) {
			// não nos importamos com o conteúdo, apenas se retorna 200
			return []domain.Post{{}, {}}, nil
		},
	}

	h := NewProductHandlers(psMock)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/products/followed/10/list?order=date_desc", nil)
	c.Request = req
	c.Params = gin.Params{{Key: "userId", Value: "10"}}

	h.FollowedLastTwoWeeks(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestProductHandlers_PromoCount_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewProductHandlers(&productServiceMock{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/products/promo-pub/count?user_id=abc", nil)
	c.Request = req

	h.PromoCount(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	want := `{"error":"Parâmetro inválido: user_id"}`
	if w.Body.String() != want {
		t.Fatalf("body = %q, want %q", w.Body.String(), want)
	}
}

func TestProductHandlers_PromoCount_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)

	psMock := &productServiceMock{
		PromoCountFn: func(userID int) (domain.User, int, error) {
			return domain.User{ID: 1, Name: "User 1"}, 5, nil
		},
	}

	h := NewProductHandlers(psMock)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/products/promo-pub/count?user_id=1", nil)
	c.Request = req

	h.PromoCount(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestProductHandlers_PromoList_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewProductHandlers(&productServiceMock{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/products/promo-pub/list?user_id=abc", nil)
	c.Request = req

	h.PromoList(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	want := `{"error":"Parâmetro inválido: user_id"}`
	if w.Body.String() != want {
		t.Fatalf("body = %q, want %q", w.Body.String(), want)
	}
}

func TestProductHandlers_PromoList_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)

	psMock := &productServiceMock{
		PromoListFn: func(userID int) (domain.User, []domain.Post, error) {
			return domain.User{ID: 1, Name: "User 1"}, []domain.Post{{}}, nil
		},
	}

	h := NewProductHandlers(psMock)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/products/promo-pub/list?user_id=1", nil)
	c.Request = req

	h.PromoList(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestProductHandlers_DeleteMyPost_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	psMock := &productServiceMock{
		DeleteMyPostFn: func(userID, postID int) error {
			return nil
		},
	}

	h := NewProductHandlers(psMock)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodDelete, "/products/me/123", nil)
	c.Request = req
	c.Set("auth_user_id", 1)
	c.Params = gin.Params{{Key: "postId", Value: "123"}}

	h.DeleteMyPost(c)

	// Gin pode retornar 200 ou 204 para StatusNoContent dependendo da versão
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d or %d, body=%s", w.Code, http.StatusNoContent, http.StatusOK, w.Body.String())
	}
}

func TestProductHandlers_DeleteMyPost_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewProductHandlers(&productServiceMock{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodDelete, "/products/me/123", nil)
	c.Request = req
	c.Params = gin.Params{{Key: "postId", Value: "123"}}

	h.DeleteMyPost(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestProductHandlers_DeleteMyPost_InvalidPostID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewProductHandlers(&productServiceMock{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodDelete, "/products/me/abc", nil)
	c.Request = req
	c.Set("auth_user_id", 1)
	c.Params = gin.Params{{Key: "postId", Value: "abc"}}

	h.DeleteMyPost(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
