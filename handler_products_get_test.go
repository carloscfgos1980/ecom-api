package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/carloscfgos1980/ecom-api/internal/database"
)

func TestHandlerProductsGet_Success(t *testing.T) {
	now := time.Now()
	mock := &mockDB{
		getProductsFn: func(_ context.Context) ([]database.Product, error) {
			return []database.Product{
				{ID: 1, Name: "Widget", Price: "9.99", Quantity: 10, Description: "A widget", CreatedAt: now, UpdatedAt: now},
				{ID: 2, Name: "Gadget", Price: "19.99", Quantity: 5, Description: "A gadget", CreatedAt: now, UpdatedAt: now},
			}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	w := httptest.NewRecorder()

	newTestConfig(mock).handlerProductsGet(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp []ProductResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("couldn't decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 products, got %d", len(resp))
	}
}

func TestHandlerProductsGet_DBError(t *testing.T) {
	mock := &mockDB{
		getProductsFn: func(_ context.Context) ([]database.Product, error) {
			return nil, errors.New("db error")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	w := httptest.NewRecorder()

	newTestConfig(mock).handlerProductsGet(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestHandlerProductsGetByID_Success(t *testing.T) {
	now := time.Now()
	mock := &mockDB{
		getProductByIDFn: func(_ context.Context, id int64) (database.Product, error) {
			return database.Product{
				ID:          id,
				Name:        "Widget",
				Price:       "9.99",
				Quantity:    10,
				Description: "A widget",
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
	req.SetPathValue("productID", "1")
	w := httptest.NewRecorder()

	newTestConfig(mock).handlerProductsGetByID(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var resp ProductResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("couldn't decode response: %v", err)
	}
	if resp.ID != 1 {
		t.Errorf("expected product ID 1, got %d", resp.ID)
	}
}

func TestHandlerProductsGetByID_InvalidID(t *testing.T) {
	mock := &mockDB{}

	req := httptest.NewRequest(http.MethodGet, "/products/abc", nil)
	req.SetPathValue("productID", "abc")
	w := httptest.NewRecorder()

	newTestConfig(mock).handlerProductsGetByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerProductsGetByID_NotFound(t *testing.T) {
	mock := &mockDB{
		getProductByIDFn: func(_ context.Context, _ int64) (database.Product, error) {
			return database.Product{}, errors.New("not found")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/products/99", nil)
	req.SetPathValue("productID", "99")
	w := httptest.NewRecorder()

	newTestConfig(mock).handlerProductsGetByID(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
