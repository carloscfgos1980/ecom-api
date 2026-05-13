package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/carloscfgos1980/ecom-api/internal/database"
	"github.com/google/uuid"
)

func newTestConfig(db DB) *apiConfig {
	return &apiConfig{
		db:        db,
		jwtSecret: "test-secret",
		port:      "8080",
	}
}

func TestHandlerUsersCreate_Success(t *testing.T) {
	customerID := uuid.New()
	now := time.Now()

	mock := &mockDB{
		getCustomerByEmailFn: func(_ context.Context, _ string) (database.Customer, error) {
			return database.Customer{}, sql.ErrNoRows
		},
		createCustomerFn: func(_ context.Context, arg database.CreateCustomerParams) (database.Customer, error) {
			return database.Customer{
				ID:        customerID,
				CreatedAt: now,
				UpdatedAt: now,
				Email:     arg.Email,
				Password:  arg.Password,
			}, nil
		},
	}

	body, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "Password1!",
	})
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	w := httptest.NewRecorder()

	newTestConfig(mock).handlerUsersCreate(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp CustomerResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("couldn't decode response: %v", err)
	}
	if resp.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", resp.Email)
	}
}

func TestHandlerUsersCreate_EmailAlreadyExists(t *testing.T) {
	mock := &mockDB{
		getCustomerByEmailFn: func(_ context.Context, _ string) (database.Customer, error) {
			return database.Customer{Email: "test@example.com"}, nil
		},
	}

	body, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "Password1!",
	})
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	w := httptest.NewRecorder()

	newTestConfig(mock).handlerUsersCreate(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestHandlerUsersCreate_InvalidEmail(t *testing.T) {
	mock := &mockDB{}

	body, _ := json.Marshal(map[string]string{
		"email":    "not-an-email",
		"password": "Password1!",
	})
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	w := httptest.NewRecorder()

	newTestConfig(mock).handlerUsersCreate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerUsersCreate_WeakPassword(t *testing.T) {
	mock := &mockDB{}

	body, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "weak",
	})
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	w := httptest.NewRecorder()

	newTestConfig(mock).handlerUsersCreate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
