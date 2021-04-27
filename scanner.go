package main

import (
	"io"
	"regexp"
)

type ScannerTokenType int

const (
	KeywordQuery ScannerTokenType = iota + 1
	Identifier

	Other     = 0
	Undefined = -1
)

type ScannerToken struct {
	tokenType   ScannerTokenType
	text        string
	lnnr, colnr int
}

type Scanner struct {
	buffer      string
	size, pos   int
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
	var scanner Scanner = Scanner{
		buffer: string(inputString),
		size:   len(inputString),
		pos:    0,
		lnnr:   1,
		colnr:  1,
	}

	var tokenStream []ScannerToken = make([]ScannerToken, 0)

	// Scan as long as we have text left to read
	for scanner.pos < scanner.size-1 {
		for _, ele := range scannerMatchList {
			// Test if the Regex matches
			match := ele.getRegex().FindStringIndex(scanner.buffer[scanner.pos:])
			if match == nil || match[0] != 0 {
				continue
			}

			// Regex matched. Get matched string.
			matchLength := match[1]
			matchedString := scanner.buffer[scanner.pos : scanner.pos+matchLength]

			// Check which type of matcher we have
			switch matcher := ele.(type) {
			case ScannerMatcherSpecial:
				// We have a special matcher which changes the scanner state (manually)
				matcher.matchAction(matchedString, &scanner)
				// Advance the scanner
				scanner.pos += matchLength
			case ScannerMatcherRegular:
				// Normal matcher which produces a token.
				// Generate the initial token
				inputToken := ScannerToken{
					tokenType: Undefined,
					text:  matchedString,
					lnnr:  scanner.lnnr,
					colnr: scanner.colnr,
				}
				// Call matching action to modify the raw token and append it to the token list
				outputToken := matcher.matchAction(matchedString, &inputToken)
				if outputToken != nil {
					tokenStream = append(tokenStream, *outputToken)
				}

				// Advance the scanner
				scanner.colnr += matchLength
				scanner.pos += matchLength
			}

			break
		}
	}

	return tokenStream
}
