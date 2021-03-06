package main

import (
	"os"

	"github.com/p-program/kube-killer/cmd"
	_ "github.com/p-program/kube-killer/core" // init kubernetes config
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	cmd.Execute()
}
