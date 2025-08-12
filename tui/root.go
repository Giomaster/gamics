package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles ----------------------------------------------------------------------
var (
	listGamesAppStyle = lipgloss.NewStyle().Padding(1, 2)

	listGamesTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#25A065")).
				Padding(0, 1)

	listGamesStatusMessageStyle = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
					Render
)

// Data ------------------------------------------------------------------------
// item implements list.Item
// All games (implemented and coming soon) live in the same list.
// Not-implemented items have an empty gameId and are blocked on ENTER with a status message.
type item struct {
	title       string
	description string
	gameId      string // empty => coming soon
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }
func (i item) GameId() string      { return i.gameId }

// GameMeta models our catalog.
type GameMeta struct {
	Title       string
	Description string
	ID          string // empty means not implemented yet
}

func catalog() []GameMeta {
	return []GameMeta{
		{Title: "Snake", Description: "Guide the snake, eat food, grow and survive.", ID: SNAKE_GAME_UI},
		{Title: "Tic‑Tac‑Toe", Description: "3×3 noughts and crosses.", ID: ""},
		{Title: "Connect Four", Description: "Drop discs and make a line of four.", ID: ""},
		{Title: "Hangman", Description: "Guess the word, one letter at a time.", ID: ""},
		{Title: "2048", Description: "Slide tiles to reach 2048.", ID: ""},
		{Title: "Minesweeper", Description: "Uncover cells without hitting mines.", ID: ""},
		{Title: "Lights Out", Description: "Toggle lights to turn all off.", ID: ""},
		{Title: "15‑Puzzle", Description: "Slide tiles into order.", ID: ""},
		{Title: "8‑Puzzle", Description: "Smaller sliding puzzle variant.", ID: ""},
		{Title: "Sokoban", Description: "Push crates onto goals.", ID: ""},
		{Title: "Dots and Boxes", Description: "Draw lines, complete boxes.", ID: ""},
		{Title: "Nim", Description: "Take turns removing matches.", ID: ""},
		{Title: "21 Sticks", Description: "Variant of Nim to avoid the last stick.", ID: ""},
		{Title: "Rock‑Paper‑Scissors", Description: "Best of luck vs CPU.", ID: ""},
		{Title: "Higher or Lower", Description: "Guess if next number is higher.", ID: ""},
		{Title: "Guess the Number", Description: "Binary search your way to victory.", ID: ""},
		{Title: "Mastermind", Description: "Crack the color/code pattern.", ID: ""},
		{Title: "Bulls and Cows", Description: "Number‑guessing with feedback.", ID: ""},
		{Title: "Reversi (Othello)", Description: "Flip discs to dominate the board.", ID: ""},
		{Title: "Checkers", Description: "Draughts on an 8×8 board.", ID: ""},
		{Title: "Gomoku", Description: "Five‑in‑a‑row on a grid.", ID: ""},
		{Title: "Hex", Description: "Connect opposite sides.", ID: ""},
		{Title: "Battleship", Description: "Sink the enemy fleet.", ID: ""},
		{Title: "Peg Solitaire", Description: "Jump pegs to leave one.", ID: ""},
		{Title: "Memory (Concentration)", Description: "Match pairs from hidden cards.", ID: ""},
		{Title: "Simon", Description: "Repeat the sequence.", ID: ""},
		{Title: "Tower of Hanoi", Description: "Move disks with rules.", ID: ""},
		{Title: "Hitori", Description: "Logic puzzle on a grid.", ID: ""},
		{Title: "Nonogram (Picross)", Description: "Fill cells by clues to draw.", ID: ""},
		{Title: "Sudoku", Description: "Place digits 1‑9 without repeats.", ID: ""},
		{Title: "Wordle", Description: "Guess a 5‑letter word in 6 tries.", ID: ""},
		{Title: "Boggle", Description: "Find words in letter dice.", ID: ""},
		{Title: "Word Search", Description: "Locate hidden words in a grid.", ID: ""},
		{Title: "Anagrams", Description: "Rearrange letters to form words.", ID: ""},
		{Title: "Kakuro", Description: "Crossword‑like number sums.", ID: ""},
		{Title: "Minesweeper Tiny", Description: "5×5 quick variant.", ID: ""},
		{Title: "Treasure Hunt", Description: "Hot/Cold grid‑based search.", ID: ""},
		{Title: "Chomp", Description: "Take bites from a chocolate grid.", ID: ""},
		{Title: "Fox and Geese", Description: "Classic asymmetrical chase.", ID: ""},
		{Title: "Nine Men’s Morris", Description: "Form mills, remove pieces.", ID: ""},
		{Title: "Mancala", Description: "Sow stones, capture pits.", ID: ""},
		{Title: "Yahtzee", Description: "Roll dice, score categories.", ID: ""},
		{Title: "Pig (Dice)", Description: "Risk points rolling a die.", ID: ""},
		{Title: "Blackjack", Description: "Hit or stand to 21.", ID: ""},
		{Title: "Craps", Description: "Pass line dice betting.", ID: ""},
		{Title: "Slot Machine", Description: "Spin ASCII reels.", ID: ""},
		{Title: "Coin Toss", Description: "Heads or tails generator.", ID: ""},
		{Title: "Flappy Bird (ASCII)", Description: "One‑key obstacle dodging.", ID: ""},
		{Title: "Pong", Description: "Two paddles, one ball.", ID: ""},
		{Title: "Breakout", Description: "Break bricks with a paddle.", ID: ""},
		{Title: "Tetris", Description: "Fit falling tetrominoes.", ID: ""},
	}
}

