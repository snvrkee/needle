package needle

import (
	"fmt"
	"needle/internal/needle/evaluator"
	"needle/internal/needle/parser"
	"needle/internal/needle/scanner"
	"needle/internal/needle/token"
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
		return errs[0] //TODO
	}
	return n.ev.EvalScript(script)
}

func (n *Needle) Run_debug(source []rune) error {
	s := scanner.New(source)

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

	fmt.Println("== time ==")
	fmt.Printf("program end in %v\n", time.Since(start))
	
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
