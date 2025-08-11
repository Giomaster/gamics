package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	snakeAppTitleStyle = lipgloss.
				NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#BCF0D7", Dark: "#9BE8C3"}).
				Padding(0, 2).
				BorderStyle(lipgloss.Border{
			Top:         "-.-",
			Right:       "‡",
			Bottom:      ".-.",
			Left:        "‡",
			TopLeft:     "•",
			TopRight:    "•",
			BottomLeft:  "*",
			BottomRight: "*",
		}).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#135334", Dark: "#2FC67D"})

	snakeAppStatsStyle = lipgloss.
				NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#BCF0D7", Dark: "#9BE8C3"})

	snakeAppStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{
			Light: "#600",
			Dark:  "#F00",
		}).
		BorderStyle(lipgloss.Border{
			Left:        "▐",
			Right:       "▌",
			Top:         "▄",
			Bottom:      "▀",
			TopLeft:     "▟",
			TopRight:    "▙",
			BottomLeft:  "▜",
			BottomRight: "▛",
		}).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#135334", Dark: "#2FC67D"}).
		Background(lipgloss.AdaptiveColor{Light: "#BCF0D7", Dark: "#04110A"})
)

type coordinates struct {
	posX int
	posY int
}

type terminal struct {
	width  int
	height int
}

type snake struct {
	position  []snakePos
	direction string
	speed     float64
}

type snakePos struct {
	pos   coordinates
	order int
}

type food struct {
	position coordinates
	expire   time.Time
}

type game struct {
	score  int
	status string
}

type snakeModel struct {
	terminal terminal
	snake    snake
	food     food
	game     game
}

func NewSnakeModel() snakeModel {
	return snakeModel{}
}

func (m model) SnakeGameUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.snakeGame.terminal.width = msg.Width
		m.snakeGame.terminal.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) SnakeGameView() string {
	titleRender := horizontalCenterTitle(snakeAppTitleStyle, "Snake Game", m.snakeGame.terminal)
	stats := snakeAppStatsStyle.Render(fmt.Sprintf("Food: %s\nScore: %d", "█", m.snakeGame.game.score))
	snakeBox := snakeAppStyle.Width(m.snakeGame.terminal.width - 2).
		Height(m.snakeGame.terminal.height - 10).
		Render("█")

	return fmt.Sprintf("%s\n\n%s\n\n%s", titleRender, snakeBox, stats)
}

func horizontalCenterTitle(titleStyle lipgloss.Style, title string, t terminal) string {
	block := titleStyle.Render(title)

	w := lipgloss.Width(block)
	marginLeft := (t.width - w) / 2

	return titleStyle.
		Margin(0, marginLeft).
		Render(title)
}
