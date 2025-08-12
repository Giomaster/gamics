package tui

import (
	"fmt"
	"gamics/draw"
	"gamics/internal"
	"log"
	"math"
	"math/rand"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

// ----------------------------------------------------------------------------------
// Constants & globals
// ----------------------------------------------------------------------------------
const (
	SNAKE_GAME_LIGHT_BG = "#BCF0D7"
	SNAKE_GAME_DARK_BG  = "#04110A"

	gameTitle          = "Snake Game"
	foodTTL            = 10 * time.Second
	hungerResetSeconds = 30
	runTickEvery       = 100 * time.Millisecond
)

var (
	snakeCfg = viper.New()
	renderer = lipgloss.NewRenderer(os.Stdout)

	snakeBoxWarn = lipgloss.
			NewStyle().
			Padding(1, 2).
			Foreground(lipgloss.AdaptiveColor{Light: "#FFF", Dark: "#000"}).
			Background(lipgloss.AdaptiveColor{Light: "#333", Dark: "#AAA"})

	snakeBoxOption = lipgloss.
			NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#999", Dark: "#555"})

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
			Foreground(lipgloss.AdaptiveColor{Light: "#600", Dark: "#F00"}).
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
)

// ----------------------------------------------------------------------------------
// Messages (Bubble Tea)
// ----------------------------------------------------------------------------------
type tickRunSnakeGameMsg struct{}
type tickMsg struct{ gen int }
type tickBlinkFoodMsg struct{}
type tickHungerFoodMsg struct{ dieIn int }

// ----------------------------------------------------------------------------------
// Data types
// ----------------------------------------------------------------------------------
type Coordinates struct {
	X int `yaml:"x" mapstructure:"x"`
	Y int `yaml:"y" mapstructure:"y"`
}

type SnakePos struct {
	Position Coordinates `yaml:"position" mapstructure:"position"`
	Order    int         `yaml:"order"    mapstructure:"order"`
}

type Snake struct {
	Position          []SnakePos `yaml:"position"          mapstructure:"position"`
	Direction         string     `yaml:"direction"         mapstructure:"direction"`
	Speed             float64    `yaml:"speed"             mapstructure:"speed"`
	RenderedDirection string     `yaml:"renderedDirection" mapstructure:"renderedDirection"`
	DieByHungerIn     int        `yaml:"dieByHungerIn"     mapstructure:"dieByHungerIn"`
}

type Food struct {
	Position Coordinates `yaml:"position" mapstructure:"position"`
	Color    bool        `yaml:"color"    mapstructure:"color"`
	Expire   time.Time   `yaml:"expire"   mapstructure:"expire"`
}

type Option struct {
	Text   string              `yaml:"text"  mapstructure:"text"`
	Action func(m model) model `yaml:"-"` // função não serializa
}

type Options struct {
	Items  []Option `yaml:"items"  mapstructure:"items"`
	Cursor int      `yaml:"cursor" mapstructure:"cursor"`
}

type Game struct {
	Score   int     `yaml:"score"  mapstructure:"score"`
	Status  string  `yaml:"status" mapstructure:"status"`
	Options Options `yaml:"options" mapstructure:"options"`
}

type SnakeModel struct {
	Snake   Snake `yaml:"snake"    mapstructure:"snake"`
	Food    Food  `yaml:"food"     mapstructure:"food"`
	Game    Game  `yaml:"game"     mapstructure:"game"`
	TickGen int   `yaml:"tickGen"  mapstructure:"tickGen"`
}

// ----------------------------------------------------------------------------------
// Constructors
// ----------------------------------------------------------------------------------
func InitNewSnakeModel() SnakeModel {
	return SnakeModel{Game: Game{Status: "start"}}
}

func NewSnakeModel() SnakeModel {
	snakeBody := []SnakePos{
		{Position: Coordinates{X: 5, Y: 5}, Order: 0},
		{Position: Coordinates{X: 4, Y: 5}, Order: 1},
		{Position: Coordinates{X: 3, Y: 5}, Order: 2},
	}
	return SnakeModel{
		Game:  Game{Status: "running"},
		Food:  Food{Color: true, Position: Coordinates{X: -1, Y: -1}, Expire: time.Now().Add(foodTTL)},
		Snake: Snake{Position: snakeBody, Direction: "right", RenderedDirection: "right", Speed: 1, DieByHungerIn: hungerResetSeconds},
	}
}

func RestartSnakeModel(m model) SnakeModel {
	snakeBody := []SnakePos{
		{Position: Coordinates{X: 5, Y: 5}, Order: 0},
		{Position: Coordinates{X: 4, Y: 5}, Order: 1},
		{Position: Coordinates{X: 3, Y: 5}, Order: 2},
	}

	m.snakeGame.TickGen = 0
	m.snakeGame.Game = Game{Status: "running", Score: 0}
	m.snakeGame.Snake = Snake{Position: snakeBody, Direction: "right", Speed: 1, DieByHungerIn: hungerResetSeconds}
	m.snakeGame.Food = generateFood(m.snakeGame.Snake, m.snakeGame.Food, m.terminal)
	return m.snakeGame
}

// ----------------------------------------------------------------------------------
// Config helpers (Viper)
// ----------------------------------------------------------------------------------
func userDirOrDie() string {
	user, err := internal.GetUser()
	if err != nil {
		log.Fatal(err)
	}

	userDir := path.Join(".", ".gamics", user)
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		log.Fatalf("user directory does not exist, please register in first")
	}
	return userDir
}

