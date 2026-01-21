package store

import "testing"

func TestSeedDefault(t *testing.T) {
	st := NewMemoryStore()
	SeedDefault(st)

	// garante que seed colocou os usuários esperados
	if _, ok := st.GetUser(123); !ok {
		t.Fatalf("expected SeedDefault to insert user 123")
	}
	if _, ok := st.GetUser(234); !ok {
		t.Fatalf("expected SeedDefault to insert user 234")
	}
	if _, ok := st.GetUser(6932); !ok {
		t.Fatalf("expected SeedDefault to insert user 6932")
	}
	if _, ok := st.GetUser(4698); !ok {
		t.Fatalf("expected SeedDefault to insert user 4698")
	}
}
