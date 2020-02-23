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
type shortNames struct {
	set map[string]bool
}

func NewShortNames() *shortNames {
	return &shortNames{make(map[string]bool)}
}

func (shortNames *shortNames) nextShortName(words ...string) (string, bool) {
	// iterate over all characters in all provided words
	for _, word := range words {
		for _, char := range word {
			newName := string(char)
			_, ok := shortNames.set[newName]
			// find a character that is not yet in use
			if !ok {
				// add new value to set
				shortNames.set[newName] = true
				return newName, true
			}
		}
	}

	// TODO implement fallback using random character or multiple characters

	return "", false
}

/* ++++++++++++++++++++++ */

func isVariable(token string) bool {
	return token == "var"
}

func main() {
	code, err := ioutil.ReadFile("./data/simple.go")
	check(err)

	// TODO create file name dynamically from original file name
	file, err := os.Create("./data/simple_min.go")
	check(err)
	defer file.Close()

	var renamedVariables = make(map[string]string)
	var shortNames = NewShortNames()

	var s scanner.Scanner
	s.Init(strings.NewReader(string(code)))
	s.Whitespace ^= 1<<'\t' | 1<<'\n' | 1<<'\v' | 1<<'\f' | 1<<'\r' | 1<<' ' // don't skip tab and newline

	// scanner ignores comments and whitespace by defaut!
	for token := s.Scan(); token != scanner.EOF; token = s.Scan() {
		tokenText := s.TokenText()

		// is tokenText a word that needs to be replaced?
		replacement, exists := renamedVariables[tokenText]
		if exists {
			file.Write([]byte(replacement))
		} else {
			file.Write([]byte(tokenText))
		}

		if isVariable(tokenText) {
			s.Scan() // space
			s.Scan() // var name
			varName := s.TokenText()
			s.Scan() // space
			s.Scan() // type
			varType := s.TokenText()
			// TODO this should be a function to be reused for other cases (short hand assignment and params)

			// only write a new entry for the variable if it is not already in the map
			_, variableExists := renamedVariables[varName]
			if !variableExists {
				// only write if we found a new shortName
				newShortName, shortNameFound := shortNames.nextShortName(varType, varName)
				if shortNameFound {
					renamedVariables[varName] = newShortName
					file.Write([]byte(" " + newShortName + " " + varType))
				}
			}
		}

		// TODO find variables assigned with :=

		// TODO find function params

		// TODO function return values

		fmt.Printf("%s: %s\n", s.Position, tokenText)
	}

	for varName, shortName := range renamedVariables {
		fmt.Printf("%s: %s\n", varName, shortName)
	}

	// scan again and write everything with updates
	// can't do it in the same loop, scanner removes all whitespace and code would look horrible

	// replace:
	// words from the map
	// exclude? before (
	// exclude package declaration

}
