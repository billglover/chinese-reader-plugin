package home

import (
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func handleRequest(rw http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)
	log.Infof(ctx, "/api request received")
}

func init() {
	http.HandleFunc("/api", handleRequest)
}
