package store

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"socialmeli/internal/domain"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type SQLStore struct {
	db *sql.DB
}

func (s *SQLStore) DeletePost(userID, postID int) error {
	// Apaga apenas se pertencer ao usuario
	res, err := s.db.Exec(`DELETE FROM posts WHERE id=$1 AND user_id=$2`, postID, userID)
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		// pode ser inexistente ou de outro user
		// checa se existe
		var owner int
		err := s.db.QueryRow(`SELECT user_id FROM posts WHERE id=$1`, postID).Scan(&owner)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrPostNotFound
			}
			return err
		}
		return ErrPostForbidden
	}
	return nil
}

func NewSQLStore(dsn string) (*SQLStore, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if err := ApplyMigrations(db, "db/migrations"); err != nil {
		return nil, err
	}
	return &SQLStore{db: db}, nil
}

func (s *SQLStore) GetUser(id int) (domain.User, bool) {
	var u domain.User
	err := s.db.QueryRow(`SELECT id, name, is_seller FROM users WHERE id=$1`, id).
		Scan(&u.ID, &u.Name, &u.IsSeller)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, false
		}
		return domain.User{}, false
	}
	return u, true
}

func (s *SQLStore) Follow(userID, sellerID int) error {
	if _, ok := s.GetUser(userID); !ok {
		return ErrUserNotFound
	}
	if _, ok := s.GetUser(sellerID); !ok {
		return ErrUserNotFound
	}

	_, err := s.db.Exec(`
		INSERT INTO follows (user_id, seller_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, seller_id) DO NOTHING
	`, userID, sellerID)
	return err
}

func (s *SQLStore) Unfollow(userID, sellerID int) error {
	if _, ok := s.GetUser(userID); !ok {
		return ErrUserNotFound
	}
	if _, ok := s.GetUser(sellerID); !ok {
		return ErrUserNotFound
	}

	_, err := s.db.Exec(`DELETE FROM follows WHERE user_id=$1 AND seller_id=$2`, userID, sellerID)
	return err
}

