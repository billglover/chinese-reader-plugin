package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"unicode"
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

	fmt.Println(mreq)

	score, markup := scoreText(mreq.Text)

	mresp := Response{
		Text:   mreq.Text,
		Score:  score,
		Markup: markup,
	}

	header := rw.Header()
	header.Add("Content-Type", "application/json")
	json.NewEncoder(rw).Encode(mresp)

}

func scoreText(t string) (int, string) {

	markup := ""
	score := 0
	count := 0

	f, err := os.Open("data/words.txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	words := map[string]bool{}

	r := bufio.NewReader(f)
	for {
		if b, _, err := r.ReadLine(); err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println(err)
				return score, markup
			}
		} else {
			words[string(b)] = true
		}
	}

	for _, c := range []rune(t) {
		if unicode.Is(unicode.Han, c) == true {

			count++

			if ok := words[string(c)]; ok == true {
				score++
				markup = markup + "<b>" + string(c) + "</b>"
				continue
			}

		}
		markup = markup + string(c)
	}

	if len(t) == 0 {
		return 0, markup
	}

	fmt.Println("score:", score)
	fmt.Println("count:", count)

	score = score * 100 / count
	return score, markup
}

func main() {
	http.HandleFunc("/api", handleRequest)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
