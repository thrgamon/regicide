package main

import (
	"bytes"
	"reflect"
	"regexp"
	"testing"
)

func TestPrintsResults(t *testing.T) {
	buffer := bytes.Buffer{}
	result := []int{3, 6}
	userString := "012345678910"
	PrintResults(&buffer, userString, result)

	got := buffer.String()
	want := userString + "\n"

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
