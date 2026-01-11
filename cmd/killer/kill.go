package killer

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/p-program/kube-killer/core"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

var (
	resources     []string
	namespace     string
	allNamespaces bool
	dryRun        bool
	interactive   bool
	mafia         bool
	half          bool
)

func init() {
	// flags := NewKillCommand.Flags()
	// flags.StringArrayVarP(&resources, "kill", "k", nil, "kill resource")
	// NewKillCommand.MarkFlagRequired("kill")

}

func NewKillCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "kill",
		Short: "Kill kubernetes's resource",
		Long:  `Delete unused Kubernetes resources. Supported resources: Pods, ConfigMaps, Secrets, Services, PVs, PVCs, Jobs, etc.`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				log.Error().Msgf("%s:len(args) < 1", "kill.go")
				return
			}
			if allNamespaces {
				namespace = ""
			}
			SelectKiller(args)
		}}
	c.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "working namespace")
	c.PersistentFlags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "If true, delete the targeted resources across all namespaces except kube-system")
	c.PersistentFlags().BoolVarP(&dryRun, "dryrun", "d", false, "dryRun")
	c.PersistentFlags().BoolVarP(&interactive, "interactive", "i", false, "If true, a prompt asks whether resources can be deleted")
	c.PersistentFlags().BoolVar(&mafia, "mafia", false, "If true, kill all resources (mafia mode)")
	c.PersistentFlags().BoolVar(&half, "half", false, "If true and mafia=true, randomly delete half of the resources")
	return c
}

