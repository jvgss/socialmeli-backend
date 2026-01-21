package test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"socialmeli/internal/domain"
	apphttp "socialmeli/internal/http"
	"socialmeli/internal/service"
	"socialmeli/internal/store"
)

func newTestServer() *httptest.Server {
	st := store.NewMemoryStore()
	st.SeedUsers([]domain.User{
		{ID: 123, Name: "user123"},
		{ID: 234, Name: "vendedor1"},
	})
	us := service.NewUserService(st)
	ps := service.NewProductService(st)
	as := service.NewAuthService(st)
	r := apphttp.NewRouter(us, ps, as)
	return httptest.NewServer(r)
}

func Test_IT_US0001_Follow(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/users/123/follow/234", nil)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, res.StatusCode)
}

func Test_IT_US0002_FollowersCount(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	req1, _ := http.NewRequest(http.MethodPost, srv.URL+"/users/123/follow/234", nil)
	_, _ = http.DefaultClient.Do(req1)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/users/234/followers/count", nil)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, res.StatusCode)
}

func Test_IT_US0005_Publish(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	body := []byte(`{
	  "user_id": 234,
	  "date": "29-04-2021",
	  "product": {
	    "product_id": 1,
	    "product_name": "Cadeira Gamer",
	    "type": "Gamer",
	    "brand": "Racer",
	    "color": "Red Black",
	    "notes": "Special Edition"
	  },
	  "category": 100,
	  "price": 1500.50
	}`)

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/products/publish", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, res.StatusCode)
}
