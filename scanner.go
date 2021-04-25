package main

import (
	"io"
	"regexp"
)

type ScannerTokenType int

const (
	KeywordQuery ScannerTokenType = iota
	Identifier
	Other
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

type ScannerMatcher struct {
	matchRegex  *regexp.Regexp
	matchAction func(*Scanner, string, *ScannerToken) *ScannerToken
}

// A List containing the regular expressions that define the tokens and a corresponding
// action to create such a token.
// The input is a pre-defined token. The function only has to change the corresponding values.
var scannerMatchList = []ScannerMatcher{
	{regexp.MustCompilePOSIX("^[ \t]+"),
		func(s *Scanner, input string, token *ScannerToken) *ScannerToken {
			s.colnr += len(input)
			return nil
		}},

	{regexp.MustCompilePOSIX("^\n"),
		func(s *Scanner, input string, token *ScannerToken) *ScannerToken {
			s.lnnr++
			s.colnr = 1
			return nil
		}},

	{regexp.MustCompilePOSIX("^[A-Za-z0-9]+"),
		func(s *Scanner, input string, token *ScannerToken) *ScannerToken {
			s.colnr += len(input)
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
			match := ele.matchRegex.FindStringIndex(scanner.buffer[scanner.pos:])
			if match == nil || match[0] != 0 {
				continue
			}

			// Regex matched. Create and initialize the token.
			matchLength := match[1]
			matchedString := scanner.buffer[scanner.pos : scanner.pos+matchLength]
			inputToken := ScannerToken{
				text:  matchedString,
				lnnr:  scanner.lnnr,
				colnr: scanner.colnr,
			}
			// Call matching action to modify the raw token and append it to the token list
			outputToken := ele.matchAction(&scanner, matchedString, &inputToken)
			if outputToken != nil {
				tokenStream = append(tokenStream, *outputToken)
			}

			// Advance the scanner
			scanner.pos += matchLength

			break
		}
	}

	return tokenStream
}
