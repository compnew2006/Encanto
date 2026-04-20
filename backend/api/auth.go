package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// In a real application, this should be an environment variable.
var jwtSecretKey = []byte("super-secret-encanto-key")

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// User context stored in JWT
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Settings definition
type UserSettings struct {
	Theme         string `json:"theme"`
	Language      string `json:"language"`
	ThemePreset   string `json:"theme_preset"`
	SidebarPinned bool   `json:"sidebar_pinned"`
}

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"` // user's role in this specific org
}

// Rich User Model for Frontend Consumption
type UserResponse struct {
	ID                  string         `json:"id"`
	Email               string         `json:"email"`
	Name                string         `json:"name"`
	Avatar              string         `json:"avatar"`
	Status              string         `json:"status"`
	Role                string         `json:"role"` // active role in current context
	Settings            UserSettings   `json:"settings"`
	Organizations       []Organization `json:"organizations"`
	CurrentOrganization Organization   `json:"current_organization"`
}

const mockEmail = "admin@example.com"
const mockPassword = "password123"

type SwitchOrgRequest struct {
	OrgID string `json:"org_id"`
}

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	user, err := s.store.GetUserByEmail(req.Email, req.Password)
	if err != nil {
		http.Error(w, `{"error":"Invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	activeOrgID := user.CurrentOrganization.ID

	// Refresh org context cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "org_context",
		Value:    activeOrgID,
		Expires:  time.Now().Add(365 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Generate JWT
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		http.Error(w, `{"error":"Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	// Set HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"user":    user,
	})
}

func (s *Server) SwitchOrg(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("user").(*Claims)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req SwitchOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}
	
	if !s.store.OrgAccessible(req.OrgID) {
		http.Error(w, `{"error":"Organization access denied"}`, http.StatusForbidden)
		return
	}
	user := s.store.GetUserResponse(req.OrgID)

	http.SetCookie(w, &http.Cookie{
		Name:     "org_context",
		Value:    user.CurrentOrganization.ID,
		Expires:  time.Now().Add(365 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Organization switched safely",
		"user":    user,
	})
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	// Erase cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Erase org_context
	http.SetCookie(w, &http.Cookie{
		Name:     "org_context",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Logged out"}`))
}

func (s *Server) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*Claims)
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	activeOrgID := ""
	if cookie, err := r.Cookie("org_context"); err == nil {
		activeOrgID = cookie.Value
	}
	user := s.store.GetUserResponse(activeOrgID)
	
	// Replace with DB query based on claims.UserID later
	if user.ID != claims.UserID {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": user,
	})
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
				return
			}
			http.Error(w, `{"error":"Bad request"}`, http.StatusBadRequest)
			return
		}

		tokenStr := cookie.Value
		claims := &Claims{}
		
		tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecretKey, nil
		})

		if err != nil || !tkn.Valid {
			http.Error(w, `{"error":"Unauthorized - Invalid session"}`, http.StatusUnauthorized)
			return
		}

		// Attach user info to context
		ctx := context.WithValue(r.Context(), "user", claims)
		r = r.WithContext(ctx)
		
		next.ServeHTTP(w, r)
	})
}
