package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/scanner"
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

// func (buffer *buffer) Read(pos int) (string, bool) {
//   if pos+1 > len(buffer.buffer) {
//     return "", false
//   }
//   return buffer.buffer[pos], true
// }
//
// func (buffer *buffer) Print() {
//   fmt.Println("Printing buffer:")
//   for _, word := range buffer.buffer {
//     fmt.Println(word)
//   }
// }

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

	var s scanner.Scanner
	s.Init(strings.NewReader(string(code)))
	// sanner ignores comments, so those will be removed!
	// don't skip any whitespace while scanning
	s.Whitespace ^= 1<<'\t' | 1<<'\n' | 1<<'\v' | 1<<'\f' | 1<<'\r' | 1<<' '

	// create the read buffer
	buffer := NewBuffer(4)

	for token := s.Scan(); token != scanner.EOF; token = s.Scan() {
		tokenText := s.TokenText()
		buffer.Push(tokenText)

		// is tokenText a word that needs to be replaced?
		replacement, exists := renamedVariables[tokenText]
		if exists {
			buffer.Replace(0, replacement)
			// } else {
			//   file.Write([]byte(tokenText))
		}

		// is it a variable declaration with var?
		if isVariable(tokenText) {
			s.Scan() // space
			s.Scan() // var name
			varName := s.TokenText()
			// s.Scan() // space
			// s.Scan() // type or =
			// varType := s.TokenText()
			// TODO this should be a function to be reused for other cases (short hand assignment and params)

			// only write a new entry for the variable if it is not already in the map
			_, variableExists := renamedVariables[varName]
			if !variableExists {
				// only write if we found a new shortName
				newShortName, shortNameFound := shortNames.nextShortName( /*varType, */ varName)
				if shortNameFound {
					renamedVariables[varName] = newShortName
					buffer.Push("")
					buffer.Push(newShortName)
					// file.Write([]byte(" " + newShortName [> + " " + varType<]))
				}
			}
		}

		tokenToWrite, bufferFull := buffer.Skim()
		if bufferFull {
			file.Write([]byte(tokenToWrite))
		}

		// TODO find variables assigned with :=

		// TODO find function params

		// TODO function return values

		// fmt.Printf("%s: %s\n", s.Position, tokenText)
	}
	// empty the buffer
	for token, ok := buffer.Shift(); ok; token, ok = buffer.Shift() {
		file.Write([]byte(token))
	}

	// DEBUG: output all replaced names
	for varName, shortName := range renamedVariables {
		fmt.Printf("%s: %s\n", varName, shortName)
	}

}
