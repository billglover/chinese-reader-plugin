package token

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateToken(t *testing.T) {
	req, _ := http.NewRequest("POST", "/token", nil)
	response := executeRequest(req)

	if http.StatusCreated != response.Code {
		t.Errorf("unexpected response code: want %d, got %d\n", http.StatusCreated, response.Code)
	}

	var tok Token
	json.Unmarshal(response.Body.Bytes(), &tok)
	if tok.ID == "" {
		t.Errorf("unexpected token length: got '%s'\n", tok.ID)
	}

	if tok.Created.IsZero() {
		t.Errorf("unexpected zero value for created date: '%s'\n", tok.Created)
	}

	if 1000 != tok.Remaining {
		t.Errorf("unexpected remaining value: want %d, got %d\n", 1000, tok.Remaining)
	}
}

func TestFetchToken(t *testing.T) {

	// create a token for use in testing
	req, _ := http.NewRequest("POST", "/token", nil)
	response := executeRequest(req)

	var tok1 Token
	json.Unmarshal(response.Body.Bytes(), &tok1)

	// fetch the token we have just created
	req, _ = http.NewRequest("GET", "/token/"+tok1.ID, nil)
	response = executeRequest(req)
	if http.StatusOK != response.Code {
		t.Errorf("unexpected response code: want %d, got %d\n", http.StatusOK, response.Code)
	}

	var tok2 Token
	json.Unmarshal(response.Body.Bytes(), &tok2)
	if tok1.ID != tok2.ID {
		t.Errorf("unexpected token: want '%s', got '%s'\n", tok1.ID, tok2.ID)
	}
}

// ExecuteRequest is a helper function that executes a test against an httptest.ResponseRecorder
// allowing us to capture the responses for use in testing.
func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	r := register()
	r.ServeHTTP(rr, req)
	return rr
}
