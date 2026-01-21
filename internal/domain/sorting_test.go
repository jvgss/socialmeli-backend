package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"socialmeli/internal/domain"
)

func Test_T0003_OrderAlphabeticalExists(t *testing.T) {
	require.NoError(t, domain.ValidateOrderForUsers("name_asc"))
	require.NoError(t, domain.ValidateOrderForUsers("name_desc"))
	require.Error(t, domain.ValidateOrderForUsers("banana"))
}

func Test_T0004_OrderByNameAscDesc(t *testing.T) {
	users := []domain.User{
		{ID: 1, Name: "Carlos"},
		{ID: 2, Name: "Ana"},
	}
	domain.SortUsersByName(users, "name_asc")
	require.Equal(t, "Ana", users[0].Name)

	domain.SortUsersByName(users, "name_desc")
	require.Equal(t, "Carlos", users[0].Name)
}

func Test_T0005_OrderDateExists(t *testing.T) {
	require.NoError(t, domain.ValidateOrderForPosts("date_asc"))
	require.NoError(t, domain.ValidateOrderForPosts("date_desc"))
	require.Error(t, domain.ValidateOrderForPosts("name_asc"))
}

func Test_T0006_OrderByDate(t *testing.T) {
	a := domain.Post{Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	b := domain.Post{Date: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)}
	posts := []domain.Post{b, a}

	domain.SortPostsByDate(posts, "date_asc")
	require.True(t, posts[0].Date.Before(posts[1].Date))

	domain.SortPostsByDate(posts, "date_desc")
	require.True(t, posts[0].Date.After(posts[1].Date))
}
