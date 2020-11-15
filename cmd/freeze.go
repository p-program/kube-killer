package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type Icebox struct {
}

func NewIcebox() *Icebox {
	i := Icebox{}

	return &i
}

// Freeze set deployment/statefulset/ zero size
func (i *Icebox) Freeze() {

}

func init() {

}

var (
	namespace string
	dryRun    bool
)

func NewFreezeCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "freeze",
		Short: "freeze kubernetes's resource to zero ",
		Long:  ``,
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				log.Error().Msgf("%s:len(args) < 2", "freeze.go")
				return
			}
		}}
	c.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "working namespace")
	c.PersistentFlags().BoolVarP(&dryRun, "dryrun", "d", false, "dryRun")
	return c
}
