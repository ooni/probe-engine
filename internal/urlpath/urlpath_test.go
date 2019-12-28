package urlpath

import "testing"

func TestAppend(t *testing.T) {
	if Append("/foo", "bar/baz") != "/foo/bar/baz" {
		t.Fatal("unexpected result")
	}
	if Append("/foo", "/bar/baz") != "/foo/bar/baz" {
		t.Fatal("unexpected result")
	}
	if Append("/foo/", "bar/baz") != "/foo/bar/baz" {
		t.Fatal("unexpected result")
	}
	if Append("/foo/", "/bar/baz") != "/foo/bar/baz" {
		t.Fatal("unexpected result")
	}
}
