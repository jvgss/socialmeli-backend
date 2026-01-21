package store

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"socialmeli/internal/domain"
)

var (
	ErrUserNotFound    = errors.New("Usuário inexistente.")
	ErrPostNotFound    = errors.New("Publicação inexistente.")
	ErrPostForbidden   = errors.New("Você não pode apagar uma publicação que não é sua.")
	ErrEmailTaken      = errors.New("E-mail já cadastrado.")
	ErrAccountNotFound = errors.New("Conta inexistente.")
)

type MemoryStore struct {
	mu sync.RWMutex

	users map[int]domain.User
	// contas com credenciais
	accounts       map[int]domain.Account
	accountByEmail map[string]int
	nextUserID     int

	// followers: sellerId -> set(userId)
	followers map[int]map[int]struct{}
	// followed: userId -> set(sellerId)
	followed map[int]map[int]struct{}

	posts      []domain.Post
	nextPostID int
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:          map[int]domain.User{},
		accounts:       map[int]domain.Account{},
		accountByEmail: map[string]int{},
		nextUserID:     1,
		followers:      map[int]map[int]struct{}{},
		followed:       map[int]map[int]struct{}{},
		posts:          []domain.Post{},
		nextPostID:     1,
	}
}

func (s *MemoryStore) SeedUsers(users []domain.User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, u := range users {
		s.users[u.ID] = u
		if u.ID >= s.nextUserID {
			s.nextUserID = u.ID + 1
		}
	}
}

func (s *MemoryStore) GetUser(id int) (domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}

func (s *MemoryStore) ListUsers(order string) ([]domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.User, 0, len(s.users))
	for _, u := range s.users {
		out = append(out, u)
	}
	// order: name_asc | name_desc
	sort.Slice(out, func(i, j int) bool {
		ai := strings.ToLower(out[i].Name)
		aj := strings.ToLower(out[j].Name)
		if order == domain.NameDesc {
			return ai > aj
		}
		return ai < aj
	})
	return out, nil
}

func (s *MemoryStore) CreateUser(name string, isSeller bool) (domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextUserID
	s.nextUserID++
	u := domain.User{ID: id, Name: name, IsSeller: isSeller}
	s.users[id] = u
	return u, nil
}

func (s *MemoryStore) GetAccount(id int) (domain.Account, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.accounts[id]
	return a, ok
}

func (s *MemoryStore) GetAccountByEmail(email string) (domain.Account, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.accountByEmail[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return domain.Account{}, false
	}
	a, ok := s.accounts[id]
	return a, ok
}

func (s *MemoryStore) CreateAccount(name, email, passwordHash string, isSeller bool) (domain.Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	normEmail := strings.ToLower(strings.TrimSpace(email))
	if _, exists := s.accountByEmail[normEmail]; exists {
		return domain.Account{}, ErrEmailTaken
	}
	id := s.nextUserID
	s.nextUserID++

	acc := domain.Account{
		ID:           id,
		Name:         name,
		Email:        normEmail,
		IsSeller:     isSeller,
		CreatedAt:    time.Now().UTC(),
		PasswordHash: passwordHash,
	}
	s.accounts[id] = acc
	s.accountByEmail[normEmail] = id

	// também cria o User “social” para follow/list
	s.users[id] = domain.User{ID: id, Name: name, IsSeller: isSeller}
	return acc, nil
}

func (s *MemoryStore) UpdateAvatar(userID int, avatarURL string) (domain.Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	acc, ok := s.accounts[userID]
	if !ok {
		return domain.Account{}, ErrAccountNotFound
	}
	acc.AvatarURL = avatarURL
	s.accounts[userID] = acc
	return acc, nil
}

func (s *MemoryStore) Follow(userID, sellerID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[userID]; !ok {
		return ErrUserNotFound
	}
	if _, ok := s.users[sellerID]; !ok {
		return ErrUserNotFound
	}

	if s.followers[sellerID] == nil {
		s.followers[sellerID] = map[int]struct{}{}
	}
	if s.followed[userID] == nil {
		s.followed[userID] = map[int]struct{}{}
	}

	s.followers[sellerID][userID] = struct{}{}
	s.followed[userID][sellerID] = struct{}{}
	return nil
}

func (s *MemoryStore) Unfollow(userID, sellerID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[userID]; !ok {
		return ErrUserNotFound
	}
	if _, ok := s.users[sellerID]; !ok {
		return ErrUserNotFound
	}

	if s.followers[sellerID] != nil {
		delete(s.followers[sellerID], userID)
	}
	if s.followed[userID] != nil {
		delete(s.followed[userID], sellerID)
	}
	return nil
}

func (s *MemoryStore) FollowersOf(sellerID int) ([]domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.users[sellerID]; !ok {
		return nil, ErrUserNotFound
	}

	set := s.followers[sellerID]
	out := make([]domain.User, 0, len(set))
	for uid := range set {
		if u, ok := s.users[uid]; ok {
			out = append(out, u)
		}
	}
	return out, nil
}

func (s *MemoryStore) FollowedBy(userID int) ([]domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.users[userID]; !ok {
		return nil, ErrUserNotFound
	}

	set := s.followed[userID]
	out := make([]domain.User, 0, len(set))
	for sid := range set {
		if u, ok := s.users[sid]; ok {
			out = append(out, u)
		}
	}
	return out, nil
}

func (s *MemoryStore) AddPost(p domain.Post) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[p.UserID]; !ok {
		return 0, ErrUserNotFound
	}

	p.PostID = s.nextPostID
	s.nextPostID++
	s.posts = append(s.posts, p)
	return p.PostID, nil
}

func (s *MemoryStore) PostsFromSellersSince(sellerIDs []int, since time.Time) []domain.Post {
	s.mu.RLock()
	defer s.mu.RUnlock()

	set := map[int]struct{}{}
	for _, id := range sellerIDs {
		set[id] = struct{}{}
	}

	out := []domain.Post{}
	for _, p := range s.posts {
		if _, ok := set[p.UserID]; ok && (p.Date.Equal(since) || p.Date.After(since)) {
			out = append(out, p)
		}
	}
	return out
}

func (s *MemoryStore) PromoPostsBySeller(sellerID int) []domain.Post {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Post{}
	for _, p := range s.posts {
		if p.UserID == sellerID && p.HasPromo {
			out = append(out, p)
		}
	}
	return out
}

func (s *MemoryStore) PostsByUser(userID int) []domain.Post {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Post{}
	for _, p := range s.posts {
		if p.UserID == userID {
			out = append(out, p)
		}
	}
	return out
}

func (s *MemoryStore) DeletePost(userID, postID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// garante que usuario exista
	if _, ok := s.users[userID]; !ok {
		return ErrUserNotFound
	}

	idx := -1
	for i, p := range s.posts {
		if p.PostID == postID {
			if p.UserID != userID {
				return ErrPostForbidden
			}
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrPostNotFound
	}
	// remove mantendo ordem
	s.posts = append(s.posts[:idx], s.posts[idx+1:]...)
	return nil
}
