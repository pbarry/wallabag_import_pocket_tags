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
			"https://www.reddit.com/r/California/comments/7clnq6/the_hidden_history_of_how_california_was_built_on/?st=J9YFKDC5&sh=05c66022",
			"https://www.reddit.com/r/California/comments/7clnq6/the_hidden_history_of_how_california_was_built_on/",
		},
		{
			"https://techcrunch.com/2018/04/28/emissary-wants-to-make-sales-networking-obsolete/",
			"https://techcrunch.com/2018/04/28/emissary-wants-to-make-sales-networking-obsolete/?guccounter=1&guce_referrer=aHR0cDovL3d3dy5nb29nbGUuY28udWsvdXJsP3NhPXQmc291cmNlPXdlYiZjZD0x&guce_referrer_sig=AQAAAJFe0OxYz8NQcrhbCOJdODOYLL29VOMQckZTJ_CZUpFGqcT40s0Kzkg1AlTHkEvIO3I5DddV7GfhH7rvPDygcJIYuk5xwsPygLIntRjyEPEj47gybA2cTN5rrKQOc-sX6H2FV93viCNNeGI9m-xO1S0GUhDyO6GDun3nzLt9lCvl",
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
