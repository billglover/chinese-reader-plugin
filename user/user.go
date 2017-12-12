package user

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"github.com/billglover/uid"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
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
	switch r.Method {
	case http.MethodPost:
		userPost(w, r)
	case http.MethodGet:
		userGet(w, r)
	case http.MethodPut:
		userPut(w, r)
	default:
		http.Error(w, "method not implemented", http.StatusNotImplemented)
	}
}

func userPost(w http.ResponseWriter, r *http.Request) {
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
		log.Errorf(ctx, "unable to create user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Infof(ctx, "created user: %s", k.StringID())

	// return the user record
	header := w.Header()
	header.Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

func userGet(w http.ResponseWriter, r *http.Request) {
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
}

func userPut(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	id := path.Base(r.URL.Path)

	// if the ID is empty we consider it an invalid request
	if id == "." || id == "/" {
		http.Error(w, "missing user token", http.StatusBadRequest)
		return
	}

	// extract the updates from the payload
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf(ctx, "unable to read request body: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer r.Body.Close()

	newu := new(User)
	err = json.Unmarshal(body, &newu)
	if err != nil {
		log.Errorf(ctx, "unable to parse request body: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// create a key based on the supplied user ID and query datastore
	userKey := datastore.NewKey(ctx, "users", id, 0, nil)

	oldu := new(User)
	if err := datastore.Get(ctx, userKey, oldu); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Only update fields that allow modification. Leavethe other fields untouched.
	oldu.Remaining = newu.Remaining

	k, err := datastore.Put(ctx, userKey, oldu)
	if err != nil {
		log.Errorf(ctx, "unable to modify user %s: %v", oldu.ID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Infof(ctx, "modified user: %s", k.StringID())

	// return the user record
	header := w.Header()
	header.Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(oldu)
}
