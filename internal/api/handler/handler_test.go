package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ethannself/cloud-drive-b/internal/api/handler"
	"github.com/ethannself/cloud-drive-b/internal/storage"
)

func TestMain(m *testing.M) {
	os.Setenv("DATABASE_URL", "postgres://ethan:1@localhost:5432/cloud_drive_test?sslmode=disable")
	storage.InitDataStore()
	os.Exit(m.Run())
}

func cleanup(t *testing.T, email string) {
	t.Helper()
	err := storage.GetDataStore().DeleteUser(email)
	if err != nil {
		t.Logf("cleanup failed for %s: %v", email, err)
	}
}

func TestDefaultHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.DefaultHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "Hello, World!\n" {
		t.Errorf("unexpected body: %q", rr.Body.String())
	}
}

func TestRegisterHandler_Success(t *testing.T) {
	t.Cleanup(func() { cleanup(t, "alice@example.com") })

	body, _ := json.Marshal(map[string]string{
		"username": "alice",
		"password": "secret",
		"email":    "alice@example.com",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.RegisterHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %q", resp["status"])
	}
}

func TestRegisterHandler_BadJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader([]byte("not-json")))
	rr := httptest.NewRecorder()

	handler.RegisterHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestRegisterHandler_MissingFields(t *testing.T) {
	cases := []map[string]string{
		{"username": "alice", "email": "alice@example.com"},
		{"username": "alice", "password": "secret"},
		{"password": "secret", "email": "alice@example.com"},
		{},
	}
	for _, c := range cases {
		body, _ := json.Marshal(c)
		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.RegisterHandler(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("fields %v: expected 400, got %d", c, rr.Code)
		}
	}
}

func TestRegisterHandler_DuplicateUser(t *testing.T) {
	t.Cleanup(func() { cleanup(t, "duplicate@example.com") })

	body, _ := json.Marshal(map[string]string{
		"username": "dup", "password": "secret", "email": "duplicate@example.com",
	})

	// first registration
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	handler.RegisterHandler(httptest.NewRecorder(), req)

	// second should fail
	body, _ = json.Marshal(map[string]string{
		"username": "dup", "password": "secret", "email": "duplicate@example.com",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.RegisterHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

// --- LoginHandler ---

func TestLoginHandler_Success(t *testing.T) {
	t.Cleanup(func() { cleanup(t, "bob@example.com") })

	// seed user
	regBody, _ := json.Marshal(map[string]string{
		"username": "bob", "password": "pass123", "email": "bob@example.com",
	})
	handler.RegisterHandler(httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(regBody)))

	body, _ := json.Marshal(map[string]string{
		"email": "bob@example.com", "password": "pass123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.LoginHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["token"] == "" {
		t.Error("expected a non-empty token")
	}
	if resp["status"] != "logged_in" {
		t.Errorf("expected logged_in, got %q", resp["status"])
	}
}

func TestLoginHandler_BadJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader([]byte("not-json")))
	rr := httptest.NewRecorder()

	handler.LoginHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestLoginHandler_MissingFields(t *testing.T) {
	cases := []map[string]string{
		{"email": "bob@example.com"},
		{"password": "pass123"},
		{},
	}
	for _, c := range cases {
		body, _ := json.Marshal(c)
		req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.LoginHandler(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("fields %v: expected 400, got %d", c, rr.Code)
		}
	}
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	body, _ := json.Marshal(map[string]string{
		"email": "nobody@example.com", "password": "wrongpass",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.LoginHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}
