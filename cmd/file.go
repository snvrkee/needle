package cmd

import (
	"needle/internal/needle"
	"os"
)

func RunFile(filePath string) error {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	source := []rune(string(bytes))
	state := needle.New()
	return state.Run(source)
}

func RunFile_debug(filePath string) error {
		bytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	source := []rune(string(bytes))
	state := needle.New()
	return state.Run_debug(source)
}
