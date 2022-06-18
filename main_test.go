package main

import (
	"bytes"
	"reflect"
	"regexp"
	"testing"
)

func TestPrintsResults(t *testing.T) {
	buffer := bytes.Buffer{}
	result := [][]int{{0, 1}}
	userString := "012345678910"
	PrintResults(&buffer, userString, result)

	got := buffer.String()
	want := "\x1b[41m0\x1b[0m12345678910"

	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestReturnsMatch(t *testing.T) {
	regex := regexp.MustCompile(`a.`)
	comparitor := "paranormal"
	matches := ReturnsMatch(regex, comparitor)

	got := matches
	want := [][]int{{1, 3}, {3, 5}, {8, 10}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q want %q", got, want)
	}
}

func BenchmarkPrintResults(b *testing.B) {
  buffer := bytes.Buffer{}
  result := [][]int{{0, 1}, {2,4}, {5, 6}, {7, 10}, {22, 46}, {59, 100}, {101, 102}}
  userString := "Thw quick brown fox jumped over the lazy dog ang the lazy dog was jumped over by the quick brown fox.Thw quick brown fox jumped over the lazy dog ang the lazy dog was jumped over by the quick brown fox.Thw quick brown fox jumped over the lazy dog ang the lazy dog was jumped over by the quick brown fox.Thw quick brown fox jumped over the lazy dog ang the lazy dog was jumped over by the quick brown fox.Thw quick brown fox jumped over the lazy dog ang the lazy dog was jumped over by the quick brown fox.Thw quick brown fox jumped over the lazy dog ang the lazy dog was jumped over by the quick brown fox.Thw quick brown fox jumped over the lazy dog ang the lazy dog was jumped over by the quick brown fox.Thw quick brown fox jumped over the lazy dog ang the lazy dog was jumped over by the quick brown fox."

  for n := 0; n < b.N; n++ {
    PrintResults(&buffer, userString, result)
  }
}
