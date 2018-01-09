package words

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/file"
	"google.golang.org/appengine/log"
)

func init() {
	r := mux.NewRouter()
	r.HandleFunc("/words", PostWordsHandler).Methods("POST")
	r.HandleFunc("/words/{id}", GetWordsHandler).Methods("GET")
	r.HandleFunc("/words/{id}", DeleteWordsHandler).Methods("DELETE")
	r.HandleFunc("/words/{id}", PutWordsHandler).Methods("PUT")

	http.Handle("/", r)
}

func PostWordsHandler(w http.ResponseWriter, r *http.Request) {

	// We get the context for the incoming HTTP request
	// https://cloud.google.com/appengine/docs/standard/go/reference#NewContext
	ctx := appengine.NewContext(r)

	// Get the token value from the uploaded data
	token := r.FormValue("token")
	if token == "" {
		respondWithError(w, http.StatusBadRequest, "no token provided")
		return
	}

	// We expect our file to be uploaded as part of a multipart
	// form. We could probably do more here to validate the form
	// that has been uploaded.s
	f, fh, err := r.FormFile("words")
	if err == http.ErrMissingFile {
		respondWithError(w, http.StatusBadRequest, "no file provided")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("bad request: %v:", err))
		return
	}

	// TODO: blog post this
	// Write the file to Google Cloud Storage
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Errorf(ctx, "failed to get client: %v", err)
	}

	bucketName, err := file.DefaultBucketName(ctx)
	if err != nil {
		log.Errorf(ctx, "failed to get default GCS bucket name: %v", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("storage service failure: %v:", err))
		return
	}
	bucket := client.Bucket(bucketName)

	log.Infof(ctx, "Received file %s for token %s", fh.Filename, token)
	log.Infof(ctx, "Creating file /%v/%v\n", bucketName, token)

	obj := bucket.Object(token)
	objw := obj.NewWriter(ctx)

	if _, err := io.Copy(objw, f); err != nil {
		log.Errorf(ctx, "failed to copy file: %v", err.Error())
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("storage service failure: %v:", err))
		return
	}
	if err := objw.Close(); err != nil {
		log.Errorf(ctx, "failed to close storage object: %v", err.Error())
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("storage service failure: %v:", err))
		return
	}

	f.Close()

	respondWithJSON(w, http.StatusCreated, nil)
}

func GetWordsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	vars := mux.Vars(r)
	id := vars["id"]

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Errorf(ctx, "failed to get client: %v", err)
	}

	bucketName, err := file.DefaultBucketName(ctx)
	if err != nil {
		log.Errorf(ctx, "failed to get default GCS bucket name: %v", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("storage service failure: %v:", err))
		return
	}
	bucket := client.Bucket(bucketName)

	obj := bucket.Object(id)
	objr, err := obj.NewReader(ctx)
	if err != nil {
		log.Errorf(ctx, "unable to read object: %v", err)
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("record not found: %s:", id))
		return
	}

	io.Copy(w, objr)
	objr.Close()
}

func DeleteWordsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	vars := mux.Vars(r)
	id := vars["id"]

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Errorf(ctx, "failed to get client: %v", err)
	}

	bucketName, err := file.DefaultBucketName(ctx)
	if err != nil {
		log.Errorf(ctx, "failed to get default GCS bucket name: %v", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("storage service failure: %v:", err))
		return
	}
	bucket := client.Bucket(bucketName)

	obj := bucket.Object(id)
	err = obj.Delete(ctx)
	if err != nil {
		log.Errorf(ctx, "failed to delete object: %v", err)
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("record not found: %s:", id))
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

func PutWordsHandler(w http.ResponseWriter, r *http.Request) {

	// TODO: check that the file exists before accepting files of an arbitrary name

	ctx := appengine.NewContext(r)

	vars := mux.Vars(r)
	id := vars["id"]

	f, _, err := r.FormFile("words")
	if err == http.ErrMissingFile {
		respondWithError(w, http.StatusBadRequest, "no file provided")
		return
	}
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("bad request: %v:", err))
		return
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Errorf(ctx, "failed to get client: %v", err)
	}

	bucketName, err := file.DefaultBucketName(ctx)
	if err != nil {
		log.Errorf(ctx, "failed to get default GCS bucket name: %v", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("storage service failure: %v:", err))
		return
	}
	bucket := client.Bucket(bucketName)

	obj := bucket.Object(id)
	objw := obj.NewWriter(ctx)

	if _, err := io.Copy(objw, f); err != nil {
		log.Errorf(ctx, "failed to copy file: %v", err.Error())
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("storage service failure: %v:", err))
		return
	}
	if err := objw.Close(); err != nil {
		log.Errorf(ctx, "failed to close storage object: %v", err.Error())
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("storage service failure: %v:", err))
		return
	}

	f.Close()

	respondWithJSON(w, http.StatusCreated, nil)
}