func confirmDelete(resourceType, resourceName, namespace string) bool {
	if !interactive {
		return true
	}
	reader := bufio.NewReader(os.Stdin)
	nsStr := namespace
	if nsStr == "" {
		nsStr = "all namespaces"
	}
	fmt.Printf("? Are you sure to delete %s/%s in namespace %s? (y/N): ", resourceType, resourceName, nsStr)
	response, err := reader.ReadString('\n')
	if err != nil {
		log.Error().Err(err).Msg("Failed to read input")
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func SelectKiller(args []string) error {
	resourceType := args[0]

	// Get Kubernetes client for namespace helper
	clientset, err := getKubernetesClient()
	if err != nil {
		return err
	}

	switch resourceType {
	case "cm", "configmap":
		if allNamespaces {
			return processNamespaces(clientset, namespace, allNamespaces, func(ns string) error {
				k, err := NewConfigmapKiller(ns)
				if err != nil {
					return err
				}
				if dryRun {
					k.DryRun()
				}
				if mafia {
					k.BlackHand()
					if half {
						k.SetHalf()
					}
				}
				return k.Kill()
			})
		}
		k, err := NewConfigmapKiller(namespace)
		if err != nil {
			return err
		}
		if dryRun {
			k.DryRun()
		}
		if mafia {
			k.BlackHand()
			if half {
				k.SetHalf()
			}
		}
		return k.Kill()
	case "d", "deploy", "deployment":
		if allNamespaces {
			return processNamespaces(clientset, namespace, allNamespaces, func(ns string) error {
				k, err := NewDeploymentKiller(ns)
				if err != nil {
					return err
				}
				if dryRun {
					k.DryRun()
				}
				k.Kill()
				return nil
			})
		}
		k, err := NewDeploymentKiller(namespace)
		if err != nil {
			return err
		}
		if dryRun {
			k.DryRun()
		}
		k.Kill()
		return nil
	case "pv":
		// PV is cluster-scoped, no need for namespace iteration
		k, err := NewPVKiller()
		if err != nil {
			return err
		}
		if dryRun {
			k.DryRun()
		}
		if mafia {
			k.BlackHand()
			if half {
				k.SetHalf()
			}
		}
		return k.Kill()
	case "pvc":
		if allNamespaces {
			return processNamespaces(clientset, namespace, allNamespaces, func(ns string) error {
				k, err := NewPVCKiller(ns)
				if err != nil {
					return err
				}
				if dryRun {
					k.DryRun()
				}
				if mafia {
					k.BlackHand()
					if half {
						k.SetHalf()
					}
				}
				return k.Kill()
			})
		}
		k, err := NewPVCKiller(namespace)
		if err != nil {
			return err
		}
		if dryRun {
			k.DryRun()
		}
		if mafia {
			k.BlackHand()
			if half {
				k.SetHalf()
			}
		}
		return k.Kill()
	case "p", "po", "pod":
		if allNamespaces {
			return processNamespaces(clientset, namespace, allNamespaces, func(ns string) error {
				k, err := NewPodKiller(ns)
				if err != nil {
					return err
				}
				if dryRun {
					k.DryRun()
				}
				if mafia {
					k.BlackHand()
					if half {
						k.SetHalf()
					}
				}
				return k.Kill()
			})
		}
		k, err := NewPodKiller(namespace)
		if err != nil {
			return err
		}
		if dryRun {
			k.DryRun()
		}
		if mafia {
			k.BlackHand()
			if half {
				k.SetHalf()
			}
		}
		return k.Kill()
	case "s", "svc", "service":
		if allNamespaces {
			return processNamespaces(clientset, namespace, allNamespaces, func(ns string) error {
				k, err := NewServiceKiller(ns)
				if err != nil {
					return err
				}
				if dryRun {
					k.DryRun()
				}
				if mafia {
					k.BlackHand()
					if half {
						k.SetHalf()
					}
				}
				return k.Kill()
			})
		}
		k, err := NewServiceKiller(namespace)
		if err != nil {
			return err
		}
		if dryRun {
			k.DryRun()
		}
		if mafia {
			k.BlackHand()
			if half {
				k.SetHalf()
			}
		}
		return k.Kill()
	case "job", "jobs":
		if allNamespaces {
			return processNamespaces(clientset, namespace, allNamespaces, func(ns string) error {
				k, err := NewJobKiller(ns)
				if err != nil {
					return err
				}
				if dryRun {
					k.DryRun()
				}
				if mafia {
					k.BlackHand()
					if half {
						k.SetHalf()
					}
				}
				return k.Kill()
			})
		}
		k, err := NewJobKiller(namespace)
		if err != nil {
			return err
		}
		if dryRun {
			k.DryRun()
		}
		if mafia {
			k.BlackHand()
			if half {
				k.SetHalf()
			}
		}
		return k.Kill()
	case "secret", "secrets":
		if allNamespaces {
			return processNamespaces(clientset, namespace, allNamespaces, func(ns string) error {
				k, err := NewSecretKiller(ns)
				if err != nil {
					return err
				}
				if dryRun {
					k.DryRun()
				}
				if mafia {
					k.BlackHand()
					if half {
						k.SetHalf()
					}
				}
				return k.Kill()
			})
		}
		k, err := NewSecretKiller(namespace)
		if err != nil {
			return err
		}
		if dryRun {
			k.DryRun()
		}
		if mafia {
			k.BlackHand()
			if half {
				k.SetHalf()
			}
		}
		return k.Kill()
	case "me":
		log.Warn().Msg("!!!WARNING!!!: PLEASE DO NOT USE. It's an unpredictable command.")
		return nil
	case "n", "no", "node":
		if len(args) < 2 {
			log.Error().Msg("Node name is required")
			return fmt.Errorf("node name is required")
		}
		// Node is cluster-scoped, no need for namespace iteration
		k, err := NewNodeKiller(args[1])
		if err != nil {
			return err
		}
		if dryRun {
			k.DryRun()
		}
		return k.Kill()
	case "ns", "namespace":
		if allNamespaces {
			return processNamespaces(clientset, namespace, allNamespaces, func(ns string) error {
				k, err := NewNamespaceKiller(ns)
				if err != nil {
					return err
				}
				if dryRun {
					k.DryRun()
				}
				if mafia {
					k.BlackHand().Force()
				}
				return k.Kill()
			})
		}
		k, err := NewNamespaceKiller(namespace)
		if err != nil {
			return err
		}
		if dryRun {
			k.DryRun()
		}
		if mafia {
			k.BlackHand().Force()
		}
		return k.Kill()
	case "satan":
		log.Warn().Msg("!!!WARNING!!!: PLEASE DO NOT USE.")
		return nil
	case "cr", "customresource":
		if len(args) < 2 {
			log.Error().Msg("Group pattern is required for CR deletion (e.g., '*.example.com' or 'example.com')")
			return fmt.Errorf("group pattern is required for CR deletion")
		}
		groupPattern := args[1]
		// CRKiller handles all namespaces internally when namespace is empty
		targetNamespace := namespace
		if allNamespaces {
			targetNamespace = ""
		}
		k, err := NewCRKiller(groupPattern, targetNamespace)
		if err != nil {
			return err
		}
		if dryRun {
			k.DryRun()
		}
		return k.Kill()
	case "crd", "customresourcedefinition":
		if len(args) < 2 {
			log.Error().Msg("Group pattern is required for CRD deletion (e.g., '*.example.com' or 'example.com')")
			return fmt.Errorf("group pattern is required for CRD deletion")
		}
		groupPattern := args[1]
		// CRD is cluster-scoped, no need for namespace
		k, err := NewCRDKiller(groupPattern)
		if err != nil {
			return err
		}
		if dryRun {
			k.DryRun()
		}
		return k.Kill()
	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

func getKubernetesClient() (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
}

// ExecuteKill executes the kill operation with the given parameters
// This function is designed for use by kubectl plugins and other external callers
func ExecuteKill(resourceType, ns string, allNs, dry, interactiveMode bool) error {
	// Set package-level variables temporarily
	oldNamespace := namespace
	oldAllNamespaces := allNamespaces
	oldDryRun := dryRun
	oldInteractive := interactive

	namespace = ns
	allNamespaces = allNs
	dryRun = dry
	interactive = interactiveMode

	if allNamespaces {
		namespace = ""
	}

	// Execute the kill operation
	err := SelectKiller([]string{resourceType})

	// Restore old values (though they may not be needed if this is the only caller)
	namespace = oldNamespace
	allNamespaces = oldAllNamespaces
	dryRun = oldDryRun
	interactive = oldInteractive

	return err
}
