package main

import (
	"io"
	"regexp"
)

type Scanner struct {
	buffer      string
	size, pos   int
	lnnr, colnr int
}

func newScanner(inputString string) Scanner {
	return Scanner{
		buffer: inputString,
		size:   len(inputString),
		pos:    0,
		lnnr:   1,
		colnr:  1,
	}
}

func (scanner *Scanner) nextToken() *ScannerToken {
	// Scan as long as we have text left to read
	for scanner.pos < scanner.size-1 {
		maxMatchLength := 0
		var maxMatcher ScannerMatcher = nil

		// Find longest match
		for _, ele := range scannerMatchList {
			// Test if the Regex matches
			match := ele.getRegex().FindStringIndex(scanner.buffer[scanner.pos:])
			if match == nil || match[0] != 0 {
				// No match, try next token
				continue
			}

			// Regex matched. Get matched string.
			matchLength := match[1]
			if matchLength > maxMatchLength {
				maxMatchLength = matchLength
				maxMatcher = ele
			}
		}

		if maxMatcher == nil {
			// All matchers failed
			panic("Matching failure: No match found!") // TODO: Expand
		}

		matchedString := scanner.buffer[scanner.pos : scanner.pos+maxMatchLength]

		// Check which type of matcher we have and run action
		switch matcher := maxMatcher.(type) {
		case ScannerMatcherSpecial:
			// We have a special matcher which changes the scanner state (manually)
			matcher.matchAction(matchedString, scanner)

			// Advance the scanner
			scanner.pos += maxMatchLength
		case ScannerMatcherRegular:
			// Normal matcher which produces a token.
			// Generate the initial token
			inputToken := ScannerToken{
				tokenType: Undefined,
				text:      matchedString,
				lnnr:      scanner.lnnr,
				colnr:     scanner.colnr,
			}
			// Call matching action to modify the raw token
			outputToken := matcher.matchAction(matchedString, &inputToken)
			if outputToken == nil {
				// No token was generated
				panic("Matching failure: No token generated!")
			}

			// Advance the scanner
			scanner.colnr += maxMatchLength
			scanner.pos += maxMatchLength

			return outputToken
		}

	} // END for scanner.pos < scanner.size

	// End of buffer reached, return nil
	return nil
}

type ScannerTokenType int

const (
	Keyword ScannerTokenType = iota + 1
	Identifier

	Other     = 0
	Undefined = -1
)

type ScannerToken struct {
	tokenType   ScannerTokenType
	text        string
	lnnr, colnr int
}

type ScannerMatcher interface {
	getRegex() *regexp.Regexp
}

type ScannerMatcherRegular struct {
	matchRegex  *regexp.Regexp
	matchAction func(string, *ScannerToken) *ScannerToken
}

func (matcher ScannerMatcherRegular) getRegex() *regexp.Regexp {
	return matcher.matchRegex
}

type ScannerMatcherSpecial struct {
	matchRegex  *regexp.Regexp
	matchAction func(string, *Scanner)
}

func (matcher ScannerMatcherSpecial) getRegex() *regexp.Regexp {
	return matcher.matchRegex
}

// A List containing the regular expressions that define the tokens and a corresponding
// action to create such a token.
// The input is a pre-defined token. The function only has to change the corresponding values.
var scannerMatchList = []ScannerMatcher{
	ScannerMatcherSpecial{regexp.MustCompilePOSIX("^[ \t]+"),
		func(input string, scanner *Scanner) {
			scanner.colnr += len(input)
		}},

	ScannerMatcherSpecial{regexp.MustCompilePOSIX("^\n"),
		func(input string, scanner *Scanner) {
			scanner.lnnr++
			scanner.colnr = 1
		}},

	ScannerMatcherSpecial{regexp.MustCompilePOSIX("^//.*\n"),
		func(input string, scanner *Scanner) {
			scanner.lnnr++
			scanner.colnr = 1
		}},

	ScannerMatcherRegular{regexp.MustCompilePOSIX("^if"),
		func(input string, token *ScannerToken) *ScannerToken {
			token.tokenType = Keyword
			return token
		}},

	ScannerMatcherRegular{regexp.MustCompilePOSIX("^[A-Za-z0-9]+"),
		func(input string, token *ScannerToken) *ScannerToken {
			token.tokenType = Identifier
			return token
		}},
}

func ScanningPhase(input io.Reader) []ScannerToken {
	// Load entire input file
	inputString, err := io.ReadAll(input)
	if err != nil {
		panic(err)
	}

	// Initialize Scanner and Token Stream
	var scanner Scanner = newScanner(string(inputString))

	var tokenStream []ScannerToken = make([]ScannerToken, 0)

	for token := scanner.nextToken(); token != nil; token = scanner.nextToken() {
		tokenStream = append(tokenStream, *token)
	}

	return tokenStream
}

