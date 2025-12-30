package repl

import (
	"bufio"
	"fmt"
	"io"
	"pearl/evaluator"
	"pearl/lexer"
	"pearl/object"
	"pearl/parser"
	"strings"
)

const PROMPT = "pearl> "

const LOGO = `
                      _ 
  _ __   ___  __ _ _ __| |
 | '_ \ / _ \/ _' | '__| |
 | |_) |  __/ (_| | |  | |
 | .__/ \___|\__,_|_|  |_|
 |_|   
`

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	fmt.Fprint(out, LOGO)
	fmt.Fprintln(out, "Pearl - A better Perl")
	fmt.Fprintln(out, "Type 'exit' or Ctrl+D to quit")
	fmt.Fprintln(out)

	var multilineBuffer strings.Builder
	inMultiline := false

	for {
		if inMultiline {
			fmt.Fprint(out, "...    ")
		} else {
			fmt.Fprint(out, PROMPT)
		}

		scanned := scanner.Scan()
		if !scanned {
			fmt.Fprintln(out, "\nbye!")
			return
		}

		line := scanner.Text()

		// check for exit
		if !inMultiline && (line == "exit" || line == "quit") {
			fmt.Fprintln(out, "bye!")
			return
		}

		// handle multiline input
		if inMultiline {
			multilineBuffer.WriteString("\n")
			multilineBuffer.WriteString(line)

			// check if braces are balanced
			if isBalanced(multilineBuffer.String()) {
				line = multilineBuffer.String()
				multilineBuffer.Reset()
				inMultiline = false
			} else {
				continue
			}
		} else {
			// check if we need to go multiline
			if !isBalanced(line) {
				multilineBuffer.WriteString(line)
				inMultiline = true
				continue
			}
		}

		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			// dont print null for statements that dont return anything interesting
			if evaluated.Type() != object.NULL_OBJ {
				fmt.Fprintln(out, evaluated.Inspect())
			}
		}
	}
}

func isBalanced(s string) bool {
	count := 0
	inString := false
	var stringChar byte

	for i := 0; i < len(s); i++ {
		ch := s[i]

		// handle strings
		if (ch == '"' || ch == '\'') && (i == 0 || s[i-1] != '\\') {
			if !inString {
				inString = true
				stringChar = ch
			} else if ch == stringChar {
				inString = false
			}
			continue
		}

		if inString {
			continue
		}

		// count braces
		if ch == '{' || ch == '(' || ch == '[' {
			count++
		} else if ch == '}' || ch == ')' || ch == ']' {
			count--
		}
	}

	return count <= 0 && !inString
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		fmt.Fprintln(out, "  "+msg)
	}
}
