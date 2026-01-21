package domain

import (
	"errors"
	"regexp"
)

var (
	ErrIDEmpty           = errors.New("O id não pode estar vazio.")
	ErrIDGreaterThanZero = errors.New("id deve ser maior que zero")
	ErrDateEmpty         = errors.New("A data não pode estar vazia.")
	ErrFieldEmpty        = errors.New("O campo não pode estar vazio.")
	ErrMaxLen40          = errors.New("O comprimento não pode exceder 40 caracteres.")
	ErrMaxLen15          = errors.New("O comprimento não pode exceder 15 caracteres.")
	ErrMaxLen25          = errors.New("O comprimento não pode exceder 25 caracteres.")
	ErrMaxLen80          = errors.New("O comprimento não pode exceder 80 caracteres.")
	ErrSpecialChars      = errors.New("O campo não pode conter caracteres especiais.")
	ErrCategoryEmpty     = errors.New("O campo não pode estar vazio.")
	ErrPriceEmpty        = errors.New("O campo não pode estar vazio.")
	ErrPriceMax          = errors.New("O preço máximo por produto é de 10.000.000")
)

// permite letras (inclui acentos), números e espaço. bloqueia %, &, $, etc.
var allowedText = regexp.MustCompile(`^[A-Za-zÀ-ÿ0-9 ]+$`)

func ValidateID(id int) error {
	if id == 0 {
		return ErrIDEmpty
	}
	if id < 0 {
		return ErrIDGreaterThanZero
	}
	return nil
}

func ValidateTextRequired(s string, max int, maxErr error) error {
	if s == "" {
		return ErrFieldEmpty
	}
	if len([]rune(s)) > max {
		return maxErr
	}
	if !allowedText.MatchString(s) {
		return ErrSpecialChars
	}
	return nil
}

func ValidateNotesOptional(s string) error {
	if s == "" {
		return nil
	}
	if len([]rune(s)) > 80 {
		return ErrMaxLen80
	}
	if !allowedText.MatchString(s) {
		return ErrSpecialChars
	}
	return nil
}

func ValidatePrice(price float64) error {
	if price == 0 {
		return ErrPriceEmpty
	}
	if price > 10_000_000 {
		return ErrPriceMax
	}
	return nil
}
