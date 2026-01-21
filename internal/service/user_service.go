package service

import (
	"socialmeli/internal/domain"
	"socialmeli/internal/store"
)

type UserService struct {
	st store.Store
}

func NewUserService(st store.Store) *UserService { return &UserService{st: st} }

func (s *UserService) Follow(userID, sellerID int) error {
	if err := domain.ValidateID(userID); err != nil {
		return err
	}
	if err := domain.ValidateID(sellerID); err != nil {
		return err
	}
	if userID == sellerID {
		return domain.ErrIDGreaterThanZero // reusa erro generico
	}
	return s.st.Follow(userID, sellerID)
}

func (s *UserService) Unfollow(userID, sellerID int) error {
	if err := domain.ValidateID(userID); err != nil {
		return err
	}
	if err := domain.ValidateID(sellerID); err != nil {
		return err
	}
	return s.st.Unfollow(userID, sellerID)
}

func (s *UserService) FollowersCount(sellerID int) (domain.User, int, error) {
	if err := domain.ValidateID(sellerID); err != nil {
		return domain.User{}, 0, err
	}
	u, ok := s.st.GetUser(sellerID)
	if !ok {
		return domain.User{}, 0, store.ErrUserNotFound
	}
	f, err := s.st.FollowersOf(sellerID)
	if err != nil {
		return domain.User{}, 0, err
	}
	return u, len(f), nil
}

func (s *UserService) FollowersList(sellerID int, order string) (domain.User, []domain.User, error) {
	if err := domain.ValidateID(sellerID); err != nil {
		return domain.User{}, nil, err
	}
	if err := domain.ValidateOrderForUsers(order); err != nil {
		return domain.User{}, nil, err
	}

	u, ok := s.st.GetUser(sellerID)
	if !ok {
		return domain.User{}, nil, store.ErrUserNotFound
	}

	f, err := s.st.FollowersOf(sellerID)
	if err != nil {
		return domain.User{}, nil, err
	}

	domain.SortUsersByName(f, order)
	return u, f, nil
}

func (s *UserService) FollowedList(userID int, order string) (domain.User, []domain.User, error) {
	if err := domain.ValidateID(userID); err != nil {
		return domain.User{}, nil, err
	}
	if err := domain.ValidateOrderForUsers(order); err != nil {
		return domain.User{}, nil, err
	}

	u, ok := s.st.GetUser(userID)
	if !ok {
		return domain.User{}, nil, store.ErrUserNotFound
	}

	f, err := s.st.FollowedBy(userID)
	if err != nil {
		return domain.User{}, nil, err
	}

	domain.SortUsersByName(f, order)
	return u, f, nil
}

func (s *UserService) ListUsers(order string) ([]domain.User, error) {
	if err := domain.ValidateOrderForUsers(order); err != nil {
		return nil, err
	}
	return s.st.ListUsers(order)
}

func (s *UserService) CreateUser(name string, isSeller bool) (domain.User, error) {
	if err := domain.ValidateTextRequired(name, 40, domain.ErrMaxLen40); err != nil {
		return domain.User{}, err
	}
	return s.st.CreateUser(name, isSeller)
}

// Profile

type Profile struct {
	User            domain.Account `json:"user"`
	FollowersCount  int            `json:"followers_count"`
	FollowedCount   int            `json:"followed_count"`
	PublicationsCnt int            `json:"publications_count"`
}

func (s *UserService) GetProfile(userID int) (Profile, error) {
	if err := domain.ValidateID(userID); err != nil {
		return Profile{}, err
	}
	a, ok := s.st.GetAccount(userID)
	if !ok {
		return Profile{}, store.ErrUserNotFound
	}
	followers, _ := s.st.FollowersOf(userID)
	followed, _ := s.st.FollowedBy(userID)
	posts := s.st.PostsByUser(userID)
	return Profile{User: a, FollowersCount: len(followers), FollowedCount: len(followed), PublicationsCnt: len(posts)}, nil
}

func (s *UserService) PostsByUser(userID int, order string) ([]domain.Post, error) {
	if err := domain.ValidateID(userID); err != nil {
		return nil, err
	}
	if err := domain.ValidateOrderForPosts(order); err != nil {
		return nil, err
	}
	posts := s.st.PostsByUser(userID)
	domain.SortPostsByDate(posts, order)
	return posts, nil
}

func (s *UserService) UpdateAvatar(userID int, avatarURL string) (domain.Account, error) {
	if err := domain.ValidateID(userID); err != nil {
		return domain.Account{}, err
	}
	return s.st.UpdateAvatar(userID, avatarURL)
}
