package service

import (
	"testing"
	"time"

	"socialmeli/internal/domain"
	"socialmeli/internal/store"
)

func seedUsersForProduct(st *store.MemoryStore) {
	st.SeedUsers([]domain.User{
		{ID: 1, Name: "Buyer", IsSeller: false},
		{ID: 2, Name: "SellerA", IsSeller: true},
		{ID: 3, Name: "SellerB", IsSeller: true},
	})
}

// helper para montar payload válido
func validPayload() PublishPayload {
	return PublishPayload{
		UserID: 1,
		Date:   "01-01-2026",
		Product: domain.Product{
			ProductID:   10,
			ProductName: "Mouse Gamer",
			Type:        "peripheral",
			Brand:       "BrandX",
			Color:       "Black",
			Notes:       "",
		},
		Category: 1,
		Price:    100.0,
		HasPromo: false,
		Discount: 0,
	}
}

func TestParseDate_Empty(t *testing.T) {
	_, err := parseDate("")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != domain.ErrDateEmpty {
		t.Fatalf("expected ErrDateEmpty, got %v", err)
	}
}

func TestParseDate_InvalidFormat(t *testing.T) {
	_, err := parseDate("2026-01-01")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != ErrDateFormat {
		t.Fatalf("expected ErrDateFormat, got %v", err)
	}
}

func TestParseDate_Success(t *testing.T) {
	dt, err := parseDate("02-01-2026")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if dt.Year() != 2026 || dt.Month() != time.January || dt.Day() != 2 {
		t.Fatalf("unexpected date: %v", dt)
	}
}

