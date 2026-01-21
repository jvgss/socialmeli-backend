package domain

import (
	"errors"
	"sort"
	"strings"
)

var (
	ErrInvalidOrder = errors.New("Tipo de ordenação inexistente.")
)

const (
	NameAsc  = "name_asc"
	NameDesc = "name_desc"
	DateAsc  = "date_asc"
	DateDesc = "date_desc"
)

func ValidateOrderForUsers(order string) error {
	if order == "" {
		return nil
	}
	switch strings.ToLower(order) {
	case NameAsc, NameDesc:
		return nil
	default:
		return ErrInvalidOrder
	}
}

func ValidateOrderForPosts(order string) error {
	if order == "" {
		return nil
	}
	switch strings.ToLower(order) {
	case DateAsc, DateDesc:
		return nil
	default:
		return ErrInvalidOrder
	}
}

func SortUsersByName(users []User, order string) {
	switch strings.ToLower(order) {
	case NameAsc:
		sort.Slice(users, func(i, j int) bool { return users[i].Name < users[j].Name })
	case NameDesc:
		sort.Slice(users, func(i, j int) bool { return users[i].Name > users[j].Name })
	}
}

func SortPostsByDate(posts []Post, order string) {
	switch strings.ToLower(order) {
	case DateAsc:
		sort.Slice(posts, func(i, j int) bool { return posts[i].Date.Before(posts[j].Date) })
	case DateDesc, "":
		// padrão: mais recentes primeiro
		sort.Slice(posts, func(i, j int) bool { return posts[i].Date.After(posts[j].Date) })
	}
}