func bindSnakeViper(userDir string) {
	snakeCfg.SetConfigName("snake")
	snakeCfg.SetConfigType("yaml")
	snakeCfg.AddConfigPath(userDir)
}

func createSessionGame() string {
	userDir := userDirOrDie()
	bindSnakeViper(userDir)

	snakeCfg.SetDefault("snake", Snake{
		Position: []SnakePos{
			{Position: Coordinates{X: 5, Y: 5}, Order: 0},
			{Position: Coordinates{X: 4, Y: 5}, Order: 1},
			{Position: Coordinates{X: 3, Y: 5}, Order: 2},
		},
		Direction:         "right",
		RenderedDirection: "right",
		Speed:             1,
		DieByHungerIn:     hungerResetSeconds,
	})
	snakeCfg.SetDefault("score", 0)
	snakeCfg.SetDefault("food", Food{Color: true, Position: Coordinates{X: -1, Y: -1}, Expire: time.Now().Add(foodTTL)})

	if err := snakeCfg.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalf("could not read config file: %v", err)
		}
		cfgPath := path.Join(userDir, "snake.yaml")
		if err := snakeCfg.SafeWriteConfigAs(cfgPath); err != nil {
			log.Fatalf("could not create config file at %s: %v", cfgPath, err)
		}
	}
	return userDir
}

func updateConfig(m SnakeModel) {
	snakeCfg.Set("snake", m.Snake)
	snakeCfg.Set("score", m.Game.Score)
	snakeCfg.Set("food", m.Food)
	if err := snakeCfg.WriteConfig(); err != nil {
		log.Fatalf("could not write config file: %v", err)
	}
}

func endSessionGame() {
	userDir := userDirOrDie()

	cfgPath := path.Join(userDir, "snake.yaml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return
	}

	// Persist high score in a separate file with explicit path
	profile := viper.New()
	profile.Set("snake-highscore", snakeCfg.GetInt("score"))
	if err := profile.WriteConfigAs(path.Join(userDir, "profile.yaml")); err != nil {
		log.Fatalf("could not write profile file: %v", err)
	}

	if err := os.Remove(cfgPath); err != nil {
		log.Fatalf("could not remove snake.yaml file: %v", err)
	}
}

func checkIfSnakeYamlFileExists() bool {
	userDir := userDirOrDie()
	_, err := os.Stat(path.Join(userDir, "snake.yaml"))
	return err == nil
}

func ContinueSnakeModel() SnakeModel {
	userDir := userDirOrDie()
	bindSnakeViper(userDir)

	if err := snakeCfg.ReadInConfig(); err != nil {
		log.Fatalf("could not read snake config file: %v", err)
	}

	var m SnakeModel
	m.Game.Score = snakeCfg.GetInt("score")
	if err := snakeCfg.UnmarshalKey("snake", &m.Snake); err != nil {
		log.Fatalf("could not unmarshal snake: %v", err)
	}
	if err := snakeCfg.UnmarshalKey("food", &m.Food); err != nil {
		log.Fatalf("could not unmarshal food: %v", err)
	}
	m.Game.Status = "running"
	return m
}