func TestPublish_InvalidUserID(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsersForProduct(st)
	svc := NewProductService(st)

	p := validPayload()
	p.UserID = 0

	_, err := svc.Publish(p)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestPublish_InvalidDate(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsersForProduct(st)
	svc := NewProductService(st)

	p := validPayload()
	p.Date = "2026-01-01" // formato errado

	_, err := svc.Publish(p)
	if err != ErrDateFormat {
		t.Fatalf("expected ErrDateFormat, got %v", err)
	}
}

func TestPublish_WithPromo_InvalidDiscount(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsersForProduct(st)
	svc := NewProductService(st)

	p := validPayload()
	p.HasPromo = true
	p.Discount = 0 // inválido

	_, err := svc.Publish(p)
	if err == nil {
		t.Fatalf("expected error for discount <= 0")
	}

	p.Discount = 101 // inválido
	_, err = svc.Publish(p)
	if err == nil {
		t.Fatalf("expected error for discount > 100")
	}
}

func TestPublish_WithPromo_Success(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsersForProduct(st)
	svc := NewProductService(st)

	p := validPayload()
	p.HasPromo = true
	p.Discount = 10

	id, err := svc.Publish(p)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if id == 0 {
		t.Fatalf("expected post id > 0")
	}
}

func TestPublish_InvalidProductID(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsersForProduct(st)
	svc := NewProductService(st)

	p := validPayload()
	p.Product.ProductID = 0

	_, err := svc.Publish(p)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestPublish_EmptyCategory(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsersForProduct(st)
	svc := NewProductService(st)

	p := validPayload()
	p.Category = 0

	_, err := svc.Publish(p)
	if err != domain.ErrCategoryEmpty {
		t.Fatalf("expected ErrCategoryEmpty, got %v", err)
	}
}

func TestPublish_InvalidPrice(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsersForProduct(st)
	svc := NewProductService(st)

	p := validPayload()
	p.Price = 0

	_, err := svc.Publish(p)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestPublish_Success_AddsPost(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsersForProduct(st)
	svc := NewProductService(st)

	p := validPayload()
	id, err := svc.Publish(p)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if id != 1 {
		t.Fatalf("expected post id=1, got %d", id)
	}
}

func TestFollowedLastTwoWeeks_InvalidID(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsersForProduct(st)
	svc := NewProductService(st)

	_, err := svc.FollowedLastTwoWeeks(0, domain.DateDesc)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestFollowedLastTwoWeeks_InvalidOrder(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsersForProduct(st)
	svc := NewProductService(st)

	_, err := svc.FollowedLastTwoWeeks(1, "invalid")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestFollowedLastTwoWeeks_Success_SortsDesc(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsersForProduct(st)
	svc := NewProductService(st)

	// user 1 segue seller 2 e 3
	if err := st.Follow(1, 2); err != nil {
		t.Fatalf("follow error: %v", err)
	}
	if err := st.Follow(1, 3); err != nil {
		t.Fatalf("follow error: %v", err)
	}

	// cria 2 posts em datas diferentes (ambos dentro das 2 semanas)
	now := time.Now()
	_, _ = st.AddPost(domain.Post{
		UserID:   2,
		Date:     now.Add(-2 * 24 * time.Hour),
		DateStr:  "x",
		Product:  domain.Product{ProductID: 1, ProductName: "P1", Type: "t", Brand: "b", Color: "c"},
		Category: 1, Price: 10,
	})
	_, _ = st.AddPost(domain.Post{
		UserID:   3,
		Date:     now.Add(-1 * 24 * time.Hour),
		DateStr:  "x",
		Product:  domain.Product{ProductID: 2, ProductName: "P2", Type: "t", Brand: "b", Color: "c"},
		Category: 1, Price: 10,
	})

	posts, err := svc.FollowedLastTwoWeeks(1, domain.DateDesc)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(posts))
	}

	// DateDesc: mais recente primeiro
	if posts[0].Date.Before(posts[1].Date) {
		t.Fatalf("expected desc order, got %v then %v", posts[0].Date, posts[1].Date)
	}
}

func TestPromoCount_InvalidID(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsersForProduct(st)
	svc := NewProductService(st)

	_, _, err := svc.PromoCount(0)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestPromoCount_UserNotFound(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsersForProduct(st)
	svc := NewProductService(st)

	_, _, err := svc.PromoCount(999)
	if err != store.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestPromoCount_Success(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsersForProduct(st)
	svc := NewProductService(st)

	now := time.Now()
	_, _ = st.AddPost(domain.Post{
		UserID:   2,
		Date:     now,
		DateStr:  "x",
		Product:  domain.Product{ProductID: 1, ProductName: "P1", Type: "t", Brand: "b", Color: "c"},
		Category: 1,
		Price:    10,
		HasPromo: true,
		Discount: 5,
	})
	_, _ = st.AddPost(domain.Post{
		UserID:   2,
		Date:     now,
		DateStr:  "x",
		Product:  domain.Product{ProductID: 2, ProductName: "P2", Type: "t", Brand: "b", Color: "c"},
		Category: 1,
		Price:    10,
		HasPromo: false,
		Discount: 0,
	})

	u, count, err := svc.PromoCount(2)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if u.ID != 2 {
		t.Fatalf("expected user 2, got %+v", u)
	}
	if count != 1 {
		t.Fatalf("expected 1 promo, got %d", count)
	}
}

func TestPromoList_InvalidID(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsersForProduct(st)
	svc := NewProductService(st)

	_, _, err := svc.PromoList(0)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestPromoList_UserNotFound(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsersForProduct(st)
	svc := NewProductService(st)

	_, _, err := svc.PromoList(999)
	if err != store.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestPromoList_Success_SortsDateDesc(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsersForProduct(st)
	svc := NewProductService(st)

	now := time.Now()
	older := now.Add(-48 * time.Hour)

	_, _ = st.AddPost(domain.Post{
		UserID:   2,
		Date:     older,
		DateStr:  "x",
		Product:  domain.Product{ProductID: 1, ProductName: "Old", Type: "t", Brand: "b", Color: "c"},
		Category: 1,
		Price:    10,
		HasPromo: true,
		Discount: 5,
	})
	_, _ = st.AddPost(domain.Post{
		UserID:   2,
		Date:     now,
		DateStr:  "x",
		Product:  domain.Product{ProductID: 2, ProductName: "New", Type: "t", Brand: "b", Color: "c"},
		Category: 1,
		Price:    10,
		HasPromo: true,
		Discount: 2,
	})

	_, posts, err := svc.PromoList(2)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("expected 2 promo posts, got %d", len(posts))
	}
	// DateDesc: mais recente primeiro
	if posts[0].Date.Before(posts[1].Date) {
		t.Fatalf("expected DateDesc order, got %v then %v", posts[0].Date, posts[1].Date)
	}
}
