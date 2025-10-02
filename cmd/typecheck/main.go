package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/xingleixu/TG-Script/lexer"
	"github.com/xingleixu/TG-Script/parser"
	"github.com/xingleixu/TG-Script/types"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: typecheck <file.tg>")
		os.Exit(1)
	}

	filename := os.Args[1]
	
	// Read the source file
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Create lexer
	l := lexer.New(string(source))

	// Create parser
	p := parser.New(l)

	// Parse the program
	program := p.ParseProgram()

	// Check for parser errors
	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Printf("  %s\n", err)
		}
		os.Exit(1)
	}

	// Create type checker
	checker := types.NewTypeChecker()

	// Perform type checking
	errors := checker.Check(program)

	// Report results
	if len(errors) == 0 {
		fmt.Printf("✓ Type checking passed for %s\n", filename)
	} else {
		fmt.Printf("✗ Type checking failed for %s:\n", filename)
		for _, err := range errors {
			fmt.Printf("  %s\n", err.Error())
		}
		os.Exit(1)
	}
}