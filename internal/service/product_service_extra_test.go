package service

import (
	"testing"

	"socialmeli/internal/domain"
	"socialmeli/internal/store"
)

func TestProductService_DeleteMyPost_Success(t *testing.T) {
	st := store.NewMemoryStore()
	st.SeedUsers([]domain.User{
		{ID: 1, Name: "User", IsSeller: true},
	})
	postID, _ := st.AddPost(domain.Post{
		UserID:   1,
		Product:  domain.Product{ProductID: 1, ProductName: "P1", Type: "t", Brand: "b", Color: "c"},
		Category: 1,
		Price:    10,
	})

	svc := NewProductService(st)
	err := svc.DeleteMyPost(1, postID)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestProductService_DeleteMyPost_InvalidUserID(t *testing.T) {
	st := store.NewMemoryStore()
	svc := NewProductService(st)

	err := svc.DeleteMyPost(0, 1)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestProductService_DeleteMyPost_InvalidPostID(t *testing.T) {
	st := store.NewMemoryStore()
	st.SeedUsers([]domain.User{
		{ID: 1, Name: "User", IsSeller: true},
	})
	svc := NewProductService(st)

	err := svc.DeleteMyPost(1, 0)
	if err == nil {
		t.Fatalf("expected error")
	}
}
