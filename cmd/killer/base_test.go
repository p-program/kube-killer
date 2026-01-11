package killer

import (
	"os"
	"testing"

	"github.com/p-program/kube-killer/core"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

// skipIfNoCluster skips the test if Kubernetes cluster is not available
func skipIfNoCluster(t testing.TB) {
	if core.GLOBAL_KUBERNETES_CONFIG == nil {
		t.Skip("Skipping test: Kubernetes cluster not available")
	}
}
