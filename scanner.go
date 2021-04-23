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

type Scanner struct {
	buffer string
	pos    int
	size int
	lnnr   int
	colnr  int
}

type ScannerToken struct {
	tokenType   ScannerTokenType
	text        string
	lnnr, colnr int
}

type ScannerMatcher struct {
	matchRegex  *regexp.Regexp
	matchAction func(*Scanner, string, *ScannerToken) *ScannerToken
}

var scannerMatchList = []ScannerMatcher{
	{regexp.MustCompilePOSIX("^[ \t]+"), func(s *Scanner, input string, token *ScannerToken) *ScannerToken {
		s.colnr += len(input)
		return nil
	}},
	{regexp.MustCompilePOSIX("^\n"), func(s *Scanner, input string, token *ScannerToken) *ScannerToken {
		s.lnnr++
		s.colnr = 1
		return nil
	}},
	{regexp.MustCompilePOSIX("^[A-Za-z0-9]+"), func(s *Scanner, input string, token *ScannerToken) *ScannerToken {
		s.colnr += len(input)
		token.tokenType = Identifier
		return token
	}},
}

func ScanningPhase(input io.Reader) (tokenStream []ScannerToken) {
	inputString, err := io.ReadAll(input)
	if err != nil {
		panic(err)
	}
	var scanner Scanner = Scanner{string(inputString), 0, len(inputString) ,1, 1}

	for scanner.pos < scanner.size - 1 {
		for _, ele := range scannerMatchList {
			match := ele.matchRegex.FindStringIndex(scanner.buffer[scanner.pos:])
			if match == nil || match[0] != 0 {
				continue
			}
			matchLength := match[1]
			matchedString := scanner.buffer[scanner.pos:scanner.pos + matchLength]
			inputToken := ScannerToken{text: matchedString,
				lnnr: scanner.lnnr, colnr: scanner.colnr}
			outputToken := ele.matchAction(&scanner, matchedString, &inputToken)
			scanner.pos += matchLength
			if outputToken != nil {
				tokenStream = append(tokenStream, *outputToken)
			}
			break
		}
	}

	return
}

// func ScanningPhase(input io.Reader) (tokenStream []ScannerToken) {
// 	scanner := bufio.NewScanner(input)
// 	scanner.Split(bufio.ScanWords)

// 	for scanner.Scan() {
// 		tokenStream = append(tokenStream, ScannerToken{Other, scanner.Text()})
// 	}
// 	err := scanner.Err()
// 	if err != nil {
// 		panic(err)
// 	}

// 	return
// }