// ----------------------------------------------------------------------------------
// Update / View (Bubble Tea)
// ----------------------------------------------------------------------------------
func (m model) SnakeGameUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.snakeGame.Game.Status {
	case "running":
		return updateInRunningState(m, msg)
	case "lost":
		endSessionGame()
		return updateInLostState(m, msg)
	case "paused":
		return updateInPausedState(m, msg)
	case "start":
		return updateInStartState(m, msg)
	}
	return m, nil
}

func (m model) SnakeGameView() string {
	switch m.snakeGame.Game.Status {
	case "running":
		return viewInRunningState(m)
	case "lost":
		return viewInLostState(m)
	case "paused":
		return viewInPausedState(m)
	case "start":
		return viewInStartState(m)
	}
	return ""
}

func updateInStartState(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickStartSnakeGame:
		if !checkIfSnakeYamlFileExists() {
			createSessionGame()
			m.snakeGame = NewSnakeModel()
			m.snakeGame.Food = generateFood(m.snakeGame.Snake, m.snakeGame.Food, m.terminal)
			return m, tickRunSnakeGameCmd()
		}

		m.snakeGame.Game.Options = Options{Items: []Option{
			{Text: "Continue", Action: func(m model) model { m.snakeGame = ContinueSnakeModel(); return m }},
			{Text: "Start Over", Action: func(m model) model {
				createSessionGame()
				m.snakeGame = NewSnakeModel()
				m.snakeGame.Food = generateFood(m.snakeGame.Snake, m.snakeGame.Food, m.terminal)
				return m
			}},
		}, Cursor: 0}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up":
			if m.snakeGame.Game.Options.Cursor > 0 {
				m.snakeGame.Game.Options.Cursor--
			}
			return m, nil
		case "down":
			max := len(m.snakeGame.Game.Options.Items) - 1
			if m.snakeGame.Game.Options.Cursor < max {
				m.snakeGame.Game.Options.Cursor++
			}
			return m, nil
		case "enter":
			cur := m.snakeGame.Game.Options.Cursor
			if cur >= 0 && cur < len(m.snakeGame.Game.Options.Items) {
				m = m.snakeGame.Game.Options.Items[cur].Action(m)
				if m.snakeGame.Game.Status == "running" {
					return m, tickRunSnakeGameCmd()
				}
			}
			return m, nil
		}
	}
	return m, nil
}

func updateInRunningState(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickHungerFoodMsg:
		if msg.dieIn <= 0 {
			m.snakeGame.Game.Status = "lost"
			return m, nil
		}
		m.snakeGame.Snake.DieByHungerIn--
		return m, foodHungerTickCmd(1*time.Second, m.snakeGame.Snake.DieByHungerIn)

	case tickBlinkFoodMsg:
		rem := time.Until(m.snakeGame.Food.Expire)
		switch {
		case rem <= 0:
			m.snakeGame.Food = generateFood(m.snakeGame.Snake, m.snakeGame.Food, m.terminal)
			return m, foodBlinkTickCmd(500 * time.Millisecond)
		case rem <= 3*time.Second:
			m.snakeGame.Food.Color = !m.snakeGame.Food.Color
			return m, foodBlinkTickCmd(60 * time.Millisecond)
		case rem <= 5*time.Second:
			m.snakeGame.Food.Color = !m.snakeGame.Food.Color
			return m, foodBlinkTickCmd(120 * time.Millisecond)
		default:
			return m, foodBlinkTickCmd(500 * time.Millisecond)
		}

	case tickMsg:
		if msg.gen != m.snakeGame.TickGen {
			return m, nil
		}
		m.snakeGame.Snake.RenderedDirection = m.snakeGame.Snake.Direction
		m.snakeGame = updateSnakeSituation(m)
		if checkIfUserLose(m.snakeGame.Snake, m.terminal) {
			m.snakeGame.Game.Status = "lost"
			return m, nil
		}
		updateConfig(m.snakeGame)
		return m, tickCmd(m.snakeGame.TickGen, m.terminal, m.snakeGame.Snake)

	case tickRunSnakeGameMsg:
		m.snakeGame.TickGen++
		cmds := []tea.Cmd{
			foodBlinkTickCmd(500 * time.Millisecond),
			tickCmd(m.snakeGame.TickGen, m.terminal, m.snakeGame.Snake),
			foodHungerTickCmd(1*time.Second, m.snakeGame.Snake.DieByHungerIn),
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch msg.String() {
		case "p", "q", "ctrl+c":
			m.snakeGame.Game.Status = "paused"
			return m, nil
		case "up", "down", "left", "right":
			opp := map[string]string{"up": "down", "down": "up", "left": "right", "right": "left"}
			if m.snakeGame.Snake.RenderedDirection != opp[msg.String()] {
				m.snakeGame.Snake.Direction = msg.String()
			}
			return m, nil
		}
	}
	return m, nil
}

