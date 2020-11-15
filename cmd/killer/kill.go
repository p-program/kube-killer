package killer

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	resources []string
	namespace string
	dryRun    bool
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
		Long:  ``,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				log.Error().Msgf("%s:len(args) < 1", "kill.go")
				return
			}
			SelectKiller(args)
		}}
	c.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "working namespace")
	c.PersistentFlags().BoolVarP(&dryRun, "dryrun", "d", false, "dryRun")
	return c
}

func SelectKiller(args []string) error {
	resourceType := args[0]
	switch resourceType {
	case "cm", "configmap":
		break
	case "d", "deploy":
		k, err := NewDeploymentKiller(namespace)
		if err != nil {
			return err
		}
		if dryRun {
			k.DryRun()
		}
		k.Kill()
		break
	case "pv":
		k, err := NewPVKiller()
		if err != nil {
			return err
		}
		if dryRun {
			k.DryRun()
		}
		k.Kill()
		break
	case "pvc":
		k, err := NewPVCKiller(namespace)
		if err != nil {
			return err
		}
		if dryRun {
			k.DryRun()
		}
		k.Kill()
		break
	case "p", "po", "pod":
		k, err := NewPodKiller(namespace)
		if err != nil {
			return err
		}
		if dryRun {
			k.DryRun()
		}
		k.Kill()
		break
	case "s", "svc", "service":
		k, err := NewServiceKiller(namespace)
		if err != nil {
			return err
		}
		if dryRun {
			k.DryRun()
		}
		k.Kill()
		break
	case "me":
		break
	case "n", "no", "node":
		break
	case "ns", "namespace":
		break
	case "satan":
		break
	case "secret":
		break
	case "":
		break
	}
	return nil
}
