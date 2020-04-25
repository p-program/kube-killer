package main

import (
	"github.com/p-program/kube-killer/cmd"
	_ "github.com/p-program/kube-killer/core" // init kubernetes config
)

func main() {

	cmd.Execute()
}
