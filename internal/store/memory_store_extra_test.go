package store

import (
	"testing"
	"time"

	"socialmeli/internal/domain"
)

func TestMemoryStore_ListUsers_Success(t *testing.T) {
	s := NewMemoryStore()
	s.SeedUsers([]domain.User{
		{ID: 1, Name: "Alice", IsSeller: false},
		{ID: 2, Name: "Bob", IsSeller: true},
		{ID: 3, Name: "Charlie", IsSeller: false},
	})

	users, err := s.ListUsers("name_asc")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(users) != 3 {
		t.Fatalf("expected 3 users, got %d", len(users))
	}
	if users[0].Name != "Alice" {
		t.Fatalf("expected Alice first, got %s", users[0].Name)
	}
}

func TestMemoryStore_ListUsers_Desc(t *testing.T) {
	s := NewMemoryStore()
	s.SeedUsers([]domain.User{
		{ID: 1, Name: "Alice", IsSeller: false},
		{ID: 2, Name: "Bob", IsSeller: true},
	})

	users, err := s.ListUsers("name_desc")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if users[0].Name != "Bob" {
		t.Fatalf("expected Bob first, got %s", users[0].Name)
	}
}

func TestMemoryStore_CreateUser_Success(t *testing.T) {
	s := NewMemoryStore()

	u, err := s.CreateUser("New User", true)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if u.Name != "New User" || !u.IsSeller {
		t.Fatalf("unexpected user: %+v", u)
	}

	// verifica que foi criado
	got, ok := s.GetUser(u.ID)
	if !ok {
		t.Fatalf("expected user to exist")
	}
	if got.Name != "New User" {
		t.Fatalf("unexpected user: %+v", got)
	}
}

func TestMemoryStore_GetAccount_Success(t *testing.T) {
	s := NewMemoryStore()
	acc, _ := s.CreateAccount("User", "user@example.com", "hash", false)

	got, ok := s.GetAccount(acc.ID)
	if !ok {
		t.Fatalf("expected account to exist")
	}
	if got.Email != "user@example.com" {
		t.Fatalf("unexpected account: %+v", got)
	}
}

func TestMemoryStore_GetAccount_NotFound(t *testing.T) {
	s := NewMemoryStore()

	_, ok := s.GetAccount(999)
	if ok {
		t.Fatalf("expected account not to exist")
	}
}

func TestMemoryStore_GetAccountByEmail_Success(t *testing.T) {
	s := NewMemoryStore()
	acc, _ := s.CreateAccount("User", "user@example.com", "hash", false)

	got, ok := s.GetAccountByEmail("user@example.com")
	if !ok {
		t.Fatalf("expected account to exist")
	}
	if got.ID != acc.ID {
		t.Fatalf("unexpected account: %+v", got)
	}
}

func TestMemoryStore_GetAccountByEmail_NotFound(t *testing.T) {
	s := NewMemoryStore()

	_, ok := s.GetAccountByEmail("nonexistent@example.com")
	if ok {
		t.Fatalf("expected account not to exist")
	}
}

func TestMemoryStore_CreateAccount_Success(t *testing.T) {
	s := NewMemoryStore()

	acc, err := s.CreateAccount("User", "user@example.com", "hash123", true)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if acc.Name != "User" || acc.Email != "user@example.com" || !acc.IsSeller {
		t.Fatalf("unexpected account: %+v", acc)
	}
	if acc.PasswordHash != "hash123" {
		t.Fatalf("unexpected password hash")
	}
}

func TestMemoryStore_CreateAccount_DuplicateEmail(t *testing.T) {
	s := NewMemoryStore()
	_, _ = s.CreateAccount("User1", "user@example.com", "hash1", false)

	_, err := s.CreateAccount("User2", "user@example.com", "hash2", false)
	if err != ErrEmailTaken {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestMemoryStore_UpdateAvatar_Success(t *testing.T) {
	s := NewMemoryStore()
	acc, _ := s.CreateAccount("User", "user@example.com", "hash", false)

	updated, err := s.UpdateAvatar(acc.ID, "/static/avatars/1.jpg")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if updated.AvatarURL != "/static/avatars/1.jpg" {
		t.Fatalf("unexpected avatar: %s", updated.AvatarURL)
	}

	// verifica persistência
	got, ok := s.GetAccount(acc.ID)
	if !ok || got.AvatarURL != "/static/avatars/1.jpg" {
		t.Fatalf("avatar not persisted")
	}
}

func TestMemoryStore_UpdateAvatar_NotFound(t *testing.T) {
	s := NewMemoryStore()

	_, err := s.UpdateAvatar(999, "/static/avatars/1.jpg")
	if err != ErrAccountNotFound {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestMemoryStore_PostsByUser_Success(t *testing.T) {
	s := NewMemoryStore()
	s.SeedUsers([]domain.User{
		{ID: 1, Name: "User", IsSeller: true},
	})

	_, _ = s.AddPost(domain.Post{
		UserID:   1,
		Date:     time.Now(),
		DateStr:  "01-01-2026",
		Product:  domain.Product{ProductID: 1, ProductName: "P1", Type: "t", Brand: "b", Color: "c"},
		Category: 1,
		Price:    10,
	})
	_, _ = s.AddPost(domain.Post{
		UserID:   1,
		Date:     time.Now(),
		DateStr:  "02-01-2026",
		Product:  domain.Product{ProductID: 2, ProductName: "P2", Type: "t", Brand: "b", Color: "c"},
		Category: 1,
		Price:    20,
	})

	posts := s.PostsByUser(1)
	if len(posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(posts))
	}
}

func TestMemoryStore_PostsByUser_Empty(t *testing.T) {
	s := NewMemoryStore()
	s.SeedUsers([]domain.User{
		{ID: 1, Name: "User", IsSeller: true},
	})

	posts := s.PostsByUser(1)
	if len(posts) != 0 {
		t.Fatalf("expected 0 posts, got %d", len(posts))
	}
}

func TestMemoryStore_DeletePost_Success(t *testing.T) {
	s := NewMemoryStore()
	s.SeedUsers([]domain.User{
		{ID: 1, Name: "User", IsSeller: true},
	})

	postID, _ := s.AddPost(domain.Post{
		UserID:   1,
		Product:  domain.Product{ProductID: 1, ProductName: "P1", Type: "t", Brand: "b", Color: "c"},
		Category: 1,
		Price:    10,
	})

	err := s.DeletePost(1, postID)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	posts := s.PostsByUser(1)
	if len(posts) != 0 {
		t.Fatalf("expected 0 posts after delete, got %d", len(posts))
	}
}

func TestMemoryStore_DeletePost_NotFound(t *testing.T) {
	s := NewMemoryStore()
	s.SeedUsers([]domain.User{
		{ID: 1, Name: "User", IsSeller: true},
	})

	err := s.DeletePost(1, 999)
	if err != ErrPostNotFound {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
}

func TestMemoryStore_DeletePost_WrongUser(t *testing.T) {
	s := NewMemoryStore()
	s.SeedUsers([]domain.User{
		{ID: 1, Name: "User1", IsSeller: true},
		{ID: 2, Name: "User2", IsSeller: true},
	})

	postID, _ := s.AddPost(domain.Post{
		UserID:   1,
		Product:  domain.Product{ProductID: 1, ProductName: "P1", Type: "t", Brand: "b", Color: "c"},
		Category: 1,
		Price:    10,
	})

	err := s.DeletePost(2, postID)
	if err != ErrPostForbidden {
		t.Fatalf("expected ErrPostForbidden, got %v", err)
	}
}
