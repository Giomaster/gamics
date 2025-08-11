/*
Copyright © 2025 Gio
*/
package cmd

import (
	"fmt"
	"gamics/tui"
	"os"
	"path"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	INITIAL_PATH      = "."
	EXTENSION_CONFIGS = "yaml"
)

var (
	appCfg    = viper.New()
	playerCfg = viper.New()
)

type user struct {
	name     string
	password string
}

var rootCmd = &cobra.Command{
	Use:           "gamics",
	Short:         "Classic games in Go",
	Long:          `Gamics is a collection of classic games implemented in Go.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		return loadGeneralConfig()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := checkIfUserIsLoggedIn()
		if err != nil {
			return err
		}

		_, err = tea.NewProgram(
			tui.NewModel(tui.SNAKE_GAME_UI),
			tea.WithInputTTY(),
			tea.WithFPS(60),
			tea.WithAltScreen(),
		).Run()
		if err != nil {
			return fmt.Errorf("error running the application: %w", err)
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadGeneralConfig() error {
	gamicsDir := path.Join(INITIAL_PATH, ".gamics")
	if _, err := os.Stat(gamicsDir); os.IsNotExist(err) {
		if err := os.Mkdir(gamicsDir, 0755); err != nil {
			return fmt.Errorf("could not create .gamics directory: %w", err)
		}
	}

	appCfg.SetConfigName("config")
	appCfg.SetConfigType(EXTENSION_CONFIGS)
	appCfg.AddConfigPath(gamicsDir)
	appCfg.SetDefault("logged-user", "")

	if err := appCfg.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}

		cfgPath := path.Join(gamicsDir, "config."+EXTENSION_CONFIGS)
		if err := appCfg.SafeWriteConfigAs(cfgPath); err != nil {
			return fmt.Errorf("could not create config file at %s: %w", cfgPath, err)
		}
	}

	return nil
}

func checkIfUserIsLoggedIn() error {
	loggedUser := appCfg.GetString("logged-user")
	if loggedUser == "" {
		return fmt.Errorf("please log in or register first")
	}

	return nil
}