func updateInLostState(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminal.Width = msg.Width
		m.terminal.Height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.snakeGame.TickGen++
			m.snakeGame = RestartSnakeModel(m)
			cmds := []tea.Cmd{
				foodBlinkTickCmd(500 * time.Millisecond),
				foodHungerTickCmd(1*time.Second, hungerResetSeconds),
				tickCmd(m.snakeGame.TickGen, m.terminal, m.snakeGame.Snake),
			}
			return m, tea.Batch(cmds...)
		}
	}
	return m, nil
}

func updateInPausedState(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminal.Width = msg.Width
		m.terminal.Height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.snakeGame.Game.Status = "running"
			m.snakeGame.TickGen++
			cmds := []tea.Cmd{
				foodBlinkTickCmd(10 * time.Millisecond),
				foodHungerTickCmd(1*time.Second, m.snakeGame.Snake.DieByHungerIn),
				tickCmd(m.snakeGame.TickGen, m.terminal, m.snakeGame.Snake),
			}
			return m, tea.Batch(cmds...)
		}
	}
	return m, nil
}

func viewInStartState(m model) string {
	if !checkIfSnakeYamlFileExists() {
		l := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#060", Dark: "#0B0"}).Bold(true)
		return fullCenterBox(l, fmt.Sprintf("%s\n\n%s", draw.SNAKE_LOSE, "CARREGANDO..."), m.terminal)
	}

	var tw strings.Builder
	for i := 0; i < len(m.snakeGame.Game.Options.Items); i++ {
		txt := snakeBoxOption
		if i == m.snakeGame.Game.Options.Cursor {
			txt = txt.Foreground(lipgloss.AdaptiveColor{Light: "#0F0", Dark: "#060"}).Bold(true)
		}
		tw.WriteString(txt.Render(m.snakeGame.Game.Options.Items[i].Text) + "\n")
	}

	message := fmt.Sprintf("You are already in a game session. What do you want to do?\n\n%s", tw.String())
	return fullCenterBox(snakeBoxWarn, message, m.terminal)
}

func viewInRunningState(m model) string {
	w, h := fieldSize(m.terminal)
	foodBar := strings.Repeat("♥", max(m.snakeGame.Snake.DieByHungerIn, 0))

	title := horizontalCenterBox(snakeAppTitleStyle, gameTitle, m.terminal)
	stats := snakeAppStatsStyle.Render(fmt.Sprintf("Hunger: %ds: %s\nScore: %d", m.snakeGame.Snake.DieByHungerIn, foodBar, m.snakeGame.Game.Score))
	snakeBox := snakeAppStyle.Width(w).Height(h).Render(drawApp(m.snakeGame.Food, m.snakeGame.Snake, m.terminal))
	return fmt.Sprintf("%s\n\n%s\n\n%s", title, snakeBox, stats)
}

