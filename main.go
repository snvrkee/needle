package main

import (
	"log"
	"needle/cmd"
	"os"
)

func main() {
	var err error
	if len(os.Args) == 1 {
		err = cmd.RunRepl()
	} else {
		if len(os.Args) > 2 && os.Args[2] == "--debug" {
			err = cmd.RunFile_debug(os.Args[1])
		} else {
			err = cmd.RunFile(os.Args[1])
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}
