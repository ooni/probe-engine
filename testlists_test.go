package engine

import (
	"testing"
)

func TestLookupTestListsSuccess(t *testing.T) {
	sess := newSessionForTesting(t)
	config := sess.NewTestListsConfig()
	config.Limit = 14
	client := sess.NewTestListsClient()
	urls, err := client.Fetch(config)
	if err != nil {
		t.Fatal(err)
	}
	for _, url := range urls {
		t.Logf("%s\t%s\t%s", url.CountryCode(), url.CategoryCode(), url.URL())
	}
}

func TestLookupTestListsFailure(t *testing.T) {
	sess := newSessionForTesting(t)
	config := sess.NewTestListsConfig()
	config.BaseURL = "\t"
	client := sess.NewTestListsClient()
	urls, err := client.Fetch(config)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if urls != nil {
		t.Fatal("expected nil urls")
	}
}
