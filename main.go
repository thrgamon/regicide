package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
  "github.com/fatih/color"
)

func main() {
  userRegex := os.Args[1]
  userString := os.Args[2]
  re := regexp.MustCompile(userRegex)
  results := ReturnsMatch(re, userString)
  for _, result := range results {
    PrintResults(os.Stdout, userString, result)
  }
}

func PrintResults(w io.Writer, userString string, matchIndex[]int) {
  // Get the beginning and the end of the match
  ms := matchIndex[0]
  me := matchIndex[1]

  // Split the string around the matches
  prefix := userString[:ms]
  result := userString[ms:me]
  suffix := userString[me:]

  // Setup the color function
  red := color.New(color.FgRed).SprintFunc()

  // Print, highlighting the match
  fmt.Fprintf(w, "%s%s%s\n", prefix, red(result), suffix)
}

func ReturnsMatch(re *regexp.Regexp, comparitor string) (results [][]int){
  ba := []byte(comparitor) 
  return re.FindAllIndex(ba, -1)
}
