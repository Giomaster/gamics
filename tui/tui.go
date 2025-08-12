package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	LIST_GAMES_UI = "listGames"
	SNAKE_GAME_UI = "snakeGame"
)

var initTableGames = map[string]func() tea.Cmd{
	SNAKE_GAME_UI: tickStartSnakeGameCmd,
}

type tickStartSnakeGame struct{}

type model struct {
	listGames listGamesModel
	snakeGame SnakeModel
	terminal  Terminal
	currentUI string
}

type Terminal struct {
	Width  int
	Height int
}

func NewModel(currUi string) model {
	lgm := InitModelListGames()
	snake := InitNewSnakeModel()

	return model{
		listGames: lgm,
		snakeGame: snake,
		currentUI: currUi,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminal.Width = msg.Width
		m.terminal.Height = msg.Height
	}

	switch m.currentUI {
	case LIST_GAMES_UI:
		return m.ListGamesUpdate(msg)
	case SNAKE_GAME_UI:
		return m.SnakeGameUpdate(msg)
	}

	return nil, nil
}

func (m model) View() string {
	switch m.currentUI {
	case LIST_GAMES_UI:
		return m.ListGamesView()
	case SNAKE_GAME_UI:
		return m.SnakeGameView()
	}

	return ""
}

func tickStartSnakeGameCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickStartSnakeGame{}
	})
}