// Key maps --------------------------------------------------------------------
type listKeyMap struct {
	toggleSpinner    key.Binding
	toggleTitleBar   key.Binding
	toggleStatusBar  key.Binding
	togglePagination key.Binding
	toggleHelpMenu   key.Binding
	insertItem       key.Binding // kept for demo parity
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		insertItem:       key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add item")),
		toggleSpinner:    key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "toggle spinner")),
		toggleTitleBar:   key.NewBinding(key.WithKeys("T"), key.WithHelp("T", "toggle title")),
		toggleStatusBar:  key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "toggle status")),
		togglePagination: key.NewBinding(key.WithKeys("P"), key.WithHelp("P", "toggle pagination")),
		toggleHelpMenu:   key.NewBinding(key.WithKeys("H"), key.WithHelp("H", "toggle help")),
	}
}

type delegateKeyMap struct {
	choose key.Binding
	remove key.Binding
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		choose: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "choose")),
		remove: key.NewBinding(key.WithKeys("x", "backspace"), key.WithHelp("x", "delete")),
	}
}

// Model -----------------------------------------------------------------------
type listGamesModel struct {
	list         list.Model
	keys         *listKeyMap
	delegateKeys *delegateKeyMap
}

// Init ------------------------------------------------------------------------
func InitModelListGames() listGamesModel {
	// Build list from the entire catalog (implemented + coming soon)
	items := make([]list.Item, 0)
	for _, g := range catalog() {
		items = append(items, item{title: g.Title, description: g.Description, gameId: g.ID})
	}

	delegate := newItemDelegate(newDelegateKeyMap())
	gameList := list.New(items, delegate, 0, 0)
	gameList.Title = "Games"
	gameList.Styles.Title = listGamesTitleStyle

	listKeys := newListKeyMap()
	gameList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.toggleSpinner,
			listKeys.insertItem,
			listKeys.toggleTitleBar,
			listKeys.toggleStatusBar,
			listKeys.togglePagination,
			listKeys.toggleHelpMenu,
		}
	}

	return listGamesModel{
		list:         gameList,
		keys:         listKeys,
		delegateKeys: newDelegateKeyMap(),
	}
}

// Update ----------------------------------------------------------------------
func (m model) ListGamesUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := listGamesAppStyle.GetFrameSize()
		m.listGames.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		// Don't match keys when filtering
		if m.listGames.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.listGames.keys.toggleSpinner):
			return m, m.listGames.list.ToggleSpinner()

		case key.Matches(msg, m.listGames.keys.toggleTitleBar):
			v := !m.listGames.list.ShowTitle()
			m.listGames.list.SetShowTitle(v)
			m.listGames.list.SetShowFilter(v)
			m.listGames.list.SetFilteringEnabled(v)
			return m, nil

		case key.Matches(msg, m.listGames.keys.toggleStatusBar):
			m.listGames.list.SetShowStatusBar(!m.listGames.list.ShowStatusBar())
			return m, nil

		case key.Matches(msg, m.listGames.keys.togglePagination):
			m.listGames.list.SetShowPagination(!m.listGames.list.ShowPagination())
			return m, nil

		case key.Matches(msg, m.listGames.keys.toggleHelpMenu):
			m.listGames.list.SetShowHelp(!m.listGames.list.ShowHelp())
			return m, nil

		case key.Matches(msg, m.listGames.delegateKeys.choose):
			if it, ok := m.listGames.list.SelectedItem().(item); ok {
				// If gameId is empty, it's a coming-soon game: show status and do nothing
				if it.GameId() == "" {
					msg := "\"" + it.Title() + "\" is coming soon! Stay tuned."
					return m, m.listGames.list.NewStatusMessage(listGamesStatusMessageStyle(msg))
				}
				// Otherwise, start the game normally
				m.currentUI = it.GameId()
				if startGame, ok := initTableGames[m.currentUI]; ok {
					cmds = append(cmds, startGame())
				}
				return m, tea.Batch(cmds...)
			}
			return m, nil
		}
	}

	// Delegate update
	newListModel, cmd := m.listGames.list.Update(msg)
	m.listGames.list = newListModel
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View ------------------------------------------------------------------------
func (m model) ListGamesView() string {
	return listGamesAppStyle.Render(m.listGames.list.View())
}

// Delegate --------------------------------------------------------------------
func newItemDelegate(keys *delegateKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	// Let the main Update handle ENTER to decide between implemented vs coming soon.
	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		var title string
		if i, ok := m.SelectedItem().(item); ok {
			title = i.Title()
		} else {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			// ENTER handled in ListGamesUpdate; do nothing here to avoid duplicate messages.
			case key.Matches(msg, keys.remove):
				idx := m.Index()
				m.RemoveItem(idx)
				return m.NewStatusMessage(listGamesStatusMessageStyle("Deleted " + title))
			}
		}
		return nil
	}

	help := []key.Binding{keys.choose, keys.remove}
	d.ShortHelpFunc = func() []key.Binding { return help }
	d.FullHelpFunc = func() [][]key.Binding { return [][]key.Binding{help} }
	return d
}

// Help maps (optional) --------------------------------------------------------
func (d delegateKeyMap) ShortHelp() []key.Binding  { return []key.Binding{d.choose, d.remove} }
func (d delegateKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{{d.choose, d.remove}} }
