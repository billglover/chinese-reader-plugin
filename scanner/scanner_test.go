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
	dMarkup := "我知道<b>一</b>，<b>二</b>，<b>三</b>，<b>四</b>，和<b>五</b>，<b>八</b>点吧。"

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
