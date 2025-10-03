package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/xingleixu/TG-Script/compiler"
	"github.com/xingleixu/TG-Script/lexer"
	"github.com/xingleixu/TG-Script/parser"
	"github.com/xingleixu/TG-Script/types"
	"github.com/xingleixu/TG-Script/vm"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	switch command {
	case "run":
		handleRun(os.Args[2:])
	case "compile":
		handleCompile(os.Args[2:])
	case "exec":
		handleExec(os.Args[2:])
	case "fmt":
		handleFormat(os.Args[2:])
	case "check":
		handleCheck(os.Args[2:])
	case "migrate":
		handleMigrate(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Printf("TG-Script %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`TG-Script %s - High-performance TypeScript-compatible scripting language

Usage:
  tg <command> [arguments]

Commands:
  run <file.tg>              Run TG-Script file
  compile <file.tg> [-o output]  Compile to bytecode
  exec <file.tgc>            Execute bytecode file
  fmt <file.tg>              Format code
  check <file.tg>            Check syntax and types
  migrate <file.ts>          Migrate from TypeScript
  version                    Show version information
  help                       Show help information

Examples:
  tg run hello.tg            # Run script
  tg compile hello.tg -o hello.tgc  # Compile script
  tg fmt hello.tg            # Format code
  tg migrate hello.ts        # Migrate TypeScript file

For more information visit: https://github.com/xingleixu/TG-Script
`, version)
}

func handleRun(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Please specify a .tg file to run")
		os.Exit(1)
	}
	
	filename := args[0]
	
	// Check file extension
	if !strings.HasSuffix(filename, ".tg") {
		fmt.Printf("Error: File must have .tg extension, got: %s\n", filename)
		os.Exit(1)
	}
	
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("Error: File not found: %s\n", filename)
		os.Exit(1)
	}
	
	// Read source code
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		os.Exit(1)
	}
	
	// Execute the script
	if err := executeScript(string(source), filename); err != nil {
		fmt.Printf("Error executing script: %v\n", err)
		os.Exit(1)
	}
}

func executeScript(source, filename string) error {
	// Lexical analysis
	l := lexer.New(source)
	
	// Parse
	p := parser.New(l)
	program := p.ParseProgram()
	
	// Check for parser errors
	if errors := p.Errors(); len(errors) > 0 {
		fmt.Printf("Parser errors in %s:\n", filename)
		for _, err := range errors {
			fmt.Printf("  %s\n", err)
		}
		return fmt.Errorf("parsing failed")
	}
	
	// Type checking
	checker := types.NewTypeChecker()
	typeErrors := checker.Check(program)
	
	// Check for type errors
	if len(typeErrors) > 0 {
		fmt.Printf("Type errors in %s:\n", filename)
		for _, err := range typeErrors {
			fmt.Printf("  %s\n", err.Error())
		}
		return fmt.Errorf("type checking failed")
	}
	
	// Compile
	function, err := compiler.CompileFunction(program)
	if err != nil {
		return fmt.Errorf("compilation failed: %v", err)
	}
	
	// Execute
	machine := vm.NewVM()
	closure := vm.NewClosure(function)
	result, err := machine.Execute(closure, []vm.Value{})
	if err != nil {
		return fmt.Errorf("execution failed: %v", err)
	}
	
	// Print result if it's not nil
	if !result.IsNil() {
		fmt.Printf("Result: %v\n", result)
	}
	
	return nil
}

func checkScript(source, filename string) error {
	// Lexical analysis
	l := lexer.New(source)
	
	// Parse
	p := parser.New(l)
	program := p.ParseProgram()
	
	// Check for parser errors
	if errors := p.Errors(); len(errors) > 0 {
		fmt.Printf("Parser errors in %s:\n", filename)
		for _, err := range errors {
			fmt.Printf("  %s\n", err)
		}
		return fmt.Errorf("parsing failed")
	}
	

	
	// Type checking
	checker := types.NewTypeChecker()
	typeErrors := checker.Check(program)
	
	// Check for type errors
	if len(typeErrors) > 0 {
		fmt.Printf("Type errors in %s:\n", filename)
		for _, err := range typeErrors {
			fmt.Printf("  %s\n", err.Error())
		}
		return fmt.Errorf("type checking failed")
	}
	
	return nil
}

func handleCompile(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Please specify a .tg file to compile")
		os.Exit(1)
	}
	
	filename := args[0]
	output := filename[:len(filename)-3] + ".tgc" // Default output filename
	
	// Parse -o argument
	for i, arg := range args {
		if arg == "-o" && i+1 < len(args) {
			output = args[i+1]
			break
		}
	}
	
	fmt.Printf("Compiling TG-Script file: %s -> %s\n", filename, output)
	// TODO: Implement compilation logic
	fmt.Println("Note: Compile functionality not yet implemented")
}

func handleExec(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Please specify a .tgc file to execute")
		os.Exit(1)
	}
	
	filename := args[0]
	fmt.Printf("Executing bytecode file: %s\n", filename)
	// TODO: Implement bytecode execution logic
	fmt.Println("Note: Execute functionality not yet implemented")
}

func handleFormat(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Please specify a .tg file to format")
		os.Exit(1)
	}
	
	filename := args[0]
	fmt.Printf("Formatting TG-Script file: %s\n", filename)
	// TODO: Implement code formatting logic
	fmt.Println("Note: Format functionality not yet implemented")
}

func handleCheck(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Please specify a .tg file to check")
		os.Exit(1)
	}
	
	filename := args[0]
	
	// Check file extension
	if !strings.HasSuffix(filename, ".tg") {
		fmt.Printf("Error: File must have .tg extension, got: %s\n", filename)
		os.Exit(1)
	}
	
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("Error: File not found: %s\n", filename)
		os.Exit(1)
	}
	
	// Read source code
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		os.Exit(1)
	}
	
	// Perform syntax and type checking
	if err := checkScript(string(source), filename); err != nil {
		fmt.Printf("Check failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("✓ Check passed for %s\n", filename)
}

func handleMigrate(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Please specify a .ts file to migrate")
		os.Exit(1)
	}
	
	filename := args[0]
	fmt.Printf("Migrating TypeScript file: %s\n", filename)
	// TODO: Implement TypeScript migration logic
	fmt.Println("Note: Migration functionality not yet implemented")
}