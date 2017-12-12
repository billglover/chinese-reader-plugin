package token

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/billglover/uid"
	"github.com/gorilla/mux"
)

type Token struct {
	ID        string    `json:"id,omitempty"`
	Created   time.Time `json:"created,omitempty"`
	Remaining int32     `json:"remaining,omitempty"`
}

func init() {
	r := register()
	http.Handle("/", r)
}

func register() *mux.Router {
	r := mux.NewRouter()
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
