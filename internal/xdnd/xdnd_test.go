package xdnd

import (
	"slices"
	"testing"
)

func TestURIListFiles(t *testing.T) {
	list := "# a comment\r\nfile:///tmp/a%20b.txt\r\nfile://localhost/etc/hosts\r\nhttps://example.com/x\r\n\r\nnot a uri at all\r\n"
	want := []string{"/tmp/a b.txt", "/etc/hosts"}
	if got := uriListFiles(list); !slices.Equal(got, want) {
		t.Errorf("uriListFiles = %q, want %q", got, want)
	}
	if got := uriListFiles(""); len(got) != 0 {
		t.Errorf("empty list gave %q", got)
	}
}
