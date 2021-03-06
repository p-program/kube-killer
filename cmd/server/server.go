package server

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type KubeKillerServer struct {
}

func (s *KubeKillerServer) NewKubeKillerServer() *KubeKillerServer {

	server := KubeKillerServer{}
	return &server
}

func (s *KubeKillerServer) Run() {

	// config := rest.InClusterConfig()
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	<-ch
	fmt.Println("end")
}
