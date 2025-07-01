package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"strings"
	"sync"
	"time"

	"github.com/ayutaz/orochi/internal/errors"
	"golang.org/x/crypto/bcrypt"
)

// TokenStore manages authentication tokens.
type TokenStore struct {
	mu     sync.RWMutex
	tokens map[string]*Token
}

// Token represents an authentication token.
type Token struct {
	Value     string
	Username  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// User represents a user with credentials.
type User struct {
	Username     string
	PasswordHash string
}

// Manager handles authentication.
type Manager struct {
	mu         sync.RWMutex
	users      map[string]*User
	tokenStore *TokenStore
	tokenTTL   time.Duration
}

// NewManager creates a new authentication manager.
func NewManager() *Manager {
	return &Manager{
		users:      make(map[string]*User),
		tokenStore: &TokenStore{tokens: make(map[string]*Token)},
		tokenTTL:   24 * time.Hour, // Tokens expire after 24 hours
	}
}

// CreateUser creates a new user with the given credentials.
func (m *Manager) CreateUser(username, password string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[username]; exists {
		return errors.AlreadyExistsf("user %s already exists", username)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.InternalErrorf("failed to hash password: %v", err)
	}

	m.users[username] = &User{
		Username:     username,
		PasswordHash: string(hash),
	}

	return nil
}

// Authenticate verifies username and password and returns a token.
func (m *Manager) Authenticate(username, password string) (string, error) {
	m.mu.RLock()
	user, exists := m.users[username]
	m.mu.RUnlock()

	if !exists {
		return "", errors.AuthenticationFailedf("invalid username or password")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.AuthenticationFailedf("invalid username or password")
	}

	// Generate token
	tokenValue, err := generateToken()
	if err != nil {
		return "", err
	}

	token := &Token{
		Value:     tokenValue,
		Username:  username,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(m.tokenTTL),
	}

	m.tokenStore.mu.Lock()
	m.tokenStore.tokens[tokenValue] = token
	m.tokenStore.mu.Unlock()

	// Clean up expired tokens
	go m.cleanupExpiredTokens()

	return tokenValue, nil
}

// ValidateToken checks if a token is valid and returns the username.
func (m *Manager) ValidateToken(tokenValue string) (string, error) {
	m.tokenStore.mu.RLock()
	token, exists := m.tokenStore.tokens[tokenValue]
	m.tokenStore.mu.RUnlock()

	if !exists {
		return "", errors.AuthenticationFailedf("invalid token")
	}

	if time.Now().After(token.ExpiresAt) {
		// Token expired, remove it
		m.tokenStore.mu.Lock()
		delete(m.tokenStore.tokens, tokenValue)
		m.tokenStore.mu.Unlock()
		return "", errors.AuthenticationFailedf("token expired")
	}

	return token.Username, nil
}

// Logout invalidates a token.
func (m *Manager) Logout(tokenValue string) {
	m.tokenStore.mu.Lock()
	delete(m.tokenStore.tokens, tokenValue)
	m.tokenStore.mu.Unlock()
}

// cleanupExpiredTokens removes expired tokens from the store.
func (m *Manager) cleanupExpiredTokens() {
	m.tokenStore.mu.Lock()
	defer m.tokenStore.mu.Unlock()

	now := time.Now()
	for value, token := range m.tokenStore.tokens {
		if now.After(token.ExpiresAt) {
			delete(m.tokenStore.tokens, value)
		}
	}
}

// generateToken generates a secure random token.
func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", errors.InternalErrorf("failed to generate token: %v", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// ParseBasicAuth extracts username and password from Basic Auth header.
func ParseBasicAuth(authHeader string) (username, password string, ok bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(authHeader, prefix) {
		return
	}

	encoded := authHeader[len(prefix):]
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return
	}

	return parts[0], parts[1], true
}

// ParseBearerToken extracts token from Bearer Auth header.
func ParseBearerToken(authHeader string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", false
	}
	return authHeader[len(prefix):], true
}

// ConstantTimeCompare performs a constant-time comparison of two strings.
func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
