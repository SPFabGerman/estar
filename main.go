package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	inputfile := flag.String("i", "", "File to parse.")
	flag.Parse()

	if (*inputfile) == "" {
		log.Fatal("No input file specified!")
	}

	filehandler, err := os.Open(*inputfile)
	if err != nil {
		log.Fatal(err)
	}
	defer filehandler.Close()

	tokenStream := ScanningPhase(filehandler)
	fmt.Printf("%v", tokenStream)
}
