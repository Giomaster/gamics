package tui

import (
	"fmt"
	"gamics/internal"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	SNAKE_GAME_LIGHT_BG = "#BCF0D7"
	SNAKE_GAME_DARK_BG  = "#04110A"
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
		Background(lipgloss.AdaptiveColor{Light: SNAKE_GAME_LIGHT_BG, Dark: SNAKE_GAME_DARK_BG})
	renderer = lipgloss.NewRenderer(os.Stdout)
)

type tickMsg struct{ gen int }
type tickFoodMsg struct{}

type coordinates struct {
	posX int
	posY int
}

type terminal struct {
	width  int
	height int
}

type snake struct {
	position          []snakePos
	direction         string
	speed             float64
	renderedDirection string
}

type snakePos struct {
	pos   coordinates
	order int
}

type food struct {
	position coordinates
	color    bool
	expire   time.Time
}

type option struct {
	text   string
	action func(m *model) model
}

type options struct {
	items  []option
	cursor int
}

type game struct {
	score   int
	status  string
	options options
}

type snakeModel struct {
	terminal terminal
	snake    snake
	food     food
	game     game
	tickGen  int
}

func NewSnakeModel() snakeModel {
	snakeBody := []snakePos{
		{pos: coordinates{posX: 5, posY: 5}, order: 0},
		{pos: coordinates{posX: 4, posY: 5}, order: 1},
		{pos: coordinates{posX: 3, posY: 5}, order: 2},
	}

	return snakeModel{
		game: game{
			status: "running",
		},
		food: food{
			color: true,
			position: coordinates{
				posX: -1,
				posY: -1,
			},
		},
		snake: snake{
			position:          snakeBody,
			direction:         "right",
			renderedDirection: "right",
			speed:             1,
		},
	}
}

func RestartSnakeModel(m model) snakeModel {
	snakeBody := []snakePos{
		{pos: coordinates{posX: 5, posY: 5}, order: 0},
		{pos: coordinates{posX: 4, posY: 5}, order: 1},
		{pos: coordinates{posX: 3, posY: 5}, order: 2},
	}

	m.snakeGame.tickGen = 0
	m.snakeGame.game = game{
		status: "running",
		score:  0,
	}

	m.snakeGame.snake = snake{
		position:  snakeBody,
		direction: "right",
		speed:     1,
	}

	m.snakeGame.food = generateFood(m.snakeGame.snake, m.snakeGame.food, m.snakeGame.terminal)
	return m.snakeGame
}

func (m model) SnakeGameUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.snakeGame.game.status {
	case "running":
		return updateInRunningState(m, msg)
	case "lost":
		return updateInLostState(m, msg)
	}

	return m, nil
}

func (m model) SnakeGameView() string {
	switch m.snakeGame.game.status {
	case "running":
		return viewInRunningState(m)
	case "lost":
		return viewInLostState(m)
	}

	return ""
}

func updateInRunningState(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickFoodMsg:
		rem := time.Until(m.snakeGame.food.expire)
		switch {
		case rem <= 0:
			m.snakeGame.food = generateFood(m.snakeGame.snake, m.snakeGame.food, m.snakeGame.terminal)
			return m, foodTickCmd(500 * time.Millisecond)

		case rem <= 3*time.Second:
			m.snakeGame.food.color = !m.snakeGame.food.color
			return m, foodTickCmd(60 * time.Millisecond)

		case rem <= 5*time.Second:
			m.snakeGame.food.color = !m.snakeGame.food.color
			return m, foodTickCmd(120 * time.Millisecond)

		default:
			return m, foodTickCmd(500 * time.Millisecond)
		}

	case tickMsg:
		if msg.gen != m.snakeGame.tickGen {
			return m, nil
		}

		m.snakeGame.snake.renderedDirection = m.snakeGame.snake.direction
		m.snakeGame = updateSnakePosition(m.snakeGame)
		if checkIfUserLose(m.snakeGame.snake, m.snakeGame.terminal) {
			m.snakeGame.game.status = "lost"
			return m, nil
		}

		return m, tickCmd(m.snakeGame.tickGen, m.snakeGame.terminal, m.snakeGame.snake)

	case tea.WindowSizeMsg:
		m.snakeGame.terminal.width = msg.Width
		m.snakeGame.terminal.height = msg.Height

		m.snakeGame.tickGen++
		m.snakeGame.food = generateFood(m.snakeGame.snake, m.snakeGame.food, m.snakeGame.terminal)
		cmds := []tea.Cmd{
			foodTickCmd(500 * time.Millisecond),
			tickCmd(m.snakeGame.tickGen, m.snakeGame.terminal, m.snakeGame.snake),
		}

		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "down", "left", "right":
			if (msg.String() == "up" && m.snakeGame.snake.renderedDirection != "down") ||
				(msg.String() == "down" && m.snakeGame.snake.renderedDirection != "up") ||
				(msg.String() == "left" && m.snakeGame.snake.renderedDirection != "right") ||
				(msg.String() == "right" && m.snakeGame.snake.renderedDirection != "left") {
				m.snakeGame.snake.direction = msg.String()
			}
			return m, nil
		}
	}

	return m, nil
}

