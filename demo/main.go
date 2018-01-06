package chinese_reader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	// Note: relative import path required for AppEngine
	"scanner"
)

type Request struct {
	Text string `json:"text"`
}

type Response struct {
	Text   string `json:"string"`
	Score  int    `json:"readability"`
	Markup string `json:"markup"`
}

func handleRequest(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("hi")
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	defer req.Body.Close()

	var mreq Request
	err = json.Unmarshal(body, &mreq)
	if err != nil {
		log.Println(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	known := GetKnown()
	score, markup, _ := scanner.Scan(mreq.Text, known)

	mresp := Response{
		Text:   mreq.Text,
		Score:  score,
		Markup: markup,
	}

	header := rw.Header()
	header.Add("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(mresp)

}

func GetKnown() string {
	f, err := os.Open("data/words.txt")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return string(b)
}

func init() {
	http.HandleFunc("/api", handleRequest)
}
