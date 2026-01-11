package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SupabaseAuth handles Supabase authentication
type SupabaseAuth struct {
	url        string
	serviceKey string
	httpClient *http.Client
}

// User represents a Supabase user
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// Profile represents a user profile from the profiles table
type Profile struct {
	ID        string  `json:"id"`
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatar_url"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// NewSupabaseAuth creates a new Supabase auth client
func NewSupabaseAuth(url, serviceKey string) *SupabaseAuth {
	return &SupabaseAuth{
		url:        url,
		serviceKey: serviceKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// VerifyToken verifies a Supabase JWT token and returns the user
func (s *SupabaseAuth) VerifyToken(ctx context.Context, token string) (*User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.url+"/auth/v1/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("apikey", s.serviceKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("invalid token: status %d, body: %s", resp.StatusCode, string(body))
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, nil
}

// GetProfile retrieves a user's profile from the profiles table using PostgREST API
func (s *SupabaseAuth) GetProfile(ctx context.Context, userID string) (*Profile, error) {
	url := fmt.Sprintf("%s/rest/v1/profiles?id=eq.%s&select=*", s.url, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch profile: status %d, body: %s", resp.StatusCode, string(body))
	}

	var profiles []Profile
	if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
		return nil, fmt.Errorf("failed to decode profile: %w", err)
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("profile not found")
	}

	return &profiles[0], nil
}

// UpdateProfile updates a user's profile using PostgREST API
func (s *SupabaseAuth) UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) error {
	url := fmt.Sprintf("%s/rest/v1/profiles?id=eq.%s", s.url, userID)

	body, err := json.Marshal(updates)
	if err != nil {
		return fmt.Errorf("failed to marshal updates: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update profile: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
