package token

import (
	"encoding/json"
	"net/http"
	"time"

	"google.golang.org/appengine"

	"github.com/billglover/uid"
	"github.com/gorilla/mux"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const (
	// DefaultTokenCount specifies the number of uses assigned to new tokens
	// by default. Changing this value does not impact existing tokens.
	DefaultTokenCount int = 100
)

// Token is a struct that holds details of a user token. Tokens have a unique
// identifier, a created timestamp and a counter indicating the number of times
// a token can be used before it expires.
type Token struct {
	ID        string    `json:"id,omitempty"`
	Created   time.Time `json:"created,omitempty"`
	Remaining int       `json:"remaining,omitempty"`
}

func init() {
	r := mux.NewRouter()
	r.HandleFunc("/token", PostTokenHandler).Methods("POST")
	r.HandleFunc("/token/{id}", GetTokenHandler).Methods("GET")
	r.HandleFunc("/token/{id}", PatchTokenHandler).Methods("PATCH")

	http.Handle("/", r)
}

// PostTokenHandler handles an HTTP POST request. It creates a new token
// and sets the remaining use counter to the default value specified in
// the constants.
func PostTokenHandler(w http.ResponseWriter, r *http.Request) {

	id, err := uid.NextStringID()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	t := Token{
		ID:        id,
		Created:   time.Now(),
		Remaining: DefaultTokenCount,
	}

	ctx := appengine.NewContext(r)
	tokenKey := datastore.NewKey(ctx, "tokens", t.ID, 0, nil)
	resultKey, err := datastore.Put(ctx, tokenKey, &t)
	if err != nil {
		log.Errorf(ctx, "unable to create new token: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to create new token")
		return
	}

	log.Infof(ctx, "created token: %s", resultKey.StringID())
	respondWithJSON(w, http.StatusCreated, t)
}

// GetTokenHandler handles an HTTP GET request. It returns the token
// that corresponds to the ID provided in the path.
func GetTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	var t Token

	vars := mux.Vars(r)
	id := vars["id"]

	tokenKey := datastore.NewKey(ctx, "tokens", id, 0, nil)
	if err := datastore.Get(ctx, tokenKey, &t); err != nil {
		log.Errorf(ctx, "unable to locate token: %v", err)
		respondWithError(w, http.StatusNotFound, "unable to locate token")
		return
	}

	respondWithJSON(w, http.StatusOK, t)
}

// PatchTokenHandler handles an update to the remaining use counter. It ensures that the
// counter can only be reduced by 1 on each update. All other update requests are
// treated as invalid.
func PatchTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	var t Token

	vars := mux.Vars(r)
	id := vars["id"]

	action := r.URL.Query().Get("action")
	if action != "use" {
		log.Errorf(ctx, "invalid action requested")
		respondWithError(w, http.StatusBadRequest, "invalid action requested")
		return
	}

	tokenKey := datastore.NewKey(ctx, "tokens", id, 0, nil)
	if err := datastore.Get(ctx, tokenKey, &t); err != nil {
		log.Errorf(ctx, "unable to locate token: %v", err)
		respondWithError(w, http.StatusNotFound, "unable to locate token")
		return
	}

	t.Remaining--

	resultKey, err := datastore.Put(ctx, tokenKey, &t)
	if err != nil {
		log.Errorf(ctx, "unable to modify token: %v", err)
		respondWithError(w, http.StatusInternalServerError, "unable to modify token")
		return
	}

	log.Infof(ctx, "used token: %s, remaining: %d", resultKey.StringID(), t.Remaining)
	respondWithJSON(w, http.StatusOK, t)

}

// RespondWithError is a helper function that sets the HTTP status code and returns
// a JSON formatted error payload.
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// RespondWithJSON is a helper function that sets the HTTP status code and marshals
// a struct into a JSON payload.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
