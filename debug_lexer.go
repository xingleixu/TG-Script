package main

import (
	"fmt"
	"github.com/xingleixu/TG-Script/lexer"
)

func main() {
	input := "function(x, y) { x + y; }"
	l := lexer.New(input)
	
	for {
		tok := l.NextToken()
		fmt.Printf("Type: %s, Literal: %s\n", tok.Type, tok.Literal)
		if tok.Type == lexer.EOF {
			break
		}
	}
}
