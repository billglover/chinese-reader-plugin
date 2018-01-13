package home

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/billglover/chinese-reader/scanner"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type Request struct {
	Text  string `json:"text"`
	Token string `json:"token"`
}

type Response struct {
	Text   string `json:"string"`
	Score  int    `json:"readability"`
	Markup string `json:"markup"`
}

func handleRequest(rw http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	log.Infof(ctx, "/api request received")

	// TODO:
	// - validate user token
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

	valid, err := validateToken(ctx, mreq.Token)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	if valid == false {
		http.Error(rw, "invalid token", http.StatusUnauthorized)
	}

	// TODO:
	// - retrieve user's word list
	words, err := retrieveWordsList(ctx, mreq.Token)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	// TODO:
	// - scan the file
	score, markup, _ := scanner.Scan(mreq.Text, words)

	mresp := Response{
		Text:   mreq.Text,
		Score:  score,
		Markup: markup,
	}

	header := rw.Header()
	header.Add("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(mresp)

	// - return the results
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

	req, _ := http.NewRequest("PATCH", scheme+"://"+tokenURL+"/token/"+token+"?action=use", nil)

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

	log.Infof(ctx, string(body))

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	return true, nil
}

func retrieveWordsList(ctx context.Context, token string) (string, error) {

	svcName := "words"
	wordsURL, err := appengine.ModuleHostname(ctx, svcName, "", "")
	if err != nil {
		return "", fmt.Errorf("unable to find service %s", svcName)
	}

	scheme := "https"
	if appengine.IsDevAppServer() {
		scheme = "http"
	}

	req, _ := http.NewRequest("GET", scheme+"://"+wordsURL+"/words/"+token, nil)

	client := urlfetch.Client(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to query internal service")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read response from internal service")
	}

	return string(body), nil
}
