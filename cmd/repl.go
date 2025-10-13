package cmd

import (
	"bufio"
	"fmt"
	"needle/internal/needle"
	"os"
)

func RunRepl() error {
	state := needle.New()
	fmt.Println("Needle ver0.0.0")
	fmt.Println("exit using ctrl+c")
	for {
		fmt.Print("> ")
		r := bufio.NewReader(os.Stdin)
		str, _ := r.ReadString('\n')
		source := []rune(str[:len(str)-2])
		err := state.Run(source)
		if err != nil {
			fmt.Println(err)
		}
	}
}
