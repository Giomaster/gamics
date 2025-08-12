/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var (
	newPlayer user
)

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user",
	Long:  `Register a new user to the Gamics platform. This command allows you to create a new account with a username and password.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return checkPlayerAvailability()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return createNewPlayer()
	},
	Example:       `gamics register --username myuser --password mypass`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.AddCommand(registerCmd)
	registerCmd.Flags().StringVar(&newPlayer.name, "username", "", "Username for the new account")
	registerCmd.Flags().StringVar(&newPlayer.password, "password", "", "Password for the new account")
	registerCmd.MarkFlagRequired("username")
	registerCmd.MarkFlagRequired("password")
	registerCmd.MarkFlagsRequiredTogether("username", "password")
}

func checkPlayerAvailability() error {
	somePlayerDir := path.Join(INITIAL_PATH, ".gamics", newPlayer.name)
	if _, err := os.Stat(somePlayerDir); !os.IsNotExist(err) {
		return fmt.Errorf("username %s is already taken", newPlayer.name)
	}

	return nil
}

func createNewPlayer() error {
	playerDir := path.Join(INITIAL_PATH, ".gamics", newPlayer.name)
	if err := os.Mkdir(playerDir, 0755); err != nil {
		return fmt.Errorf("could not create player directory: %w", err)
	}

	playerCfg.AddConfigPath(playerDir)
	playerCfg.SetConfigName(newPlayer.name)
	playerCfg.SetConfigType(EXTENSION_CONFIGS)
	playerCfg.Set("password", newPlayer.password)
	playerCfg.Set("theme", "default")

	playerPath := path.Join(playerDir, newPlayer.name+"."+EXTENSION_CONFIGS)
	err := playerCfg.SafeWriteConfigAs(playerPath)
	if err != nil {
		return fmt.Errorf("error creating player config file: %w", err)
	}

	fmt.Printf("Player %s registered successfully!\n", newPlayer.name)
	return nil
}
