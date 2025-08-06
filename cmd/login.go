/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	chall user
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to the Gamics platform",
	Long: `Log in to the Gamics platform using your username and password.
This command allows you to authenticate and access your account.`,
	Example:       `gamics login --username myuser --password mypass`,
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initSession()
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVar(&chall.name, "username", "", "Username for the account")
	loginCmd.Flags().StringVar(&chall.password, "password", "", "Password for the account")
	loginCmd.MarkFlagRequired("username")
	loginCmd.MarkFlagRequired("password")
	loginCmd.MarkFlagsRequiredTogether("username", "password")
}

func initSession() error {
	err := checkAuthentication()
	if err != nil {
		return fmt.Errorf("error initializing session: %w", err)
	}

	appCfg.Set("logged-user", chall.name)
	if err := appCfg.WriteConfig(); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	fmt.Println("Login successful!")
	return nil
}

func checkAuthentication() error {
	if chall.name == "" {
		return fmt.Errorf("username is required")
	}

	if chall.password == "" {
		return fmt.Errorf("password is required")
	}

	gamicsPath := path.Join(INITIAL_PATH, ".gamics")
	if _, err := os.Stat(gamicsPath); os.IsNotExist(err) {
		return fmt.Errorf("gamics directory does not exist at %s", gamicsPath)
	}

	userConfigDir := path.Join(gamicsPath, chall.name)
	playerCfg.SetConfigName(chall.name)
	playerCfg.SetConfigType(EXTENSION_CONFIGS)
	playerCfg.AddConfigPath(userConfigDir)

	if err := playerCfg.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading user config file: %w", err)
		}

		return fmt.Errorf("username or password is incorrect")
	}

	if playerCfg.GetString("password") != chall.password {
		return fmt.Errorf("username or password is incorrect")
	}

	return nil
}
