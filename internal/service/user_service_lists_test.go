package service

import (
	"errors"
	"testing"

	"socialmeli/internal/domain"
	"socialmeli/internal/store"
)

/*
Wrappers 100% compatíveis com store.Store:
- embutem store.Store (interface)
- automaticamente possuem TODOS os métodos da interface
- sobrescrevem só o método que queremos forçar erro
*/

type errFollowersStore struct {
	store.Store
	err error
}

func (s errFollowersStore) FollowersOf(sellerID int) ([]domain.User, error) {
	return nil, s.err
}

type errFollowedStore struct {
	store.Store
	err error
}

func (s errFollowedStore) FollowedBy(userID int) ([]domain.User, error) {
	return nil, s.err
}

func TestUserService_FollowersList_Success_Asc(t *testing.T) {
	mem := store.NewMemoryStore()
	mem.SeedUsers([]domain.User{
		{ID: 1, Name: "Buyer", IsSeller: false},
		{ID: 2, Name: "Seller", IsSeller: true},
		{ID: 3, Name: "Zoe", IsSeller: false},
		{ID: 4, Name: "Ana", IsSeller: false},
		{ID: 5, Name: "Bob", IsSeller: false},
	})

	_ = mem.Follow(3, 2)
	_ = mem.Follow(4, 2)
	_ = mem.Follow(5, 2)

	svc := NewUserService(mem)

	_, list, err := svc.FollowersList(2, "name_asc")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 followers, got %d", len(list))
	}

	// asc: Ana, Bob, Zoe
	if list[0].Name != "Ana" || list[1].Name != "Bob" || list[2].Name != "Zoe" {
		t.Fatalf("unexpected asc order: %#v", []string{list[0].Name, list[1].Name, list[2].Name})
	}
}

func TestUserService_FollowersList_ErrorFromStore(t *testing.T) {
	mem := store.NewMemoryStore()
	mem.SeedUsers([]domain.User{
		{ID: 2, Name: "Seller", IsSeller: true},
	})

	// wrapper que força erro em FollowersOf
	st := errFollowersStore{Store: mem, err: errors.New("boom")}
	svc := NewUserService(st)

	_, _, err := svc.FollowersList(2, "")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestUserService_FollowedList_Success_Desc(t *testing.T) {
	mem := store.NewMemoryStore()
	mem.SeedUsers([]domain.User{
		{ID: 1, Name: "Buyer", IsSeller: false},
		{ID: 2, Name: "Ana", IsSeller: true},
		{ID: 3, Name: "Zoe", IsSeller: true},
		{ID: 4, Name: "Bob", IsSeller: true},
	})

	_ = mem.Follow(1, 2)
	_ = mem.Follow(1, 3)
	_ = mem.Follow(1, 4)

	svc := NewUserService(mem)

	_, list, err := svc.FollowedList(1, "name_desc")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 followed, got %d", len(list))
	}

	// desc: Zoe, Bob, Ana
	if list[0].Name != "Zoe" || list[1].Name != "Bob" || list[2].Name != "Ana" {
		t.Fatalf("unexpected desc order: %#v", []string{list[0].Name, list[1].Name, list[2].Name})
	}
}

func TestUserService_FollowedList_ErrorFromStore(t *testing.T) {
	mem := store.NewMemoryStore()
	mem.SeedUsers([]domain.User{
		{ID: 1, Name: "Buyer", IsSeller: false},
	})

	// wrapper que força erro em FollowedBy
	st := errFollowedStore{Store: mem, err: errors.New("boom")}
	svc := NewUserService(st)

	_, _, err := svc.FollowedList(1, "")
	if err == nil {
		t.Fatalf("expected error")
	}
}
