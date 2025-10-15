package cmd

import (
	"needle/internal/needle"
)

func RunFile(filePath string) error {
	state := needle.New()
	return state.RunFile(filePath)
}

func RunFile_debug(filePath string) error {
	state := needle.New()
	return state.RunFile_debug(filePath)
}
