package store

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"socialmeli/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
)

func newSQLStoreWithMock(t *testing.T) (*SQLStore, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	s := &SQLStore{db: db}

	cleanup := func() { _ = db.Close() }
	return s, mock, cleanup
}

func TestSQLStore_GetUser_Found(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name", "is_seller"}).
		AddRow(1, "Joao", true)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, is_seller FROM users WHERE id=$1`)).
		WithArgs(1).
		WillReturnRows(rows)

	u, ok := s.GetUser(1)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if u.ID != 1 || u.Name != "Joao" || u.IsSeller != true {
		t.Fatalf("unexpected user: %+v", u)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_GetUser_NotFound(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, is_seller FROM users WHERE id=$1`)).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	_, ok := s.GetUser(999)
	if ok {
		t.Fatalf("expected ok=false")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_Follow_UserNotFound(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	// GetUser(userID) -> no rows
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, is_seller FROM users WHERE id=$1`)).
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	err := s.Follow(1, 2)
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_Follow_Success_ExecInsert(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	// GetUser(userID)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, is_seller FROM users WHERE id=$1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "is_seller"}).AddRow(1, "Buyer", false))

	// GetUser(sellerID)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, is_seller FROM users WHERE id=$1`)).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "is_seller"}).AddRow(2, "Seller", true))

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO follows (user_id, seller_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, seller_id) DO NOTHING
	`)).
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := s.Follow(1, 2); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_Unfollow_Success_ExecDelete(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	// GetUser(userID)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, is_seller FROM users WHERE id=$1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "is_seller"}).AddRow(1, "Buyer", false))

	// GetUser(sellerID)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, is_seller FROM users WHERE id=$1`)).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "is_seller"}).AddRow(2, "Seller", true))

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM follows WHERE user_id=$1 AND seller_id=$2`)).
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := s.Unfollow(1, 2); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_FollowersOf_UserNotFound(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	// GetUser(sellerID) -> no rows
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, is_seller FROM users WHERE id=$1`)).
		WithArgs(2).
		WillReturnError(sql.ErrNoRows)

	_, err := s.FollowersOf(2)
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_FollowersOf_Success(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	// GetUser(sellerID)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, is_seller FROM users WHERE id=$1`)).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "is_seller"}).AddRow(2, "Seller", true))

	rows := sqlmock.NewRows([]string{"id", "name", "is_seller"}).
		AddRow(1, "Buyer1", false).
		AddRow(3, "Buyer2", false)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT u.id, u.name, u.is_seller
		FROM follows f
		JOIN users u ON u.id = f.user_id
		WHERE f.seller_id = $1
	`)).
		WithArgs(2).
		WillReturnRows(rows)

	out, err := s.FollowersOf(2)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2, got %d", len(out))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_AddPost_UserNotFound(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, is_seller FROM users WHERE id=$1`)).
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	_, err := s.AddPost(domain.Post{UserID: 99})
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_AddPost_Success_ReturnsID(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	// GetUser ok
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, is_seller FROM users WHERE id=$1`)).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "is_seller"}).AddRow(2, "Seller", true))

	p := domain.Post{
		UserID:   2,
		Date:     time.Now(),
		DateStr:  "01-01-2026",
		Category: 1,
		Price:    100.0,
		HasPromo: true,
		Discount: 10.0,
		Product: domain.Product{
			ProductID:   55,
			ProductName: "Mouse",
			Type:        "peripheral",
			Brand:       "X",
			Color:       "Black",
			Notes:       "note",
		},
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).
		WithArgs(
			p.UserID, p.Date, p.DateStr,
			p.Product.ProductID, p.Product.ProductName, p.Product.Type, p.Product.Brand, p.Product.Color, p.Product.Notes, p.Product.ImageURL,
			p.Category, p.Price, p.HasPromo, p.Discount,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(123))

	id, err := s.AddPost(p)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if id != 123 {
		t.Fatalf("expected 123, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_PostsFromSellersSince_EmptySellerIDs(t *testing.T) {
	s, _, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	out := s.PostsFromSellersSince([]int{}, time.Now())
	if len(out) != 0 {
		t.Fatalf("expected empty, got %d", len(out))
	}
}

func TestSQLStore_PromoPostsBySeller_Success(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "date", "date_str",
		"product_id", "product_name", "type", "brand", "color", "notes", "image_url",
		"category", "price", "has_promo", "discount",
	}).AddRow(
		1, 2, now, "01-01-2026",
		10, "Mouse", "peripheral", "BrandX", "Black", "note", "/static/products/1.jpg",
		1, 100.0, true, 10.0,
	)

	mock.ExpectQuery(`(?s)SELECT.*FROM posts.*WHERE user_id=\$1 AND has_promo=true`).
		WithArgs(2).
		WillReturnRows(rows)

	out := s.PromoPostsBySeller(2)
	if len(out) != 1 {
		t.Fatalf("expected 1, got %d", len(out))
	}
	if out[0].UserID != 2 || !out[0].HasPromo {
		t.Fatalf("unexpected post: %+v", out[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}

}
func TestSQLStore_FollowedBy_Success(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	// GetUser(userID) ok
	mock.ExpectQuery(`SELECT id, name, is_seller FROM users WHERE id=\$1`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "is_seller"}).AddRow(1, "Buyer", false))

	// Query FollowedBy
	rows := sqlmock.NewRows([]string{"id", "name", "is_seller"}).
		AddRow(2, "SellerA", true).
		AddRow(3, "SellerB", true)

	mock.ExpectQuery(`(?s)SELECT\s+u\.id,\s+u\.name,\s+u\.is_seller.*FROM\s+follows\s+f.*JOIN\s+users\s+u.*WHERE\s+f\.user_id\s*=\s*\$1`).
		WithArgs(1).
		WillReturnRows(rows)

	out, err := s.FollowedBy(1)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2, got %d", len(out))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
func TestSQLStore_PostsFromSellersSince_EmptySellerIDs_ReturnsEmpty(t *testing.T) {
	s, _, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	out := s.PostsFromSellersSince([]int{}, time.Now())
	if len(out) != 0 {
		t.Fatalf("expected empty, got %d", len(out))
	}

}
func TestSQLStore_PostsFromSellersSince_ScanError_ReturnsEmpty(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// row propositalmente com colunas faltando (vai estourar no Scan)
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "date", "date_str",
		// faltam colunas do produto e o resto -> Scan vai falhar
	}).AddRow(1, 2, since, "01-01-2026")

	mock.ExpectQuery(`(?s)FROM\s+posts.*WHERE\s+date\s*>=\s*\$1\s+AND\s+user_id\s+IN\s+\(\$2\)`).
		WithArgs(since, 2).
		WillReturnRows(rows)

	out := s.PostsFromSellersSince([]int{2}, since)
	if len(out) != 0 {
		t.Fatalf("expected empty due to scan error, got %d", len(out))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
