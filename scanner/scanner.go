package scanner

import (
	"bufio"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Scan looks through a string of text and matches characters against
// a list of known characters. It returns an overall match score which
// indicates the percentage of characters that exist in the known list
// and a marked up version of the original text highlighting known
// characters. It returns an error if it is unable to parse the text.
func Scan(text, known string) (int, string, error) {

	found := 0
	miss := 0
	markup := ""

	r := strings.NewReader(known)
	ws, err := mapWords(r)
	if err != nil {
		return 0, markup, err
	}

	maxknown := 0
	for i := range ws {
		if utf8.RuneCountInString(i) > maxknown {
			maxknown = utf8.RuneCountInString(i)
		}
	}

	rs := []rune(text)
out:
	for i := 0; i < len(rs); i++ {
		max := i + maxknown
		if max > len(rs) {
			max = len(rs)
		}

		for mi := max; mi > i; mi-- {
			if ws[string(rs[i:mi])] == true {
				markup = markup + "<b>" + string(rs[i:mi]) + "</b>"
				found += (mi - i)
				i = mi - 1
				continue out
			}
		}
		if i > len(rs) {
			break out
		}
		markup += string(rs[i])

		if unicode.Is(unicode.Han, rs[i]) {
			miss++
		}

	}

	score := found * 100 / (found + miss)

	return score, markup, nil
}

// MapWords takes a reader on a byte stream and returns a map of words.
// Words in the byte stream should be line separated. It returns an error
// if it fails to read from the reader.
func mapWords(r io.Reader) (map[string]bool, error) {

	words := map[string]bool{}

	// use a buffered reader to make use of the readline functionality
	br := bufio.NewReader(r)

	for {
		b, _, err := br.ReadLine()

		if err != nil {
			if err == io.EOF {
				break
			}

			return words, err
		}

		words[strings.TrimSpace(string(b))] = true
	}

	return words, nil
}
