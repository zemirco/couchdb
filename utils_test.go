package couchdb

import (
	"net/url"
	"reflect"
	"testing"
)

// quote()
var quoteTests = []struct {
	in  url.Values
	out url.Values
}{
	{
		url.Values{"key": []string{"value"}},
		url.Values{"key": []string{"\"value\""}},
	},
	{
		url.Values{"descending": []string{"true"}},
		url.Values{"descending": []string{"\"true\""}},
	},
}

func TestQuote(t *testing.T) {
	for _, tt := range quoteTests {
		actual := quote(tt.in)
		if !reflect.DeepEqual(actual, tt.out) {
			t.Errorf("quote(%v): expected %v, actual %v", tt.in, tt.out, actual)
		}
	}
}

// mimeType()
var mimeTypeTests = []struct {
	in  string
	out string
}{
	{"image.jpg", "image/jpeg"},
	{"presentation.pdf", "application/pdf"},
	{"file.text", "text/plain; charset=utf-8"},
	{"archive.zip", "application/zip"},
	{"movie.avi", "video/x-msvideo"},
}

func TestMimeType(t *testing.T) {
	for _, tt := range mimeTypeTests {
		actual := mimeType(tt.in)
		if actual != tt.out {
			t.Errorf("mimeType(%s): expected %s, actual %s", tt.in, tt.out, actual)
		}
	}
}
