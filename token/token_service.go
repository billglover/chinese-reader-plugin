package token

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/billglover/uid"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type Token struct {
	ID        string    `json:"id,omitempty"`
	Created   time.Time `json:"created,omitempty"`
	Remaining int32     `json:"remaining,omitempty"`
}

func init() {
	r := register(context.Background())
	http.Handle("/", r)
}

func register(ctx context.Context) *mux.Router {
	r := mux.NewRouter()

	// TODO: pass context function to handler

	r.HandleFunc("/token", PostTokenHandler).Methods("POST")
	r.HandleFunc("/token/{id}", GetTokenHandler).Methods("GET")
	//r.HandleFunc("/token/{id}", PutTokenHandler).Methods("PUT")

	return r
}

func PostTokenHandler(w http.ResponseWriter, r *http.Request) {
	id, err := uid.NextStringID()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	t := Token{ID: id}

	ctx := appengine.NewContext(r)
	tokenKey := datastore.NewKey(ctx, "tokens", t.ID, 0, nil)

	k, err := datastore.Put(ctx, tokenKey, &t)
	if err != nil {
		log.Errorf(ctx, "unable to create token: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Infof(ctx, "created token: %s", k.StringID())

	respondWithJSON(w, http.StatusCreated, t)
}

func GetTokenHandler(w http.ResponseWriter, r *http.Request) {

	//vars := mux.Vars(r)
	//id := vars["id"]

	t := Token{}

	respondWithJSON(w, http.StatusOK, t)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
