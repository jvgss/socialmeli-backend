package service

import (
	"testing"

	"socialmeli/internal/domain"
	"socialmeli/internal/store"
)

func seedUsers(st *store.MemoryStore) {
	st.SeedUsers([]domain.User{
		{ID: 1, Name: "Buyer", IsSeller: false},
		{ID: 2, Name: "Seller", IsSeller: true},
		{ID: 3, Name: "Zoe", IsSeller: false},
		{ID: 4, Name: "Ana", IsSeller: false},
		{ID: 5, Name: "Bob", IsSeller: false},
	})
}

type failingFollowersStore struct {
	*store.MemoryStore
	err error
}

func (s *failingFollowersStore) FollowersOf(sellerID int) ([]domain.User, error) {
	return nil, s.err
}

type failingFollowedStore struct {
	*store.MemoryStore
	err error
}

func (s *failingFollowedStore) FollowedBy(userID int) ([]domain.User, error) {
	return nil, s.err
}

func TestNewUserService(t *testing.T) {
	st := store.NewMemoryStore()
	svc := NewUserService(st)
	if svc == nil {
		t.Fatalf("service should not be nil")
	}
}

func TestFollow_InvalidIDs(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsers(st)
	svc := NewUserService(st)

	if err := svc.Follow(0, 2); err == nil {
		t.Fatalf("expected error for invalid userID")
	}
	if err := svc.Follow(1, 0); err == nil {
		t.Fatalf("expected error for invalid sellerID")
	}
}

func TestFollow_Success(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsers(st)
	svc := NewUserService(st)

	if err := svc.Follow(1, 2); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	// valida via FollowersCount
	_, count, err := svc.FollowersCount(2)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 follower, got %d", count)
	}
}

func TestUnfollow_InvalidIDs(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsers(st)
	svc := NewUserService(st)

	if err := svc.Unfollow(-1, 2); err == nil {
		t.Fatalf("expected error for invalid userID")
	}
	if err := svc.Unfollow(1, -2); err == nil {
		t.Fatalf("expected error for invalid sellerID")
	}
}

func TestUnfollow_Success(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsers(st)
	svc := NewUserService(st)

	_ = svc.Follow(1, 2)
	if err := svc.Unfollow(1, 2); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	_, count, err := svc.FollowersCount(2)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 followers, got %d", count)
	}
}

func TestFollowersCount_InvalidID(t *testing.T) {
	st := store.NewMemoryStore()
	seedUsers(st)
	svc := NewUserService(st)

	_, _, err := svc.FollowersCount(0)
	if err == nil {
		t.Fatalf("expected error")
	}
}
