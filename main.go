package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const applicationName = "tricktool"
const applicationVersion = "2.0.1"

func main() {

	var cmdUpgradeConfig = &cobra.Command{
		Use:   "upgrade-config [source_toml.conf]",
		Short: "Upgrades a Trickster 1.x config to a Trickster 2.0 config",
		Long: "\nupgrade-config will read a TOML source file from Trickster 1.x\n" +
			"and write a corresponding Trickster 2.0 YAML file",
		Args: cobra.MinimumNArgs(1),
		Run:  upgradeConfig,
	}

	var rootCmd = &cobra.Command{Use: applicationName, Version: applicationVersion}

	rootCmd.AddCommand(cmdUpgradeConfig)
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
