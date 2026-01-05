package cmd

import (
	"fmt"
	"os"

	"github.com/p-program/kube-killer/cmd/killer"
	"github.com/p-program/kube-killer/cmd/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile     string
	userLicense string

	rootCmd = &cobra.Command{
		Use:   "kube-killer",
		Short: "A tool helping you kill  kubernetes‘s resource",
		Long:  `Please don't use it for bad.Hhhhhh`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&userLicense, "license", "l", "", "name of license for the project")
	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
	viper.SetDefault("author", "https://github.com/p-program")
	viper.SetDefault("license", "Mulan PSL v2")
	bindCommand()

}

func bindCommand() {
	rootCmd.AddCommand(NewVersionCommand())
	rootCmd.AddCommand(NewFreezeCommand())
	rootCmd.AddCommand(server.NewServerCommand())
	kill := killer.NewKillCommand()
	rootCmd.AddCommand(kill)
	kill.AddCommand()

}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	}

	viper.AutomaticEnv()

	// if err := viper.ReadInConfig(); err == nil {
	// 	fmt.Println("Using config file:", viper.ConfigFileUsed())
	// }
}
