package domain

import "testing"

func TestValidateID(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		wantErr error
	}{
		{
			name:    "id zero retorna ErrIDEmpty",
			id:      0,
			wantErr: ErrIDEmpty,
		},
		{
			name:    "id negativo retorna ErrIDGreaterThanZero",
			id:      -1,
			wantErr: ErrIDGreaterThanZero,
		},
		{
			name:    "id positivo válido não retorna erro",
			id:      10,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.id)
			if err != tt.wantErr {
				t.Fatalf("ValidateID(%d) error = %v, want %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTextRequired(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		max     int
		maxErr  error
		wantErr error
	}{
		{
			name:    "campo vazio retorna ErrFieldEmpty",
			s:       "",
			max:     40,
			maxErr:  ErrMaxLen40,
			wantErr: ErrFieldEmpty,
		},
		{
			name:    "tamanho maior que max retorna maxErr",
			s:       "123456",
			max:     5,
			maxErr:  ErrMaxLen40,
			wantErr: ErrMaxLen40,
		},
		{
			name:    "caracteres especiais retornam ErrSpecialChars",
			s:       "abc#",
			max:     40,
			maxErr:  ErrMaxLen40,
			wantErr: ErrSpecialChars,
		},
		{
			name:    "texto válido com acentos não retorna erro",
			s:       "Café com Leite 123",
			max:     40,
			maxErr:  ErrMaxLen40,
			wantErr: nil,
		},
		{
			name:    "tamanho igual ao limite é válido",
			s:       "12345",
			max:     5,
			maxErr:  ErrMaxLen40,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTextRequired(tt.s, tt.max, tt.maxErr)
			if err != tt.wantErr {
				t.Fatalf("ValidateTextRequired(%q, %d) error = %v, want %v", tt.s, tt.max, err, tt.wantErr)
			}
		})
	}
}

func TestValidateNotesOptional(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		wantErr error
	}{
		{
			name:    "campo vazio é válido",
			s:       "",
			wantErr: nil,
		},
		{
			name:    "nota maior que 80 retorna ErrMaxLen80",
			s:       string(make([]rune, 81)),
			wantErr: ErrMaxLen80,
		},
		{
			name:    "nota com caracteres especiais retorna ErrSpecialChars",
			s:       "nota com #",
			wantErr: ErrSpecialChars,
		},
		{
			name:    "nota válida não retorna erro",
			s:       "Observação simples 123",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// caso do comprimento > 80 precisa de fato de 81 runes
			if tt.name == "nota maior que 80 retorna ErrMaxLen80" {
				runes := make([]rune, 81)
				for i := range runes {
					runes[i] = 'a'
				}
				tt.s = string(runes)
			}

			err := ValidateNotesOptional(tt.s)
			if err != tt.wantErr {
				t.Fatalf("ValidateNotesOptional(%q) error = %v, want %v", tt.s, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePrice(t *testing.T) {
	tests := []struct {
		name    string
		price   float64
		wantErr error
	}{
		{
			name:    "preço zero retorna ErrPriceEmpty",
			price:   0,
			wantErr: ErrPriceEmpty,
		},
		{
			name:    "preço maior que 10_000_000 retorna ErrPriceMax",
			price:   10_000_000.01,
			wantErr: ErrPriceMax,
		},
		{
			name:    "preço exatamente no limite é válido",
			price:   10_000_000,
			wantErr: nil,
		},
		{
			name:    "preço positivo dentro do limite é válido",
			price:   123.45,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePrice(tt.price)
			if err != tt.wantErr {
				t.Fatalf("ValidatePrice(%f) error = %v, want %v", tt.price, err, tt.wantErr)
			}
		})
	}
}
