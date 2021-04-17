package main

import (
	"bufio"
	"io"
)

type ScannerTokenType int

const (
	KeywordQuery ScannerTokenType = iota
	Identifier
	Other
)

type ScannerToken struct {
	tokenType ScannerTokenType 
	orig string
}

func ScanningPhase (input io.Reader) (tokenStream []ScannerToken) {
	scanner := bufio.NewScanner(input)
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		tokenStream = append(tokenStream, ScannerToken{Other, scanner.Text()})
	}
	err := scanner.Err()
	if err != nil {
		panic(err)
	}

	return
}
