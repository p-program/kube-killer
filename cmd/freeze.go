package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	namespace string
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
			icebox, err := NewIcebox(namespace)
			if err != nil {
				log.Error().Err(err)
				return
			}
			icebox.Freeze(args[0], args[1])
		}}
	c.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "working namespace")
	return c
}
