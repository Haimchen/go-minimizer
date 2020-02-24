package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/scanner"
	"unicode"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

/* ++++++++++++++++++++++++++++++++++
Handling the read buffer
+++++++++++++++++++++++++++++++ */
type buffer struct {
	buffer []string
	size   int
}

func NewBuffer(size int) *buffer {

	buf := make([]string, 0, size)
	return &buffer{buf, size}
}

func (buffer *buffer) Push(token string) bool {
	if len(buffer.buffer) == buffer.size {
		return false
	}
	buffer.buffer = append(buffer.buffer, token)
	return true
}

func (buffer *buffer) Skim() (string, bool) {
	length := len(buffer.buffer)
	if length != buffer.size {
		return "", false
	}
	word := buffer.buffer[0]
	buffer.buffer = buffer.buffer[1:]
	return word, true
}

func (buffer *buffer) Shift() (string, bool) {
	if len(buffer.buffer) == 0 {
		return "", false
	}
	word := buffer.buffer[0]
	buffer.buffer = buffer.buffer[1:]
	return word, true
}

func (buffer *buffer) Replace(offset int, newWord string) bool {
	last := len(buffer.buffer) - 1
	if last-offset < 0 {
		return false
	}
	buffer.buffer[last-offset] = newWord
	return true
}

func (buffer *buffer) Read(offset int) (string, bool) {
	last := len(buffer.buffer) - 1
	if last-offset < 0 {
		return "", false
	}
	return buffer.buffer[last-offset], true
}

func (buffer *buffer) Print() {
	fmt.Println("Printing buffer:")
	for _, word := range buffer.buffer {
		fmt.Println(word)
	}
}

/* ++++++++++++++++++++++++++++++++++
Handling the short names
+++++++++++++++++++++++++++++++ */
var replacementLetters = "abcdefghijklmnopqrstuvxyz"

type shortNames struct {
	set map[string]bool
}

func NewShortNames() *shortNames {
	return &shortNames{make(map[string]bool)}
}

// TODO use multiple characters if all single characters are taken
func (shortNames *shortNames) nextShortName(providedWords ...string) (string, bool) {
	// add defaults as additonal word
	words := append(providedWords, replacementLetters)

	// iterate over all characters in all provided words
	for _, word := range words {
		for _, char := range word {
			newName := strings.ToLower(string(char))
			_, ok := shortNames.set[newName]
			// find a character that is not yet in use
			if !ok {
				// add new value to set
				shortNames.set[newName] = true
				return newName, true
			}
		}
	}

	return "", false
}

/* ++++++++++++++++++++++ */

func isVariable(token string) bool {
	return token == "var"
}

func outputFileName(inputFile string) string {
	suffix := "_min"
	outputFile := strings.Replace(inputFile, ".go", suffix+".go", 1)
	// fmt.Printf("Output: %s\n", outputFilePath)
	return outputFile
}

func isSpaceString(token string) bool {
	return token == " "
}

func isWhiteSpace(token string) bool {
	if len(token) == 1 {
		runes := []rune(token)
		return unicode.IsSpace(runes[0])
	}
	return false
}

func isWord(whiteSpaceOrWord string) bool {
	isSpace := isWhiteSpace(whiteSpaceOrWord)
	isEmpty := len(whiteSpaceOrWord) == 0

	return !isSpace && !isEmpty
}
func handleShortAssignment(renamedVariables map[string]string, shortNames *shortNames, buffer *buffer) {
	// bla :=
	// buffer: [... "bla" , ":", "="]
	// or:
	// buffer: [... "bla", " ", ":", "="]
	colon, _ := buffer.Read(1)

	// the = could also be an assignment only or part of a comparison
	// we are only interested if there is a : before
	if colon == ":" {
		// prev character can be either whitespace or the variable name
		whiteSpaceOrToken, _ := buffer.Read(2)
		offset := 2
		if !isWord(whiteSpaceOrToken) {
			whiteSpaceOrToken, _ = buffer.Read(3)
			offset = 3
		}
		if isWord(whiteSpaceOrToken) {
			newShortName, shortNameFound := shortNames.nextShortName(whiteSpaceOrToken)
			if shortNameFound {
				renamedVariables[whiteSpaceOrToken] = newShortName
				buffer.Replace(offset, newShortName)
			}
		}
	}
}

func handleVar(s *scanner.Scanner, tokenText string, renamedVariables map[string]string, shortNames *shortNames, buffer *buffer) string {
	s.Scan() // space
	s.Scan() // var name
	varName := s.TokenText()
	var overflow1, overflow2 string

	// only write a new entry for the variable if it is not already in the map
	_, variableExists := renamedVariables[varName]
	if !variableExists {
		// only write if we found a new shortName
		newShortName, shortNameFound := shortNames.nextShortName(varName)
		if shortNameFound {
			renamedVariables[varName] = newShortName
			overflow1, _ = buffer.Skim()
			buffer.Push(" ")
			overflow2, _ = buffer.Skim()
			buffer.Push(newShortName)
		}
	}
	return overflow1 + overflow2
}

func NewScanner(input []byte) *scanner.Scanner {
	var s scanner.Scanner
	s.Init(strings.NewReader(string(input)))
	// scanner ignores comments, so those will be removed!
	// don't skip any whitespace while scanning
	s.Whitespace ^= 1<<'\t' | 1<<'\n' | 1<<'\v' | 1<<'\f' | 1<<'\r' | 1<<' '
	return &s
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No file provided. Please provide a filepath as argument")
		os.Exit(0)
	}

	inputFilePath := os.Args[1]

	code, err := ioutil.ReadFile(inputFilePath)
	check(err)

	outputFilePath := outputFileName(inputFilePath)
	file, err := os.Create(outputFilePath)
	check(err)
	defer file.Close()

	var renamedVariables = make(map[string]string)
	var shortNames = NewShortNames()

	s := NewScanner(code)

	// create the read buffer
	buffer := NewBuffer(4)

	for token := s.Scan(); token != scanner.EOF; token = s.Scan() {
		tokenText := s.TokenText()
		// check for multipe space tokens in a row and only keep one
		// removing multiple whitespace tokens is harder, because it may remove linebreaks with tabs after them :(
		prevToken, hasPrev := buffer.Read(0)
		if hasPrev && isSpaceString(tokenText) && isSpaceString(prevToken) {
			buffer.Replace(0, tokenText)
		} else {
			buffer.Push(tokenText)
		}

		// is tokenText a word that needs to be replaced?
		replacement, exists := renamedVariables[tokenText]
		if exists {
			buffer.Replace(0, replacement)
		}

		// is it a variable declaration with var?
		if isVariable(tokenText) {
			overflow := handleVar(s, tokenText, renamedVariables, shortNames, buffer)
			file.Write([]byte(overflow))
		}
		// find variables assigned with :=
		if tokenText == "=" {
			handleShortAssignment(renamedVariables, shortNames, buffer)
		}

		// TODO find function params ?
		// TODO named function return values ?

		// if the buffer is full, we skim the first entry and write it
		tokenToWrite, bufferFull := buffer.Skim()
		if bufferFull {
			file.Write([]byte(tokenToWrite))
		}

		// fmt.Printf("%s: %s\n", s.Position, tokenText)
	}
	// empty the buffer after everything was read
	for token, ok := buffer.Shift(); ok; token, ok = buffer.Shift() {
		file.Write([]byte(token))
	}

	// DEBUG: output all replaced names
	for varName, shortName := range renamedVariables {
		fmt.Printf("%s: %s\n", varName, shortName)
	}

}
