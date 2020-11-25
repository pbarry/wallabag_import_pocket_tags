package main

import (
	"testing"
)

func TestABunchOfUrlsMatching(t *testing.T) {
	tests := []struct {
		p string
		w string
	}{
		{
			"https://www.reddit.com/r/Food/the_hidden_history_of_avocados/?st=J9YF8FC5&sh=05c689022",
			"https://www.reddit.com/r/Food/the_hidden_history_of_avocados",
		},
		{
			"https://techcrunch.com/2018/ExcitingNews#SomeFancyJs",
			"https://techcrunch.com/2018/ExcitingNews",
		},
	}

	for _, current := range tests {
		p2 := canonicalizeUrl(getRedirectedUrl(current.p))
		w2 := canonicalizeUrl(current.w)
		if p2 != w2 {
			t.Errorf("Failed URL match:\nPocketURL: %s \nWallabagURL: %s\n", p2, w2)
		}
	}
}
