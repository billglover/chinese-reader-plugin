package user

import (
	"encoding/json"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/billglover/uid"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

type User struct {
	ID        string    `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	Remaining int       `json:"remaining"`
}

func init() {
	http.HandleFunc("/", handleRequest)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {

		// create a user with a new unique ID
		id, _ := uid.NextStringID()
		u := User{
			ID:        id,
			CreatedAt: time.Now(),
			Remaining: 10,
		}

		// create a key for the user object
		ctx := appengine.NewContext(r)
		userKey := datastore.NewKey(ctx, "users", u.ID, 0, nil)

		// TODO: check payment before creating a user record

		// create the user record
		k, err := datastore.Put(ctx, userKey, &u)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		log.Println("created user:", k.StringID())

		// return the user record
		header := w.Header()
		header.Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(u)
		return
	}

	if r.Method == http.MethodGet {

		id := path.Base(r.URL.Path)

		// if the ID is empty we consider it an invalid request
		if id == "." || id == "/" {
			http.Error(w, "missing user token", http.StatusBadRequest)
			return
		}

		// create a key based on the supplied user ID and query datastore
		ctx := appengine.NewContext(r)
		userKey := datastore.NewKey(ctx, "users", id, 0, nil)

		u := new(User)
		if err := datastore.Get(ctx, userKey, u); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// return the user record as a JSON object
		header := w.Header()
		header.Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(u)
		return
	}

	// respond to indicate that the method is not implemented
	http.Error(w, "method not implemented", http.StatusNotImplemented)
}
