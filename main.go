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

	for token := s.Scan(); token != scanner.EOF; token = s.Scan() {
		tokenText := s.TokenText()

		// is tokenText a word that needs to be replaced?
		replacement, exists := renamedVariables[tokenText]
		if exists {
			file.Write([]byte(replacement))
		} else {
			file.Write([]byte(tokenText))
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
					file.Write([]byte(" " + newShortName /* + " " + varType*/))
				}
			}
		}

		// TODO find variables assigned with :=

		// TODO find function params

		// TODO function return values

		fmt.Printf("%s: %s\n", s.Position, tokenText)
	}

	// DEBUG: output all replaced names
	for varName, shortName := range renamedVariables {
		fmt.Printf("%s: %s\n", varName, shortName)
	}

}
