package store

import "socialmeli/internal/domain"

func SeedDefault(st *MemoryStore) {
	st.SeedUsers([]domain.User{
		{ID: 123, Name: "usuario123", IsSeller: false},
		{ID: 234, Name: "vendedor1", IsSeller: true},
		{ID: 6932, Name: "vendedor2", IsSeller: true},
		{ID: 4698, Name: "usuario1", IsSeller: false},
	})
}
