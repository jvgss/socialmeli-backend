package store

import (
	"socialmeli/internal/domain"
	"time"
)

type Store interface {
	// users (social)
	GetUser(id int) (domain.User, bool)
	ListUsers(order string) ([]domain.User, error)
	CreateUser(name string, isSeller bool) (domain.User, error)

	// auth/accounts
	GetAccount(id int) (domain.Account, bool)
	GetAccountByEmail(email string) (domain.Account, bool)
	CreateAccount(name, email, passwordHash string, isSeller bool) (domain.Account, error)
	UpdateAvatar(userID int, avatarURL string) (domain.Account, error)

	// follow graph
	Follow(userID, sellerID int) error
	Unfollow(userID, sellerID int) error
	FollowersOf(sellerID int) ([]domain.User, error)
	FollowedBy(userID int) ([]domain.User, error)

	// posts
	AddPost(p domain.Post) (int, error)
	// DeletePost remove uma publicacao do usuario. Se o post nao existir ou nao pertencer ao usuario, retorna erro.
	DeletePost(userID, postID int) error
	PostsFromSellersSince(sellerIDs []int, since time.Time) []domain.Post
	PromoPostsBySeller(sellerID int) []domain.Post
	PostsByUser(userID int) []domain.Post
}