func (s *SQLStore) FollowersOf(sellerID int) ([]domain.User, error) {
	if _, ok := s.GetUser(sellerID); !ok {
		return nil, ErrUserNotFound
	}

	rows, err := s.db.Query(`
		SELECT u.id, u.name, u.is_seller
		FROM follows f
		JOIN users u ON u.id = f.user_id
		WHERE f.seller_id = $1
	`, sellerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Name, &u.IsSeller); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *SQLStore) FollowedBy(userID int) ([]domain.User, error) {
	if _, ok := s.GetUser(userID); !ok {
		return nil, ErrUserNotFound
	}

	rows, err := s.db.Query(`
		SELECT u.id, u.name, u.is_seller
		FROM follows f
		JOIN users u ON u.id = f.seller_id
		WHERE f.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Name, &u.IsSeller); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *SQLStore) AddPost(p domain.Post) (int, error) {
	if _, ok := s.GetUser(p.UserID); !ok {
		return 0, ErrUserNotFound
	}

	var id int
	err := s.db.QueryRow(`
		INSERT INTO posts (
			user_id, date, date_str,
				product_id, product_name, type, brand, color, notes, image_url,
			category, price, has_promo, discount
		) VALUES (
			$1,$2,$3,
				$4,$5,$6,$7,$8,$9,$10,
				$11,$12,$13,$14
		)
		RETURNING id
	`,
		p.UserID, p.Date, p.DateStr,
		p.Product.ProductID, p.Product.ProductName, p.Product.Type, p.Product.Brand, p.Product.Color, p.Product.Notes, p.Product.ImageURL,
		p.Category, p.Price, p.HasPromo, p.Discount,
	).Scan(&id)

	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *SQLStore) PostsFromSellersSince(sellerIDs []int, since time.Time) []domain.Post {
	if len(sellerIDs) == 0 {
		return []domain.Post{}
	}

	// IN ($2,$3...)
	args := make([]any, 0, len(sellerIDs)+1)
	args = append(args, since)

	ph := make([]string, 0, len(sellerIDs))
	for i, id := range sellerIDs {
		args = append(args, id)
		ph = append(ph, fmt.Sprintf("$%d", i+2))
	}

	q := `
		SELECT
			id, user_id, date, date_str,
			product_id, product_name, type, brand, color, notes, image_url,
			category, price, has_promo, discount
		FROM posts
		WHERE date >= $1 AND user_id IN (` + strings.Join(ph, ",") + `)
	`

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return []domain.Post{}
	}
	defer rows.Close()

	out := []domain.Post{}
	for rows.Next() {
		var p domain.Post
		var price, discount float64
		if err := rows.Scan(
			&p.PostID, &p.UserID, &p.Date, &p.DateStr,
			&p.Product.ProductID, &p.Product.ProductName, &p.Product.Type, &p.Product.Brand, &p.Product.Color, &p.Product.Notes, &p.Product.ImageURL,
			&p.Category, &price, &p.HasPromo, &discount,
		); err != nil {
			return []domain.Post{}
		}
		p.Price = price
		p.Discount = discount
		// calcula o final para UI
		if p.HasPromo {
			p.FinalPrice = math.Round((p.Price*(1-(p.Discount/100)))*100) / 100
		} else {
			p.FinalPrice = p.Price
		}
		out = append(out, p)
	}
	return out
}

func (s *SQLStore) PromoPostsBySeller(sellerID int) []domain.Post {
	rows, err := s.db.Query(`
		SELECT
			id, user_id, date, date_str,
			product_id, product_name, type, brand, color, notes, image_url,
			category, price, has_promo, discount
		FROM posts
		WHERE user_id=$1 AND has_promo=true
	`, sellerID)
	if err != nil {
		return []domain.Post{}
	}
	defer rows.Close()

	out := []domain.Post{}
	for rows.Next() {
		var p domain.Post
		var price, discount float64
		if err := rows.Scan(
			&p.PostID, &p.UserID, &p.Date, &p.DateStr,
			&p.Product.ProductID, &p.Product.ProductName, &p.Product.Type, &p.Product.Brand, &p.Product.Color, &p.Product.Notes, &p.Product.ImageURL,
			&p.Category, &price, &p.HasPromo, &discount,
		); err != nil {
			return []domain.Post{}
		}
		p.Price = price
		p.Discount = discount
		p.FinalPrice = math.Round((p.Price*(1-(p.Discount/100)))*100) / 100
		out = append(out, p)
	}
	return out
}

func (s *SQLStore) ListUsers(order string) ([]domain.User, error) {
	ord := "ASC"
	if order == domain.NameDesc {
		ord = "DESC"
	}
	rows, err := s.db.Query(`SELECT id, name, is_seller FROM users ORDER BY LOWER(name) ` + ord)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Name, &u.IsSeller); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *SQLStore) CreateUser(name string, isSeller bool) (domain.User, error) {
	// gera ID sequencial (tabela original usa id INT PK)
	var id int
	err := s.db.QueryRow(`SELECT COALESCE(MAX(id), 0) + 1 FROM users`).Scan(&id)
	if err != nil {
		return domain.User{}, err
	}
	_, err = s.db.Exec(`INSERT INTO users (id, name, is_seller) VALUES ($1,$2,$3)`, id, name, isSeller)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{ID: id, Name: name, IsSeller: isSeller}, nil
}

func (s *SQLStore) GetAccount(id int) (domain.Account, bool) {
	var a domain.Account
	var avatar sql.NullString
	var createdAt sql.NullTime
	err := s.db.QueryRow(`
		SELECT id, name, email, is_seller, avatar_url, created_at, password_hash
		FROM users
		WHERE id=$1
	`, id).Scan(&a.ID, &a.Name, &a.Email, &a.IsSeller, &avatar, &createdAt, &a.PasswordHash)
	if err != nil {
		return domain.Account{}, false
	}
	if avatar.Valid {
		a.AvatarURL = avatar.String
	}
	if createdAt.Valid {
		a.CreatedAt = createdAt.Time
	}
	return a, true
}

func (s *SQLStore) GetAccountByEmail(email string) (domain.Account, bool) {
	var a domain.Account
	var avatar sql.NullString
	var createdAt sql.NullTime
	err := s.db.QueryRow(`
		SELECT id, name, email, is_seller, avatar_url, created_at, password_hash
		FROM users
		WHERE LOWER(email)=LOWER($1)
	`, email).Scan(&a.ID, &a.Name, &a.Email, &a.IsSeller, &avatar, &createdAt, &a.PasswordHash)
	if err != nil {
		return domain.Account{}, false
	}
	if avatar.Valid {
		a.AvatarURL = avatar.String
	}
	if createdAt.Valid {
		a.CreatedAt = createdAt.Time
	}
	return a, true
}

func (s *SQLStore) CreateAccount(name, email, passwordHash string, isSeller bool) (domain.Account, error) {
	if _, ok := s.GetAccountByEmail(email); ok {
		return domain.Account{}, ErrEmailTaken
	}

	var id int
	err := s.db.QueryRow(`SELECT COALESCE(MAX(id), 0) + 1 FROM users`).Scan(&id)
	if err != nil {
		return domain.Account{}, err
	}

	_, err = s.db.Exec(`
		INSERT INTO users (id, name, email, password_hash, is_seller, created_at)
		VALUES ($1,$2,$3,$4,$5,NOW())
	`, id, name, email, passwordHash, isSeller)
	if err != nil {
		// conflito por unique index
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return domain.Account{}, ErrEmailTaken
		}
		return domain.Account{}, err
	}

	return domain.Account{ID: id, Name: name, Email: email, IsSeller: isSeller, CreatedAt: time.Now()}, nil
}

func (s *SQLStore) UpdateAvatar(userID int, avatarURL string) (domain.Account, error) {
	_, err := s.db.Exec(`UPDATE users SET avatar_url=$1 WHERE id=$2`, avatarURL, userID)
	if err != nil {
		return domain.Account{}, err
	}
	a, ok := s.GetAccount(userID)
	if !ok {
		return domain.Account{}, ErrAccountNotFound
	}
	return a, nil
}

func (s *SQLStore) PostsByUser(userID int) []domain.Post {
	rows, err := s.db.Query(`
		SELECT
			id, user_id, date, date_str,
			product_id, product_name, type, brand, color, notes, image_url,
			category, price, has_promo, discount
		FROM posts
		WHERE user_id=$1
	`, userID)
	if err != nil {
		return []domain.Post{}
	}
	defer rows.Close()

	out := []domain.Post{}
	for rows.Next() {
		var p domain.Post
		var price, discount float64
		if err := rows.Scan(
			&p.PostID, &p.UserID, &p.Date, &p.DateStr,
			&p.Product.ProductID, &p.Product.ProductName, &p.Product.Type, &p.Product.Brand, &p.Product.Color, &p.Product.Notes, &p.Product.ImageURL,
			&p.Category, &price, &p.HasPromo, &discount,
		); err != nil {
			return []domain.Post{}
		}
		p.Price = price
		p.Discount = discount
		if p.HasPromo {
			p.FinalPrice = math.Round((p.Price*(1-(p.Discount/100)))*100) / 100
		} else {
			p.FinalPrice = p.Price
		}
		out = append(out, p)
	}
	return out
}
