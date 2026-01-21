package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
)

type tokenClaims struct {
	Sub int   `json:"sub"`
	Exp int64 `json:"exp"`
	Iat int64 `json:"iat"`
}

func tokenSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret"
	}
	return []byte(secret)
}

func b64(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func unb64(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func sign(data string) string {
	mac := hmac.New(sha256.New, tokenSecret())
	mac.Write([]byte(data))
	return b64(mac.Sum(nil))
}

func MakeToken(userID int, ttl time.Duration) (string, error) {
	header := b64([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claims := tokenClaims{Sub: userID, Iat: time.Now().Unix(), Exp: time.Now().Add(ttl).Unix()}
	pb, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := b64(pb)
	unsigned := header + "." + payload
	sig := sign(unsigned)
	return unsigned + "." + sig, nil
}

func ParseToken(token string) (tokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return tokenClaims{}, errors.New("token inválido")
	}
	unsigned := parts[0] + "." + parts[1]
	if sign(unsigned) != parts[2] {
		return tokenClaims{}, errors.New("token inválido")
	}
	pb, err := unb64(parts[1])
	if err != nil {
		return tokenClaims{}, errors.New("token inválido")
	}
	var c tokenClaims
	if err := json.Unmarshal(pb, &c); err != nil {
		return tokenClaims{}, errors.New("token inválido")
	}
	if time.Now().Unix() > c.Exp {
		return tokenClaims{}, errors.New("token expirado")
	}
	return c, nil
}
