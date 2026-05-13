package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/carloscfgos1980/ecom-api/internal/auth"
	"github.com/carloscfgos1980/ecom-api/internal/database"
	"github.com/google/uuid"
)

func TestHandlerLogin_Success(t *testing.T) {
	customerID := uuid.New()
	now := time.Now()
	password := "Password1!"
	hashed, _ := auth.HashPassword(password)

	mock := &mockDB{
		getCustomerByEmailFn: func(_ context.Context, _ string) (database.Customer, error) {
			return database.Customer{
				ID:        customerID,
				CreatedAt: now,
				UpdatedAt: now,
				Email:     "test@example.com",
				Password:  hashed,
			}, nil
		},
	}

	body, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": password,
	})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	newTestConfig(mock).handlerLogin(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("couldn't decode response: %v", err)
	}
	if resp["token"] == "" || resp["token"] == nil {
		t.Error("expected a token in the response")
	}
}

func TestHandlerLogin_EmailNotFound(t *testing.T) {
	mock := &mockDB{
		getCustomerByEmailFn: func(_ context.Context, _ string) (database.Customer, error) {
			return database.Customer{}, errors.New("not found")
		},
	}

	body, _ := json.Marshal(map[string]string{
		"email":    "ghost@example.com",
		"password": "Password1!",
	})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	newTestConfig(mock).handlerLogin(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestHandlerLogin_WrongPassword(t *testing.T) {
	customerID := uuid.New()
	now := time.Now()
	hashed, _ := auth.HashPassword("CorrectPassword1!")

	mock := &mockDB{
		getCustomerByEmailFn: func(_ context.Context, _ string) (database.Customer, error) {
			return database.Customer{
				ID:        customerID,
				CreatedAt: now,
				UpdatedAt: now,
				Email:     "test@example.com",
				Password:  hashed,
			}, nil
		},
	}

	body, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "WrongPassword1!",
	})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	newTestConfig(mock).handlerLogin(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
