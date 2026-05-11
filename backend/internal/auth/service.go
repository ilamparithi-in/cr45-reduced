package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/CR45-NITT/cr45-reduced/backend/internal/store"
)

type tokenPayload struct {
	Username string `json:"username"`
	Exp      int64  `json:"exp"`
}

type Service struct {
	repo      store.Repository
	secretKey []byte
	now       func() time.Time
}

func NewService(repo store.Repository, secret string) (*Service, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, store.ErrMissingSecretKey
	}
	return &Service{repo: repo, secretKey: []byte(secret), now: time.Now}, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (string, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return "", store.ErrInvalidLogin
	}

	storedHash, err := s.repo.GetPasswordHash(ctx, username)
	if err != nil {
		return "", err
	}
	computed := sha256.Sum256([]byte(password))
	computedHex := hex.EncodeToString(computed[:])
	if subtle.ConstantTimeCompare([]byte(computedHex), []byte(strings.TrimSpace(storedHash))) != 1 {
		return "", store.ErrInvalidLogin
	}

	payload := tokenPayload{Username: username, Exp: s.now().Add(12 * time.Hour).Unix()}
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	message := base64.RawURLEncoding.EncodeToString(encodedPayload)
	signature := s.sign(message)
	return message + "." + signature, nil
}

func (s *Service) Validate(token string) error {
	token = strings.TrimSpace(token)
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return store.ErrUnauthorized
	}
	expectedSig := s.sign(parts[0])
	if subtle.ConstantTimeCompare([]byte(expectedSig), []byte(parts[1])) != 1 {
		return store.ErrUnauthorized
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return store.ErrUnauthorized
	}
	var payload tokenPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return store.ErrUnauthorized
	}
	if payload.Username == "" || payload.Exp < s.now().Unix() {
		return store.ErrUnauthorized
	}
	return nil
}

func (s *Service) sign(message string) string {
	h := hmac.New(sha256.New, s.secretKey)
	_, _ = h.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func BearerToken(header string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", store.ErrUnauthorized
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", store.ErrUnauthorized
	}
	return token, nil
}

var ErrInvalidToken = errors.New("invalid token")