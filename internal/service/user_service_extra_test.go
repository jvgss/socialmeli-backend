package service

import (
	"testing"
	"time"

	"socialmeli/internal/domain"
	"socialmeli/internal/store"
)

func TestUserService_CreateUser_Success(t *testing.T) {
	st := store.NewMemoryStore()
	svc := NewUserService(st)

	u, err := svc.CreateUser("New User", true)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if u.Name != "New User" || !u.IsSeller {
		t.Fatalf("unexpected user: %+v", u)
	}
}

func TestUserService_CreateUser_InvalidName(t *testing.T) {
	st := store.NewMemoryStore()
	svc := NewUserService(st)

	_, err := svc.CreateUser("", false)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestUserService_GetProfile_Success(t *testing.T) {
	st := store.NewMemoryStore()
	acc, _ := st.CreateAccount("User", "user@example.com", "hash", false)

	// adiciona seguidores e seguidos
	st.SeedUsers([]domain.User{
		{ID: 2, Name: "Follower1", IsSeller: false},
		{ID: 3, Name: "Followed1", IsSeller: true},
	})
	_ = st.Follow(2, acc.ID)
	_ = st.Follow(acc.ID, 3)

	// adiciona posts
	_, _ = st.AddPost(domain.Post{
		UserID:   acc.ID,
		Date:     time.Now(),
		DateStr:  "01-01-2026",
		Product:  domain.Product{ProductID: 1, ProductName: "P1", Type: "t", Brand: "b", Color: "c"},
		Category: 1,
		Price:    10,
	})

	svc := NewUserService(st)
	prof, err := svc.GetProfile(acc.ID)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if prof.FollowersCount != 1 || prof.FollowedCount != 1 || prof.PublicationsCnt != 1 {
		t.Fatalf("unexpected profile: %+v", prof)
	}
}

func TestUserService_GetProfile_UserNotFound(t *testing.T) {
	st := store.NewMemoryStore()
	svc := NewUserService(st)

	_, err := svc.GetProfile(999)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestUserService_PostsByUser_Success(t *testing.T) {
	st := store.NewMemoryStore()
	acc, _ := st.CreateAccount("User", "user@example.com", "hash", false)

	now := time.Now()
	_, _ = st.AddPost(domain.Post{
		UserID:   acc.ID,
		Date:     now.Add(-24 * time.Hour),
		DateStr:  "01-01-2026",
		Product:  domain.Product{ProductID: 1, ProductName: "P1", Type: "t", Brand: "b", Color: "c"},
		Category: 1,
		Price:    10,
	})
	_, _ = st.AddPost(domain.Post{
		UserID:   acc.ID,
		Date:     now,
		DateStr:  "02-01-2026",
		Product:  domain.Product{ProductID: 2, ProductName: "P2", Type: "t", Brand: "b", Color: "c"},
		Category: 1,
		Price:    20,
	})

	svc := NewUserService(st)
	posts, err := svc.PostsByUser(acc.ID, "date_desc")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(posts))
	}
	// mais recente primeiro
	if posts[0].Product.ProductID != 2 {
		t.Fatalf("expected post 2 first, got %d", posts[0].Product.ProductID)
	}
}

func TestUserService_PostsByUser_InvalidOrder(t *testing.T) {
	st := store.NewMemoryStore()
	acc, _ := st.CreateAccount("User", "user@example.com", "hash", false)
	svc := NewUserService(st)

	_, err := svc.PostsByUser(acc.ID, "invalid")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestUserService_UpdateAvatar_Success(t *testing.T) {
	st := store.NewMemoryStore()
	acc, _ := st.CreateAccount("User", "user@example.com", "hash", false)
	svc := NewUserService(st)

	updated, err := svc.UpdateAvatar(acc.ID, "/static/avatars/1.jpg")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if updated.AvatarURL != "/static/avatars/1.jpg" {
		t.Fatalf("unexpected avatar: %s", updated.AvatarURL)
	}
}

func TestUserService_UpdateAvatar_InvalidID(t *testing.T) {
	st := store.NewMemoryStore()
	svc := NewUserService(st)

	_, err := svc.UpdateAvatar(0, "/static/avatars/1.jpg")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestUserService_ListUsers_Success(t *testing.T) {
	st := store.NewMemoryStore()
	st.SeedUsers([]domain.User{
		{ID: 1, Name: "Alice", IsSeller: false},
		{ID: 2, Name: "Bob", IsSeller: true},
	})
	svc := NewUserService(st)

	users, err := svc.ListUsers("name_asc")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Name != "Alice" {
		t.Fatalf("expected Alice first, got %s", users[0].Name)
	}
}

func TestUserService_ListUsers_InvalidOrder(t *testing.T) {
	st := store.NewMemoryStore()
	svc := NewUserService(st)

	_, err := svc.ListUsers("invalid")
	if err == nil {
		t.Fatalf("expected error")
	}
}
