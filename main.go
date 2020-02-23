package main

import (
	"fmt"
	"io/ioutil"
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
	// fmt.Print(string(code))

	var renamedVariables = make(map[string]string)
	var shortNames = NewShortNames()

	var s scanner.Scanner
	s.Init(strings.NewReader(string(code)))

	//
	for token := s.Scan(); token != scanner.EOF; token = s.Scan() {
		tokenText := s.TokenText()
		if isVariable(tokenText) {
			s.Scan()
			varName := s.TokenText()
			s.Scan()
			varType := s.TokenText()
			// TODO this should be a function to be reused for other cases (short hand assignment and params)

			// only write a new entry for the variable if it is not already in the map
			_, variableExists := renamedVariables[varName]
			if !variableExists {
				// only write if we found a new shortName
				newShortName, shortNameFound := shortNames.nextShortName(varType, varName)
				if shortNameFound {
					renamedVariables[varName] = newShortName
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
	// can we do it in one loop?
	// would automatically exclude package stuff and only replace declared variables names

	// replace:
	// words from the map
	// exclude? before (
	// exclude package declaration

}
