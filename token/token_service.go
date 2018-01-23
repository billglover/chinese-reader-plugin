package token

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"google.golang.org/appengine"

	"github.com/billglover/uid"
	"github.com/gorilla/mux"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/client"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

const (
	// TokenCount specifies the number of uses assigned to new tokens.
	TokenCount int = 1000

	// TokenExpiryDays specifies the number of days after which a token should expire
	TokenExpiryDays int = 365

	// StripeKey identifies the application to Stripe
	StripeKey string = "sk_test_ovUaN3GKcKu9SUM94ueaAzxf"
)

// Token is a struct that holds details of a user token. Tokens have a unique
// identifier, a created timestamp and a counter indicating the number of times
// a token can be used before it expires.
type Token struct {
	ID        string    `json:"id,omitempty"`
	Created   time.Time `json:"created,omitempty"`
	Expires   time.Time `json:"expires,omitempty"`
	Remaining int       `json:"remaining"`
	Valid     bool      `datastore:"-" json:"valid"`
}

type Request struct {
	ID string `json:"id"`
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
// TODO: create a response schema
// TODO: handle errors during payment
func PostTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	t, err := createToken(ctx)
	// if token generation fails, return without charging the user
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer r.Body.Close()

	var mreq Request
	err = json.Unmarshal(body, &mreq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	//stripeToken := r.FormValue("stripeToken")
	stripeToken := mreq.ID
	err = chargeUser(ctx, stripeToken, t.ID)
	// at this point we need to be very clear to the user whether they
	// have been charged or not.
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, t)
}

// GetTokenHandler handles an HTTP GET request. It returns the token
// that corresponds to the ID provided in the path.
// TODO: get token should be marked as internal only
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

	t.Valid = t.IsValid()
	if t.Valid == false {
		log.Infof(ctx, "unable to retrieve invalid token: %s, expired: %s, remaining: %d", t.ID, t.Expires, t.Remaining)
		respondWithJSON(w, http.StatusGone, t)
		return
	}

	respondWithJSON(w, http.StatusOK, t)
}

// PatchTokenHandler handles an update to the remaining use counter. It ensures that the
// counter can only be reduced by 1 on each update. All other update requests are
// treated as invalid.
// TODO: patch token should be marked as internal only
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

	// check the token is valid before any updates are made
	t.Valid = t.IsValid()
	if t.Valid == false {
		log.Infof(ctx, "unable to use invalid token: %s, expired: %s, remaining: %d", t.ID, t.Expires, t.Remaining)
		respondWithJSON(w, http.StatusGone, t)
		return
	}

	// reduce the number of remaining uses but cap at 0
	t.Remaining--
	if t.Remaining < 0 {
		t.Remaining = 0
	}

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

// CreateToken creates an individual token with default values
func createToken(ctx context.Context) (Token, error) {
	var t Token

	id, err := uid.NextStringID()
	if err != nil {
		return t, err
	}

	now := time.Now()

	t = Token{
		ID:        id,
		Created:   now,
		Expires:   now.AddDate(0, 0, TokenExpiryDays),
		Remaining: TokenCount,
	}

	tokenKey := datastore.NewKey(ctx, "tokens", t.ID, 0, nil)
	_, err = datastore.Put(ctx, tokenKey, &t)
	if err != nil {
		return t, err
	}

	t.Valid = t.IsValid()
	return t, nil
}

// ChargeUser attempts to charge a users card and indicates whether
// the charge was successful or not.
func chargeUser(ctx context.Context, cardToken, userToken string) error {
	stripe.Key = StripeKey

	// We create a custom client because App Engine
	httpClient := urlfetch.Client(ctx)
	stripeClient := client.New(stripe.Key, stripe.NewBackends(httpClient))

	// Charge the user's card:
	params := &stripe.ChargeParams{
		Amount:   500,
		Currency: "gbp",
		Desc:     "Chinese Reader Token",
	}
	params.AddMeta("order_id", userToken)
	params.SetSource(cardToken)

	// TODO: return more useful charge errors to the caller
	charge, err := stripeClient.Charges.New(params)
	if err != nil {
		return err
	}

	if charge.Status != "succeeded" {
		log.Errorf(ctx, charge.FailMsg)
		return fmt.Errorf(charge.FailMsg)
	}

	return nil
}

// IsValid takes a token and determines whether it is still valid
func (t Token) IsValid() bool {
	valid := true

	// check we have remaining uses
	if t.Remaining <= 0 {
		valid = false
	}

	// check the date hasn't expired
	if time.Now().After(t.Expires) {
		valid = false
	}

	return valid
}
