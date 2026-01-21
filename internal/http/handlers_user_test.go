package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"socialmeli/internal/domain"

	"github.com/gin-gonic/gin"
)

/*
	MOCK DO userService
*/

type mockUserService struct {
	followErr   error
	unfollowErr error

	lastUserID   int
	lastSellerID int
	lastOrder    string

	followersCountUser domain.User
	followersCountVal  int
	followersCountErr  error

	followersListUser domain.User
	followersListArr  []domain.User
	followersListErr  error

	followedListUser domain.User
	followedListArr  []domain.User
	followedListErr  error
}

func (m *mockUserService) Follow(userID, sellerID int) error {
	m.lastUserID = userID
	m.lastSellerID = sellerID
	return m.followErr
}

func (m *mockUserService) Unfollow(userID, sellerID int) error {
	m.lastUserID = userID
	m.lastSellerID = sellerID
	return m.unfollowErr
}

func (m *mockUserService) FollowersCount(sellerID int) (domain.User, int, error) {
	return m.followersCountUser, m.followersCountVal, m.followersCountErr
}

func (m *mockUserService) FollowersList(sellerID int, order string) (domain.User, []domain.User, error) {
	m.lastOrder = order
	return m.followersListUser, m.followersListArr, m.followersListErr
}

func (m *mockUserService) FollowedList(userID int, order string) (domain.User, []domain.User, error) {
	m.lastOrder = order
	return m.followedListUser, m.followedListArr, m.followedListErr
}

/*
	HELPERS
*/

func setupUserRouter(h *UserHandlers) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/users/:userId/follow/:userIdToFollow", h.Follow)
	r.POST("/users/:userId/unfollow/:userIdToUnfollow", h.Unfollow)
	r.GET("/users/:userId/followers/count", h.FollowersCount)
	r.GET("/users/:userId/followers/list", h.FollowersList)
	r.GET("/users/:userId/followed/list", h.FollowedList)

	return r
}

func doReq(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(nil))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

/*
	TESTES
*/

func TestFollow_InvalidParam_Returns400(t *testing.T) {
	ms := &mockUserService{}
	h := NewUserHandlersWithService(ms)
	r := setupUserRouter(h)

	w := doReq(r, http.MethodPost, "/users/abc/follow/2")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestFollow_Success(t *testing.T) {
	ms := &mockUserService{}
	h := NewUserHandlersWithService(ms)
	r := setupUserRouter(h)

	w := doReq(r, http.MethodPost, "/users/1/follow/2")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if ms.lastUserID != 1 || ms.lastSellerID != 2 {
		t.Fatalf("unexpected call values user=%d seller=%d", ms.lastUserID, ms.lastSellerID)
	}
}

func TestFollow_ServiceError_Returns400(t *testing.T) {
	ms := &mockUserService{followErr: errors.New("regra de negócio")}
	h := NewUserHandlersWithService(ms)
	r := setupUserRouter(h)

	w := doReq(r, http.MethodPost, "/users/1/follow/2")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUnfollow_InvalidParam_Returns400(t *testing.T) {
	ms := &mockUserService{}
	h := NewUserHandlersWithService(ms)
	r := setupUserRouter(h)

	w := doReq(r, http.MethodPost, "/users/1/unfollow/abc")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUnfollow_Success(t *testing.T) {
	ms := &mockUserService{}
	h := NewUserHandlersWithService(ms)
	r := setupUserRouter(h)

	w := doReq(r, http.MethodPost, "/users/1/unfollow/2")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ms.lastUserID != 1 || ms.lastSellerID != 2 {
		t.Fatalf("unexpected call values")
	}
}

func TestFollowersCount_Success(t *testing.T) {
	ms := &mockUserService{
		followersCountUser: domain.User{ID: 2, Name: "Seller"},
		followersCountVal:  5,
	}
	h := NewUserHandlersWithService(ms)
	r := setupUserRouter(h)

	w := doReq(r, http.MethodGet, "/users/2/followers/count")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp FollowersCountResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response")
	}
	if resp.UserID != 2 || resp.FollowersCount != 5 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestFollowersList_OrderPassedCorrectly(t *testing.T) {
	ms := &mockUserService{
		followersListUser: domain.User{ID: 2, Name: "Seller"},
		followersListArr: []domain.User{
			{ID: 1, Name: "Ana"},
			{ID: 3, Name: "Joao"},
		},
	}
	h := NewUserHandlersWithService(ms)
	r := setupUserRouter(h)

	w := doReq(r, http.MethodGet, "/users/2/followers/list?order=asc")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ms.lastOrder != "asc" {
		t.Fatalf("expected order=asc, got %s", ms.lastOrder)
	}
}

func TestFollowedList_Success(t *testing.T) {
	ms := &mockUserService{
		followedListUser: domain.User{ID: 1, Name: "Buyer"},
		followedListArr: []domain.User{
			{ID: 2, Name: "Seller"},
		},
	}
	h := NewUserHandlersWithService(ms)
	r := setupUserRouter(h)

	w := doReq(r, http.MethodGet, "/users/1/followed/list")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp FollowedListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json response")
	}
	if len(resp.Followed) != 1 || resp.Followed[0].UserID2 != 2 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
