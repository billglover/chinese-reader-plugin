package token

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/appengine/aetest"
)

func TestPost(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatalf("unable to create context: %v", err.Error())
	}
	defer done()

	req, err := http.NewRequest("GET", "/v1/do/foo", nil)
	if err != nil {
		t.Fatalf("unable to create request: %v", err)
	}

	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	PostTokenHandler(w, req)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header.Get("Content-Type"))
	fmt.Println(string(body))
}
