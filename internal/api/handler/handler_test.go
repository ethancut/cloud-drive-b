package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ethannself/cloud-drive-b/internal/api/handler"
	"github.com/ethannself/cloud-drive-b/internal/auth"
	"github.com/ethannself/cloud-drive-b/internal/storage"
)

const testRegKey = "test-registration-key"

var testHandler *handler.Handler

func TestMain(m *testing.M) {
	os.Setenv("DATABASE_URL", "postgres://ethan:1@localhost:5432/cloud_drive_v2_test?sslmode=disable")
	os.Setenv("JWT_SECRET", "test-jwt-key")
	os.Setenv("REGISTRATION_KEY", testRegKey)
	storage.InitDataStore()

	ts, err := auth.NewTokenService(os.Getenv("JWT_SECRET"), 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		panic("Failed to created Token Service: " + err.Error())
	}
	testHandler = handler.NewHandler(ts)
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
		"username":         "alice",
		"password":         "secret",
		"email":            "alice@example.com",
		"registration_key": testRegKey,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	testHandler.RegisterHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	err := json.NewDecoder(rr.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %q", resp["status"])
	}
	if resp["access_token"] == "" || resp["refresh_token"] == "" {
		t.Error("expected non-empty tokens")
	}
	if resp["username"] != "alice" {
		t.Errorf("expected username alice, got %q", resp["username"])
	}
}

func TestRegisterHandler_BadJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader([]byte("not-json")))
	rr := httptest.NewRecorder()

	testHandler.RegisterHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestRegisterHandler_MissingFields(t *testing.T) {
	cases := []map[string]string{
		{"username": "alice", "email": "alice@example.com", "registration_key": testRegKey},
		{"username": "alice", "password": "secret", "registration_key": testRegKey},
		{"password": "secret", "email": "alice@example.com", "registration_key": testRegKey},
		{"registration_key": testRegKey}, {},
	}
	for _, c := range cases {
		body, _ := json.Marshal(c)
		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		testHandler.RegisterHandler(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("fields %v: expected 400, got %d", c, rr.Code)
		}
	}
}

func TestRegisterHandler_DuplicateUser(t *testing.T) {
	t.Cleanup(func() { cleanup(t, "duplicate@example.com") })

	body, _ := json.Marshal(map[string]string{
		"username":         "dup",
		"password":         "secret",
		"email":            "duplicate@example.com",
		"registration_key": testRegKey,
	})

	// first registration
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	testHandler.RegisterHandler(httptest.NewRecorder(), req)

	// second should fail
	req = httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	testHandler.RegisterHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

// --- LoginHandler ---

func TestLoginHandler_Success(t *testing.T) {
	t.Cleanup(func() { cleanup(t, "bob@example.com") })

	// seed user
	regBody, _ := json.Marshal(map[string]string{
		"username": "bob", "password": "pass123", "email": "bob@example.com", "registration_key": testRegKey,
	})
	testHandler.RegisterHandler(httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(regBody)))

	body, _ := json.Marshal(map[string]string{
		"email": "bob@example.com", "password": "pass123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	testHandler.LoginHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp["access_token"] == "" {
		t.Errorf("expected non-empty access_token, got: %v", resp)
	}
	if resp["refresh_token"] == "" {
		t.Errorf("expected non-empty refresh_token, got: %v", resp)
	}
	if resp["status"] != "logged_in" {
		t.Errorf("expected logged_in, got %q", resp["status"])
	}
}

func TestLoginHandler_BadJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader([]byte("not-json")))
	rr := httptest.NewRecorder()

	testHandler.LoginHandler(rr, req)

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

		testHandler.LoginHandler(rr, req)

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

	testHandler.LoginHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}