func viewInLostState(m model) string {
	w, h := fieldSize(m.terminal)
	title := horizontalCenterBox(snakeAppTitleStyle, gameTitle, m.terminal)
	stats := snakeAppStatsStyle.Render("You lost! Press 'q' to quit or 'r' to restart.")
	snakeBox := snakeAppStyle.Width(w).
		Foreground(lipgloss.Color("#F00")).
		Background(lipgloss.Color("#600")).
		BorderForeground(lipgloss.Color("#F00")).
		Height(h).
		Render(drawApp(m.snakeGame.Food, m.snakeGame.Snake, m.terminal))
	return fmt.Sprintf("%s\n\n%s\n\n%s", title, snakeBox, stats)
}

func viewInPausedState(m model) string {
	w, h := fieldSize(m.terminal)
	title := horizontalCenterBox(snakeAppTitleStyle, gameTitle, m.terminal)
	stats := snakeAppStatsStyle.Render("Game paused. Press 'q' to quit or 'r' to resume.")
	snakeBox := snakeAppStyle.Width(w).Height(h).Render(drawApp(m.snakeGame.Food, m.snakeGame.Snake, m.terminal))
	return fmt.Sprintf("%s\n\n%s\n\n%s", title, snakeBox, stats)
}

// ----------------------------------------------------------------------------------
// Ticking / Timing
// ----------------------------------------------------------------------------------
func tickCmd(gen int, t Terminal, s Snake) tea.Cmd {
	d := tickInterval(t, s)
	return tea.Tick(d, func(time.Time) tea.Msg { return tickMsg{gen: gen} })
}

func tickRunSnakeGameCmd() tea.Cmd {
	return tea.Tick(runTickEvery, func(time.Time) tea.Msg { return tickRunSnakeGameMsg{} })
}
func foodHungerTickCmd(d time.Duration, dieIn int) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return tickHungerFoodMsg{dieIn: dieIn} })
}
func foodBlinkTickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return tickBlinkFoodMsg{} })
}

func tickInterval(t Terminal, s Snake) time.Duration {
	if s.Speed <= 0 {
		s.Speed = 0.05
	}
	const (
		baseMs  = 200.0
		minMs   = 16.0
		maxMs   = 1000.0
		refDiag = 80.0
	)
	size := math.Hypot(float64(t.Width), float64(t.Height))
	ms := s.Speed * baseMs / (1.0 + size/refDiag)
	if ms < minMs {
		ms = minMs
	} else if ms > maxMs {
		ms = maxMs
	}
	return time.Duration(ms) * time.Millisecond
}

// ----------------------------------------------------------------------------------
// Game logic & rendering
// ----------------------------------------------------------------------------------
func checkIfUserLose(s Snake, t Terminal) bool {
	w, h := fieldSize(t)
	head := s.Position[0].Position
	if head.X < 0 || head.X >= w || head.Y < 0 || head.Y >= h {
		return true
	}
	for i := 1; i < len(s.Position); i++ {
		if head.X == s.Position[i].Position.X && head.Y == s.Position[i].Position.Y {
			return true
		}
	}
	return false
}

func updateSnakeSituation(m model) SnakeModel {
	if len(m.snakeGame.Snake.Position) == 0 {
		return m.snakeGame
	}

	head := m.snakeGame.Snake.Position[0].Position
	newHead := head
	switch m.snakeGame.Snake.RenderedDirection {
	case "up":
		newHead.Y--
	case "down":
		newHead.Y++
	case "left":
		newHead.X--
	case "right":
		newHead.X++
	default:
		return m.snakeGame
	}

	// If no food eaten, move tail
	if m.snakeGame.Food.Position.X != newHead.X || m.snakeGame.Food.Position.Y != newHead.Y {
		for i := len(m.snakeGame.Snake.Position) - 1; i >= 1; i-- {
			m.snakeGame.Snake.Position[i].Position = m.snakeGame.Snake.Position[i-1].Position
			m.snakeGame.Snake.Position[i].Order = i
		}
		m.snakeGame.Snake.Position[0].Position = newHead
		m.snakeGame.Snake.Position[0].Order = 0
		return m.snakeGame
	}

	// Ate food
	m.snakeGame.Snake.DieByHungerIn = hungerResetSeconds
	m.snakeGame.Game.Score++
	m.snakeGame.Snake.Speed -= 0.05
	m.snakeGame.Snake.Position = append([]SnakePos{{Position: newHead, Order: 0}}, m.snakeGame.Snake.Position...)
	m.snakeGame.Food = generateFood(m.snakeGame.Snake, m.snakeGame.Food, m.terminal)
	for i := len(m.snakeGame.Snake.Position) - 1; i >= 1; i-- {
		m.snakeGame.Snake.Position[i].Order = i
	}
	return m.snakeGame
}

