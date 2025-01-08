package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_WriteLog(t *testing.T) {

	postBody := map[string]interface{}{
		"name": "test_name",
		"data": "test_data",
	}

	body, _ := json.Marshal(postBody)

	req, _ := http.NewRequest("POST", "/log", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(testApp.WriteLog)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("expected http.StatusAccepted but got %d", rr.Code)
	}
}
