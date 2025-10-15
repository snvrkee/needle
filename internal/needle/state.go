package needle

import (
	"fmt"
	"needle/internal/needle/evaluator"
	"needle/internal/needle/parser"
	"needle/internal/needle/scanner"
	"needle/internal/needle/token"
	"os"
	"path/filepath"
	"time"
)

type Needle struct {
	ev *evaluator.Evaluator
}

func New() *Needle {
	return &Needle{
		ev: evaluator.New(),
	}
}

func (n *Needle) Run(source []rune) error {
	s := scanner.New(source)
	script, errs := parser.New(s).Parse()
	if errs != nil {
		for _, err := range errs {
			fmt.Println("compile error: ", err)
		}
		return errs[0]
	}
	return n.ev.EvalScript(script)
}

func (n *Needle) RunFile(path string) error {
	abs, _ := filepath.Abs(path)
	dir := filepath.Dir(abs)
	n.ev.SetWorkDir(dir)
	return n.Run(readFile(path))
}

func (n *Needle) RunFile_debug(path string) error {
	abs, _ := filepath.Abs(path)
	dir := filepath.Dir(abs)
	n.ev.SetWorkDir(dir)
	fmt.Println("[file path] ->", abs)
	fmt.Println("[work dir] ->", dir)

	s := scanner.New(readFile(path))

	fmt.Println("== tokens ==")
	tokens := collectTokens(s)
	token.PrintTokens(tokens)
	s.Reset()

	script, errs := parser.New(s).Parse()

	fmt.Println("== ast ==")
	fmt.Println(script)

	if errs != nil {
		fmt.Println("== errors ==")
		for _, err := range errs {
			fmt.Println(err)
		}
		return errs[0]
	}

	start := time.Now()

	fmt.Println("== runtime ==")
	err := n.ev.EvalScript(script)

	fmt.Println("== result ==")
	fmt.Printf("program ends in %v\n", time.Since(start))

	return err
}

func collectTokens(tkz parser.Tokenizer) []*token.Token {
	tks := []*token.Token{}
	for {
		tk := tkz.NextToken()
		tks = append(tks, tk)
		if tk.Type == token.EOF {
			break
		}
	}
	return tks
}

func readFile(path string) []rune {
	b, _ := os.ReadFile(path)
	return []rune(string(b))
}
