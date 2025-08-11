package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

const (
	LIST_GAMES_UI = "listGames"
	SNAKE_GAME_UI = "snakeGame"
)

type model struct {
	listGames listGamesModel
	snakeGame snakeModel
	currentUI string
}

func NewModel(currUi string) model {
	lgm := InitModelListGames()
	snake := NewSnakeModel()

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
