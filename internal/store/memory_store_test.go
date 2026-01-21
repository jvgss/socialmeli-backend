package store

import (
	"testing"
	"time"

	"socialmeli/internal/domain"
)

func newStoreSeeded() *MemoryStore {
	s := NewMemoryStore()
	s.SeedUsers([]domain.User{
		{ID: 1, Name: "Buyer"},
		{ID: 2, Name: "SellerA"},
		{ID: 3, Name: "SellerB"},
	})
	return s
}

func TestSeedUsers_GetUser(t *testing.T) {
	s := NewMemoryStore()
	s.SeedUsers([]domain.User{{ID: 10, Name: "Joao"}})

	u, ok := s.GetUser(10)
	if !ok {
		t.Fatalf("expected user to exist")
	}
	if u.ID != 10 || u.Name != "Joao" {
		t.Fatalf("unexpected user: %+v", u)
	}

	_, ok = s.GetUser(999)
	if ok {
		t.Fatalf("expected user not to exist")
	}
}

func TestFollow_Unfollow(t *testing.T) {
	s := newStoreSeeded()

	// follow ok
	if err := s.Follow(1, 2); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	followers, err := s.FollowersOf(2)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(followers) != 1 || followers[0].ID != 1 {
		t.Fatalf("expected seller 2 to have follower 1, got %+v", followers)
	}

	followed, err := s.FollowedBy(1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(followed) != 1 || followed[0].ID != 2 {
		t.Fatalf("expected user 1 to follow seller 2, got %+v", followed)
	}

	// unfollow ok
	if err := s.Unfollow(1, 2); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	followers, _ = s.FollowersOf(2)
	if len(followers) != 0 {
		t.Fatalf("expected no followers after unfollow, got %+v", followers)
	}
	followed, _ = s.FollowedBy(1)
	if len(followed) != 0 {
		t.Fatalf("expected user 1 to follow nobody after unfollow, got %+v", followed)
	}
}

func TestFollow_UserNotFound(t *testing.T) {
	s := newStoreSeeded()

	if err := s.Follow(999, 2); err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
	if err := s.Follow(1, 999); err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestFollowersOf_FollowedBy_UserNotFound(t *testing.T) {
	s := newStoreSeeded()

	if _, err := s.FollowersOf(999); err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
	if _, err := s.FollowedBy(999); err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestAddPost_AssignsID_AndRequiresUser(t *testing.T) {
	s := newStoreSeeded()

	id1, err := s.AddPost(domain.Post{UserID: 2})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if id1 != 1 {
		t.Fatalf("expected first post id=1, got %d", id1)
	}

	id2, err := s.AddPost(domain.Post{UserID: 2})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if id2 != 2 {
		t.Fatalf("expected second post id=2, got %d", id2)
	}

	_, err = s.AddPost(domain.Post{UserID: 999})
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestPostsFromSellersSince_FiltersBySellerAndDate(t *testing.T) {
	s := newStoreSeeded()

	now := time.Now()
	t0 := now.Add(-48 * time.Hour) // 2 dias atrás
	t1 := now.Add(-24 * time.Hour) // 1 dia atrás
	t2 := now                      // hoje

	// seller 2 posts
	_, _ = s.AddPost(domain.Post{UserID: 2, Date: t0})
	_, _ = s.AddPost(domain.Post{UserID: 2, Date: t2})

	// seller 3 post
	_, _ = s.AddPost(domain.Post{UserID: 3, Date: t1})

	// since = 1 dia atrás -> pega t1 e t2 (e sellerIDs filtro)
	out := s.PostsFromSellersSince([]int{2, 3}, t1)
	if len(out) != 2 {
		t.Fatalf("expected 2 posts, got %d: %+v", len(out), out)
	}
	// garante que nenhum é anterior a t1
	for _, p := range out {
		if p.Date.Before(t1) {
			t.Fatalf("found post before since: %+v", p)
		}
	}

	// filtro por sellers: só seller 2
	out2 := s.PostsFromSellersSince([]int{2}, t1)
	if len(out2) != 1 {
		t.Fatalf("expected 1 post from seller 2 since t1, got %d: %+v", len(out2), out2)
	}
	if out2[0].UserID != 2 {
		t.Fatalf("expected seller=2, got %+v", out2[0])
	}
}

func TestPromoPostsBySeller_OnlyPromoFromSeller(t *testing.T) {
	s := newStoreSeeded()
	now := time.Now()

	// seller 2: 1 promo e 1 normal
	_, _ = s.AddPost(domain.Post{UserID: 2, Date: now, HasPromo: true})
	_, _ = s.AddPost(domain.Post{UserID: 2, Date: now, HasPromo: false})

	// seller 3: promo (não deve entrar quando buscar seller 2)
	_, _ = s.AddPost(domain.Post{UserID: 3, Date: now, HasPromo: true})

	out := s.PromoPostsBySeller(2)
	if len(out) != 1 {
		t.Fatalf("expected 1 promo post for seller 2, got %d: %+v", len(out), out)
	}
	if out[0].UserID != 2 || !out[0].HasPromo {
		t.Fatalf("unexpected promo post: %+v", out[0])
	}
}