func updateInLostState(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.snakeGame.terminal.width = msg.Width
		m.snakeGame.terminal.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.snakeGame = RestartSnakeModel(m)
			return m, tickCmd(m.snakeGame.tickGen, m.snakeGame.terminal, m.snakeGame.snake)
		}
	}

	return m, nil
}

func viewInRunningState(m model) string {
	w, h := fieldSize(m.snakeGame.terminal)
	titleRender := horizontalCenterTitle(snakeAppTitleStyle, "Snake Game", m.snakeGame.terminal)
	stats := snakeAppStatsStyle.Render(fmt.Sprintf("Food: %s\nScore: %d", "A", m.snakeGame.game.score))
	snakeBox := snakeAppStyle.Width(w).
		Height(h).
		Render(drawApp(m.snakeGame.food, m.snakeGame.snake, m.snakeGame.terminal))

	return fmt.Sprintf("%s\n\n%s\n\n%s", titleRender, snakeBox, stats)
}

func viewInLostState(m model) string {
	w, h := fieldSize(m.snakeGame.terminal)
	titleRender := horizontalCenterTitle(snakeAppTitleStyle, "Snake Game", m.snakeGame.terminal)
	stats := snakeAppStatsStyle.Render("You lost! Press 'q' to quit or 'r' to restart.")
	snakeBox := snakeAppStyle.Width(w).
		Foreground(lipgloss.Color("#F00")).
		Background(lipgloss.Color("#600")).
		BorderForeground(lipgloss.Color("#F00")).
		Height(h).
		Render(drawApp(m.snakeGame.food, m.snakeGame.snake, m.snakeGame.terminal))

	return fmt.Sprintf("%s\n\n%s\n\n%s", titleRender, snakeBox, stats)
}

func tickCmd(gen int, t terminal, s snake) tea.Cmd {
	d := tickInterval(t, s)
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg{gen: gen}
	})
}

func foodTickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickFoodMsg{}
	})
}

func tickInterval(t terminal, s snake) time.Duration {
	if s.speed <= 0 {
		s.speed = 0.05
	}
	const (
		baseMs  = 200.0
		minMs   = 16.0
		maxMs   = 1000.0
		refDiag = 80.0
	)

	size := math.Hypot(float64(t.width), float64(t.height))
	ms := s.speed * baseMs / (1.0 + size/refDiag)

	if ms < minMs {
		ms = minMs
	} else if ms > maxMs {
		ms = maxMs
	}
	return time.Duration(ms) * time.Millisecond
}

func checkIfUserLose(s snake, t terminal) bool {
	w, h := fieldSize(t)
	head := s.position[0].pos
	if head.posX < 0 ||
		head.posX >= w ||
		head.posY < 0 ||
		head.posY >= h {
		return true
	}

	for i := 1; i < len(s.position); i++ {
		if head.posX == s.position[i].pos.posX && head.posY == s.position[i].pos.posY {
			return true
		}
	}

	return false
}