func drawApp(f Food, s Snake, t Terminal) string {
	var sb strings.Builder

	colorHead, colorTail := "#0B321F", "#9BE8C3"
	colorFood := "#2FC67D"
	if renderer.HasDarkBackground() {
		colorHead, colorTail = "#49D491", "#0C321D"
		colorFood = "#F0D700"
	}

	colors := internal.InterpolateHexColors(colorHead, colorTail, len(s.Position))

	// Build positions WITHOUT mutating s.Position (avoid hidden side-effects)
	positions := make([]SnakePos, 0, len(s.Position)+1)
	positions = append(positions, s.Position...)
	positions = append(positions, SnakePos{Position: Coordinates{X: f.Position.X, Y: f.Position.Y}, Order: -1})

	sort.SliceStable(positions, func(i, j int) bool {
		if positions[i].Position.Y == positions[j].Position.Y {
			return positions[i].Position.X < positions[j].Position.X
		}
		return positions[i].Position.Y < positions[j].Position.Y
	})

	fw, fh := fieldSize(t)
	curY, curX := 0, 0
	for _, pos := range positions {
		if pos.Position.X < 0 || pos.Position.X >= fw || pos.Position.Y < 0 || pos.Position.Y >= fh {
			return ""
		}

		for curY < pos.Position.Y {
			sb.WriteString("\n")
			curX = 0
			curY++
		}
		for curX < pos.Position.X-1 {
			sb.WriteString(" ")
			curX++
		}

		part := lipgloss.NewStyle()
		if pos.Order == -1 {
			if f.Color {
				part = part.Background(lipgloss.Color(colorFood)).Foreground(lipgloss.Color(colorFood))
			}
			sb.WriteString(part.Render(" "))
			curX++
			continue
		}

		idx := max(pos.Order, 0)
		if idx >= len(colors) {
			idx = len(colors) - 1
		}
		sb.WriteString(part.Foreground(lipgloss.Color(colors[idx])).Background(lipgloss.Color(colors[idx])).Render("█"))
		curX++
	}
	return sb.String()
}

func generateFood(s Snake, f Food, t Terminal) Food {
	w, h := fieldSize(t)
	occupied := make(map[[2]int]bool, len(s.Position))
	for _, p := range s.Position {
		occupied[[2]int{p.Position.X, p.Position.Y}] = true
	}

	for tries := 0; tries < 1_000; tries++ { // safety cap
		x := rand.Intn(max(w, 1))
		y := rand.Intn(max(h, 1))
		if !occupied[[2]int{x, y}] {
			return Food{Color: true, Position: Coordinates{X: x, Y: y}, Expire: time.Now().Add(foodTTL)}
		}
	}
	// Fallback: keep previous food, extend TTL
	f.Expire = time.Now().Add(foodTTL)
	return f
}

// ----------------------------------------------------------------------------------
// Layout helpers
// ----------------------------------------------------------------------------------
func horizontalCenterBox(style lipgloss.Style, title string, t Terminal) string {
	block := style.Render(title)
	w := lipgloss.Width(block)
	marginLeft := (t.Width - w) / 2
	return style.Margin(0, marginLeft).Render(title)
}

func fullCenterBox(style lipgloss.Style, content string, t Terminal) string {
	block := style.Render(content)
	w := lipgloss.Width(block)
	marginLeft := (t.Width - w) / 2
	marginTop := (t.Height - lipgloss.Height(block)) / 2
	return style.Margin(marginTop, marginLeft).Render(content)
}

func fieldSize(t Terminal) (fw, fh int) { return t.Width - 2, t.Height - 10 }

// ----------------------------------------------------------------------------------
// Tick factories
// ----------------------------------------------------------------------------------
// NOTE: Terminal and model are defined elsewhere in the package.

// Small util (local max) — avoids pulling math for ints
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
