package main

import (
	"flag"
	"fmt"
	"os"
	"pearl/evaluator"
	"pearl/lexer"
	"pearl/object"
	"pearl/parser"
	"pearl/repl"
)

func main() {
	// cli flags
	fileFlag := flag.String("f", "", "file to run")
	evalFlag := flag.String("e", "", "evaluate expression")
	checkFlag := flag.Bool("check", false, "just check syntax, dont run")
	versionFlag := flag.Bool("version", false, "print version")
	helpFlag := flag.Bool("help", false, "show help")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Pearl - A better Perl\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  pearl                  Start the REPL\n")
		fmt.Fprintf(os.Stderr, "  pearl -f <file>        Run a file\n")
		fmt.Fprintf(os.Stderr, "  pearl -e '<code>'      Evaluate code\n")
		fmt.Fprintf(os.Stderr, "  pearl <file>           Run a file (shorthand)\n")
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *helpFlag {
		flag.Usage()
		return
	}

	if *versionFlag {
		fmt.Println("Pearl 0.1.0")
		return
	}

	// handle -e flag
	if *evalFlag != "" {
		runCode(*evalFlag, *checkFlag)
		return
	}

	// handle file argument
	filename := *fileFlag
	if filename == "" && flag.NArg() > 0 {
		filename = flag.Arg(0)
	}

	if filename != "" {
		runFile(filename, *checkFlag)
		return
	}

	// no file, start repl
	repl.Start(os.Stdin, os.Stdout)
}

func runFile(filename string, checkOnly bool) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cant read file %s: %v\n", filename, err)
		os.Exit(1)
	}

	runCode(string(data), checkOnly)
}

func runCode(code string, checkOnly bool) {
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			fmt.Fprintln(os.Stderr, msg)
		}
		os.Exit(1)
	}

	if checkOnly {
		fmt.Println("syntax ok")
		return
	}

	env := object.NewEnvironment()
	result := evaluator.Eval(program, env)

	if result != nil && result.Type() == object.ERROR_OBJ {
		fmt.Fprintln(os.Stderr, result.Inspect())
		os.Exit(1)
	}
}