func updateSnakePosition(m snakeModel) snakeModel {
	if len(m.snake.position) == 0 {
		return m
	}

	head := m.snake.position[0].pos
	newHead := head
	switch m.snake.renderedDirection {
	case "up":
		newHead.posY--
	case "down":
		newHead.posY++
	case "left":
		newHead.posX--
	case "right":
		newHead.posX++
	default:
		return m
	}

	if m.food.position.posX != newHead.posX || m.food.position.posY != newHead.posY {
		for i := len(m.snake.position) - 1; i >= 1; i-- {
			m.snake.position[i].pos = m.snake.position[i-1].pos
			m.snake.position[i].order = i
		}

		m.snake.position[0].pos = newHead
		m.snake.position[0].order = 0
		return m
	}

	m.game.score++
	m.snake.speed -= 0.05
	m.snake.position = append([]snakePos{{pos: newHead, order: 0}}, m.snake.position...)
	m.food = generateFood(m.snake, m.food, m.terminal)

	for i := len(m.snake.position) - 1; i >= 1; i-- {
		m.snake.position[i].order = i
	}

	return m
}

func drawApp(f food, s snake, t terminal) string {
	var sb strings.Builder

	colorHead, colorTail := "#0B321F", "#9BE8C3"
	colorFood := "#2FC67D"
	if renderer.HasDarkBackground() {
		colorHead, colorTail = "#49D491", "#0C321D"
		colorFood = "#F0D700"
	}

	colors := internal.InterpolateHexColors(colorHead, colorTail, len(s.position))

	s.position = append(s.position, snakePos{
		pos:   coordinates{posX: f.position.posX, posY: f.position.posY},
		order: -1,
	})

	positions := append([]snakePos(nil), s.position...)
	sort.SliceStable(positions, func(i, j int) bool {
		if positions[i].pos.posY == positions[j].pos.posY {
			return positions[i].pos.posX < positions[j].pos.posX
		}
		return positions[i].pos.posY < positions[j].pos.posY
	})

	fw, fh := fieldSize(t)
	curY, curX := 0, 0
	for _, pos := range positions {
		part := lipgloss.NewStyle()
		if pos.pos.posX < 0 || pos.pos.posX >= fw || pos.pos.posY < 0 || pos.pos.posY >= fh {
			return ""
		}

		for curY < pos.pos.posY {
			sb.WriteString("\n")
			curX = 0
			curY++
		}
		for curX < pos.pos.posX {
			curX++
			if curX == pos.pos.posX {
				break
			}

			sb.WriteString(" ")
		}

		if pos.order == -1 {
			if f.color {
				part = part.
					Background(lipgloss.Color(colorFood)).
					Foreground(lipgloss.Color(colorFood))
			}

			food := part.Render(" ")
			sb.WriteString(food)
			continue
		}

		idx := max(pos.order, 0)
		if idx >= len(colors) {
			idx = len(colors) - 1
		}

		sb.WriteString(
			part.Foreground(lipgloss.Color(colors[idx])).
				Background(lipgloss.Color(colors[idx])).
				Render("█"),
		)
	}
	return sb.String()
}

func generateFood(s snake, f food, t terminal) food {
	w, h := fieldSize(t)

	possibleFoodPosX := rand.Intn(w - 1)
	possibleFoodPosY := rand.Intn(h - 1)
	for _, pos := range s.position {
		if pos.pos.posX == possibleFoodPosX && pos.pos.posY == possibleFoodPosY {
			return generateFood(s, f, t)
		}
	}

	return food{
		color: true,
		position: coordinates{
			posX: rand.Intn(w - 1),
			posY: rand.Intn(h - 1),
		},
		expire: time.Now().Add(time.Second * 10),
	}
}

func horizontalCenterTitle(titleStyle lipgloss.Style, title string, t terminal) string {
	block := titleStyle.Render(title)

	w := lipgloss.Width(block)
	marginLeft := (t.width - w) / 2

	return titleStyle.
		Margin(0, marginLeft).
		Render(title)
}

func fieldSize(t terminal) (fw, fh int) {
	return t.width - 2, t.height - 10
}
