package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type RoundTripFunc func(req *http.Request) *http.Response

type testRoundTripper struct {
	fn RoundTripFunc
}

// logRequestHandler HTTP-обработчик для логирования.
func (app *Config) authTestPayloadHandler(w http.ResponseWriter, r *http.Request) {
	var authTestPayload authPayload
	authTestPayload.Email = "admin@example.com"
	authTestPayload.Pass = "123"
}

func (t testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.fn(req), nil
}
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: testRoundTripper{fn: fn},
	}
}

func Test_Authenticate_Success(t *testing.T) {
	// Prepare the request payload
	jsonToReturn := `
	{
		"error": false,
		"message": "test auth, success"
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
		"name": "test auth",
		"data": "info of test auth",
	}

	body, err := json.Marshal(postBody)
	if err != nil {
		t.Fatalf("failed to marshal post body: %v", err)
	}

	req, err := http.NewRequest("POST", "/authenticate", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testApp.authTestPayloadHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var payload jsonResponse
	json.NewDecoder(rr.Body).Decode(&payload)
	if payload.Error {
		t.Errorf("expected no error, got %v", payload.Message)
	}
}

func Test_Authenticate_Unauthorized(t *testing.T) {
	// Prepare the request payload
	jsonToReturn := `
	{
		"error": true,
		"message": "test auth, false"
	}
	`

	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       io.NopCloser(bytes.NewBufferString(jsonToReturn)),
			Header:     make(http.Header),
		}
	})
	testApp := &Config{Client: client} // Создание экземпляра Config

	postBody := map[string]interface{}{
		"name": "test auth",
		"data": "info of test auth",
	}

	body, err := json.Marshal(postBody)
	if err != nil {
		t.Fatalf("failed to marshal post body: %v", err)
	}

	req, err := http.NewRequest("POST", "/authenticate", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	rr, err := testApp.Client.Do(req)
	if err != nil {
		fmt.Println("Error making a request")
	}
	defer rr.Body.Close() // Не забываем закрыть тело ответа после использования

	if status := rr.StatusCode; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}

	var payload jsonResponse
	json.NewDecoder(rr.Body).Decode(&payload)
	if !payload.Error {
		t.Errorf("expected error, got none")
	}
}

func Test_Authenticate_Error(t *testing.T) {
	// Prepare the request payload
	jsonToReturn := `
	{
		"error": true,
		"message": "test auth, false"
	}
	`

	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(bytes.NewBufferString(jsonToReturn)),
			Header:     make(http.Header),
		}
	})
	testApp := &Config{Client: client} // Создание экземпляра Config

	postBody := map[string]interface{}{
		"name": "test auth",
		"data": "info of test auth",
	}

	body, err := json.Marshal(postBody)
	if err != nil {
		t.Fatalf("failed to marshal post body: %v", err)
	}

	req, err := http.NewRequest("POST", "/authenticate", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	rr, err := testApp.Client.Do(req)
	if err != nil {
		fmt.Println("Error making a request")
	}
	defer rr.Body.Close() // Не забываем закрыть тело ответа после использования

	if status := rr.StatusCode; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var payload jsonResponse
	json.NewDecoder(rr.Body).Decode(&payload)
	if !payload.Error {
		t.Errorf("expected error, got none")
	}
}
