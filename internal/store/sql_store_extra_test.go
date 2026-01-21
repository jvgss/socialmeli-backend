package store

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSQLStore_DeletePost_Success(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM posts WHERE id=$1 AND user_id=$2`)).
		WithArgs(123, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := s.DeletePost(1, 123)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_DeletePost_NotFound(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM posts WHERE id=$1 AND user_id=$2`)).
		WithArgs(999, 1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id FROM posts WHERE id=$1`)).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	err := s.DeletePost(1, 999)
	if err != ErrPostNotFound {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_DeletePost_Forbidden(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM posts WHERE id=$1 AND user_id=$2`)).
		WithArgs(123, 1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id FROM posts WHERE id=$1`)).
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(2))

	err := s.DeletePost(1, 123)
	if err != ErrPostForbidden {
		t.Fatalf("expected ErrPostForbidden, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_ListUsers_Success(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name", "is_seller"}).
		AddRow(1, "Alice", false).
		AddRow(2, "Bob", true)

	mock.ExpectQuery(`(?s)SELECT.*FROM users.*ORDER BY`).
		WillReturnRows(rows)

	users, err := s.ListUsers("name_asc")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_CreateUser_Success(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(MAX(id), 0) + 1 FROM users`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users (id, name, is_seller) VALUES ($1,$2,$3)`)).
		WithArgs(10, "New User", true).
		WillReturnResult(sqlmock.NewResult(0, 1))

	u, err := s.CreateUser("New User", true)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if u.ID != 10 || u.Name != "New User" || !u.IsSeller {
		t.Fatalf("unexpected user: %+v", u)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_GetAccount_Success(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name", "email", "is_seller", "avatar_url", "created_at", "password_hash"}).
		AddRow(1, "User", "user@example.com", false, nil, time.Now(), "hash123")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, email, is_seller, avatar_url, created_at, password_hash FROM users WHERE id=$1`)).
		WithArgs(1).
		WillReturnRows(rows)

	acc, ok := s.GetAccount(1)
	if !ok {
		t.Fatalf("expected account to exist")
	}
	if acc.Email != "user@example.com" {
		t.Fatalf("unexpected account: %+v", acc)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_GetAccountByEmail_Success(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name", "email", "is_seller", "avatar_url", "created_at", "password_hash"}).
		AddRow(1, "User", "user@example.com", false, nil, time.Now(), "hash123")

	mock.ExpectQuery(`(?s)SELECT.*FROM users.*WHERE LOWER\(email\)=LOWER\(\$1\)`).
		WithArgs("user@example.com").
		WillReturnRows(rows)

	acc, ok := s.GetAccountByEmail("user@example.com")
	if !ok {
		t.Fatalf("expected account to exist")
	}
	if acc.ID != 1 {
		t.Fatalf("unexpected account: %+v", acc)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_CreateAccount_Success(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	// GetAccountByEmail retorna false (não existe)
	mock.ExpectQuery(`(?s)SELECT.*FROM users.*WHERE LOWER\(email\)=LOWER\(\$1\)`).
		WithArgs("user@example.com").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(MAX(id), 0) + 1 FROM users`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users (id, name, email, password_hash, is_seller, created_at) VALUES ($1,$2,$3,$4,$5,NOW())`)).
		WithArgs(10, "User", "user@example.com", "hash123", true).
		WillReturnResult(sqlmock.NewResult(0, 1))

	acc, err := s.CreateAccount("User", "user@example.com", "hash123", true)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if acc.ID != 10 || acc.Email != "user@example.com" {
		t.Fatalf("unexpected account: %+v", acc)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_UpdateAvatar_Success(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET avatar_url=$1 WHERE id=$2`)).
		WithArgs("/static/avatars/1.jpg", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rows := sqlmock.NewRows([]string{"id", "name", "email", "is_seller", "avatar_url", "created_at", "password_hash"}).
		AddRow(1, "User", "user@example.com", false, "/static/avatars/1.jpg", time.Now(), "hash123")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, email, is_seller, avatar_url, created_at, password_hash FROM users WHERE id=$1`)).
		WithArgs(1).
		WillReturnRows(rows)

	acc, err := s.UpdateAvatar(1, "/static/avatars/1.jpg")
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if acc.AvatarURL != "/static/avatars/1.jpg" {
		t.Fatalf("unexpected avatar: %s", acc.AvatarURL)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestSQLStore_PostsByUser_Success(t *testing.T) {
	s, mock, cleanup := newSQLStoreWithMock(t)
	defer cleanup()

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "date", "date_str",
		"product_id", "product_name", "type", "brand", "color", "notes", "image_url",
		"category", "price", "has_promo", "discount",
	}).AddRow(
		1, 1, now, "01-01-2026",
		10, "Product", "type", "brand", "color", "notes", "",
		1, 100.0, false, 0.0,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, date, date_str, product_id, product_name, type, brand, color, notes, image_url, category, price, has_promo, discount FROM posts WHERE user_id=$1`)).
		WithArgs(1).
		WillReturnRows(rows)

	posts := s.PostsByUser(1)
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
