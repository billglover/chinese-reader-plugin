package home

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type Request struct {
	Text  string `json:"text"`
	Token string `json:"token"`
}

func handleRequest(rw http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	log.Infof(ctx, "/api request received")

	// TODO:
	// - validate user token
	// - retrieve user's word list
	// - scan the file
	// - return the results

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	defer r.Body.Close()

	var mreq Request
	err = json.Unmarshal(body, &mreq)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	validateToken(ctx, mreq.Token)
}

func init() {
	http.HandleFunc("/api", handleRequest)
}

func validateToken(ctx context.Context, token string) (bool, error) {

	svcName := "token"
	tokenURL, err := appengine.ModuleHostname(ctx, svcName, "", "")
	if err != nil {
		return false, fmt.Errorf("unable to find service %s", svcName)
	}

	scheme := "https"
	if appengine.IsDevAppServer() {
		scheme = "http"
	}

	req, _ := http.NewRequest("GET", scheme+"://"+tokenURL+"/token/"+token, nil)

	client := urlfetch.Client(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("unable to query internal service")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("unable to read response from internal service")
	}

	fmt.Println(string(body))

	return false, nil
}
