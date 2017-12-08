package scanner

import "testing"

func TestScanSingleChar(t *testing.T) {
	known := `一
		二
		三
		四
		五
		六
		七
		八
		九
		零`

	text := "我知道一，二，三，四，和五，八点吧。"
	dScore := 50
	dMarkup := "我知道<span class=\"text-primary border border-primary\">一</span>，<span class=\"text-primary border border-primary\">二</span>，<span class=\"text-primary border border-primary\">三</span>，<span class=\"text-primary border border-primary\">四</span>，和<span class=\"text-primary border border-primary\">五</span>，<span class=\"text-primary border border-primary\">八</span>点吧。"

	score, markup, err := Scan(text, known)
	if err != nil {
		t.Errorf("unexpected error returned: %s", err)
	}
	if dScore != score {
		t.Errorf("unexpected score returned: want %d, got %d", dScore, score)
	}
	if dMarkup != markup {
		t.Errorf("unexpected markup returned:\n\twant: %s\n\tgot:  %s", dMarkup, markup)
	}
}
