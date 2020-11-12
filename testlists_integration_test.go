package engine

import (
	"testing"
)

func TestQueryTestListsURLs(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess := newSessionForTesting(t)
	defer sess.Close()
	config := &TestListsURLsConfig{}
	config.AddCategory("NEWS")
	config.Limit = 7
	result, err := sess.QueryTestListsURLs(config)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected non nil result")
	}
	for idx := int64(0); idx < result.Count(); idx++ {
		entry := result.At(idx)
		if entry == nil {
			t.Fatal("expecyed non-nil entry here")
		}
		if entry.URL == "" {
			t.Fatal("expected non empty URL here")
		}
		if entry.CategoryCode != "NEWS" {
			t.Fatal("expected another category here")
		}
		if entry.CountryCode == "" {
			t.Fatal("expected non empty country-code here")
		}
	}
	if result.At(-1) != nil {
		t.Fatal("expected nil entry here")
	}
	if result.At(result.Count()) != nil {
		t.Fatal("expected nil entry here")
	}
}

func TestQueryTestListsURLsQueryFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess := newSessionForTesting(t)
	defer sess.Close()
	config := &TestListsURLsConfig{BaseURL: "\t"}
	result, err := sess.QueryTestListsURLs(config)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if result != nil {
		t.Fatal("expected nil result here")
	}
}

func TestQueryTestListsURLsNilConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test in short mode")
	}
	sess := newSessionForTesting(t)
	defer sess.Close()
	result, err := sess.QueryTestListsURLs(nil)
	if err == nil {
		t.Fatal("expected an error here")
	}
	if result != nil {
		t.Fatal("expected nil result here")
	}
}
