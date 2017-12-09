package user

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
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
		c := appengine.NewContext(r)
		userKey := datastore.NewKey(c, "users", u.ID, 0, nil)

		// TODO: check payment before creating a user record

		// create the user record
		k, err := datastore.Put(c, userKey, &u)
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

		// TODO: check for ID in the path
		// these needs to be done safely
		id := strings.Split(r.URL.Path, "/")
		log.Println(id[1])
		return
	}

	// respond to indicate that the method is not implemented
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
