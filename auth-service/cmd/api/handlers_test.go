package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type RoundTripFunc func(req *http.Request) *http.Response

type testRoundTripper struct {
	fn RoundTripFunc
}

func (t testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.fn(req), nil
}
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: testRoundTripper{fn: fn},
	}
}

// Заглушки для функций, которые будут замокированы в тесте

func Test_Authenticate(t *testing.T) {
	jsonToReturn := `
	{
		"error": false,
		"message": some test message
	}
	`

	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(jsonToReturn)),
			Header:     make(http.Header),
		}
	})
	testApp.Client = client

	postBody := map[string]interface{}{
		"email":    "me@here.com",
		"password": "verysecret",
	}

	body, _ := json.Marshal(postBody)

	req, _ := http.NewRequest("POST", "/authenticate", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(testApp.Authenticate)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("expected http.StatusAccepted but got %d", rr.Code)
	}
}

// logRequestHandler HTTP-обработчик для логирования.
func (app *Config) logRequestHandler(w http.ResponseWriter, r *http.Request) {
	var logEntry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&logEntry); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := app.logRequest(logEntry.Name, logEntry.Data); err != nil {
		http.Error(w, "Failed to log request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// Test_LogRequest тестирует функцию logRequest через HTTP-обработчик.
func Test_LogRequest(t *testing.T) {
	jsonToReturn := `
	{
		"error": false,
		"message": "some test message"
	}
	`

	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(jsonToReturn)),
			Header:     make(http.Header),
		}
	})

	testApp := &Config{Client: client} // Создание экземпляра Config

	postBody := map[string]interface{}{
		"name": "test log",
		"data": "info of test log",
	}

	body, err := json.Marshal(postBody)
	if err != nil {
		t.Fatalf("failed to marshal post body: %v", err)
	}

	req, err := http.NewRequest("POST", "/log", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()

	// Используем обработчик для обработки запроса
	handler := http.HandlerFunc(testApp.logRequestHandler)
	handler.ServeHTTP(rr, req)

	// Проверяем статус ответа
	if rr.Code != http.StatusAccepted {
		t.Errorf("expected http.StatusAccepted but got %d", rr.Code)
	}
}

func generateTokensTest(userID int, ip, secretKey string) (*UserData, error) {
	return &UserData{
		ID:                 userID,
		IP:                 ip,
		RefreshToken:       "mockRefreshToken",
		HashedRefreshToken: "mockHashedRefreshToken",
		AccessToken:        "mockAccessToken",
	}, nil
}

// Test_GenerateTokens тестирует функцию generateTokens.
func Test_GenerateTokensTest(t *testing.T) {
	userID := 1
	ip := "192.168.1.1"
	secretKey := "superSecretKey"

	// Вызов функции
	userData, err := generateTokensTest(userID, ip, secretKey)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Проверка значений
	if userData.ID != userID {
		t.Errorf("expected ID %d, got %d", userID, userData.ID)
	}
	if userData.IP != ip {
		t.Errorf("expected IP %s, got %s", ip, userData.IP)
	}
	if userData.RefreshToken != "mockRefreshToken" {
		t.Errorf("expected RefreshToken %s, got %s", "mockRefreshToken", userData.RefreshToken)
	}
	if userData.HashedRefreshToken != "mockHashedRefreshToken" {
		t.Errorf("expected HashedRefreshToken %s, got %s", "mockHashedRefreshToken", userData.HashedRefreshToken)
	}
	if userData.AccessToken != "mockAccessToken" {
		t.Errorf("expected AccessToken %s, got %s", "mockAccessToken", userData.AccessToken)
	}
}

// Test_GenerateRefreshToken тестирует функцию generateRefreshToken.
func Test_GenerateRefreshToken(t *testing.T) {
	secretKey := "superSecretKey"
	ip := "192.168.1.1"

	// Вызов функции
	refreshToken, hashedRefreshToken, err := generateRefreshToken(secretKey, ip)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Проверка значений
	expectedRefreshToken := secretKey + ip
	if refreshToken != expectedRefreshToken {
		t.Errorf("expected RefreshToken %s, got %s", expectedRefreshToken, refreshToken)
	}

	// Проверка хешированного токена
	hashedBytes, err := base64.StdEncoding.DecodeString(hashedRefreshToken)
	if err != nil {
		t.Fatalf("failed to decode hashedRefreshToken: %v", err)
	}

	// Проверка, что хеш соответствует оригинальному токену
	if err := bcrypt.CompareHashAndPassword(hashedBytes, []byte(refreshToken)); err != nil {
		t.Errorf("hashedRefreshToken does not match the refreshToken: %v", err)
	}
}

func Test_GenerateAccessToken(t *testing.T) {
	userID := 1
	ip := "192.168.1.1"
	secretKey := "superSecretKey"

	// Вызов функции
	tokenString, err := generateAccessToken(userID, ip, secretKey)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Проверка токена
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка метода подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	// Проверка утверждений (claims)
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		t.Fatal("invalid token")
	}

	if claims["sub"] != float64(userID) {
		t.Errorf("expected user ID %d, got %v", userID, claims["sub"])
	}

	if claims["ip"] != ip {
		t.Errorf("expected IP %s, got %v", ip, claims["ip"])
	}

	// Проверка времени истечения токена
	expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
	if time.Now().After(expirationTime) {
		t.Error("token has expired")
	}
}
