package tui

import (
	"math/rand/v2"
	"sync"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

type item struct {
	title       string
	description string
	gameId      string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }
func (i item) GameId() string      { return i.gameId }

type listKeyMap struct {
	toggleSpinner    key.Binding
	toggleTitleBar   key.Binding
	toggleStatusBar  key.Binding
	togglePagination key.Binding
	toggleHelpMenu   key.Binding
	insertItem       key.Binding
}

type listGamesModel struct {
	list          list.Model
	itemGenerator *randomItem
	keys          *listKeyMap
	delegateKeys  *delegateKeyMap
}

type delegateKeyMap struct {
	choose key.Binding
	remove key.Binding
}

func InitModelListGames() listGamesModel {
	var (
		itemGenerator randomItem
		delegateKeys  = newDelegateKeyMap()
		listKeys      = newListKeyMap()
	)

	// Make initial list of items
	const numItems = 2
	items := make([]list.Item, numItems)
	for i := range numItems {
		items[i] = itemGenerator.next()
	}

	// Setup list
	delegate := newItemDelegate(delegateKeys)
	gameList := list.New(items, delegate, 0, 0)
	gameList.Title = "Games"
	gameList.Styles.Title = listGamesTitleStyle
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
		list:          gameList,
		keys:          listKeys,
		delegateKeys:  delegateKeys,
		itemGenerator: &itemGenerator,
	}
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		insertItem: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add item"),
		),
		toggleSpinner: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "toggle spinner"),
		),
		toggleTitleBar: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "toggle title"),
		),
		toggleStatusBar: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "toggle status"),
		),
		togglePagination: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle pagination"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
	}
}

func (m model) ListGamesUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := listGamesAppStyle.GetFrameSize()
		m.listGames.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.listGames.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.listGames.keys.toggleSpinner):
			cmd := m.listGames.list.ToggleSpinner()
			return m, cmd

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
			if i, ok := m.listGames.list.SelectedItem().(item); ok {
				m.currentUI = i.GameId()
			}
			return m, nil

			// case key.Matches(msg, m.listGames.keys.insertItem):
			// 	m.listGames.delegateKeys.remove.SetEnabled(true)
			// 	newItem := m.listGames.itemGenerator.next()
			// 	insCmd := m.listGames.list.InsertItem(0, newItem)
			// 	statusCmd := m.listGames.list.NewStatusMessage(listGamesStatusMessageStyle("Added " + newItem.Title()))
			// 	return m, tea.Batch(insCmd, statusCmd)
			// }
		}
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.listGames.list.Update(msg)
	m.listGames.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) ListGamesView() string {
	return listGamesAppStyle.Render(m.listGames.list.View())
}

func newItemDelegate(keys *delegateKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

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
			case key.Matches(msg, keys.choose):
				return m.NewStatusMessage(listGamesStatusMessageStyle("You choose " + title))

			case key.Matches(msg, keys.remove):
				index := m.Index()
				m.RemoveItem(index)
				if len(m.Items()) == 0 {
					keys.remove.SetEnabled(false)
				}
				return m.NewStatusMessage(listGamesStatusMessageStyle("Deleted " + title))
			}
		}

		return nil
	}

	help := []key.Binding{keys.choose, keys.remove}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

// Additional short help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.choose,
		d.remove,
	}
}

// Additional full help entries. This satisfies the help.KeyMap interface and
// is entirely optional.
func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.choose,
			d.remove,
		},
	}
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
		remove: key.NewBinding(
			key.WithKeys("x", "backspace"),
			key.WithHelp("x", "delete"),
		),
	}
}

type gamesItems struct {
	title       string
	description string
	gamesId     string
}

type randomItem struct {
	display []gamesItems
	index   int
	mtx     *sync.Mutex
	shuffle *sync.Once
}

func (r *randomItem) reset() {
	r.mtx = &sync.Mutex{}
	r.shuffle = &sync.Once{}

	r.display = []gamesItems{
		{
			"Snake",
			"A classic snake game where you control a snake to eat food and grow longer.",
			SNAKE_GAME_UI,
		},
		{
			"Tetris - (Coming Soon)",
			"A classic puzzle game where you fit falling blocks together to clear lines.",
			"",
		},
	}

	r.shuffle.Do(func() {
		shuf := func(x []gamesItems) {
			rand.Shuffle(len(x), func(i, j int) { x[i], x[j] = x[j], x[i] })
		}
		shuf(r.display)
	})
}

func (r *randomItem) next() item {
	if r.mtx == nil {
		r.reset()
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	i := item{
		title:       r.display[r.index].title,
		description: r.display[r.index].description,
		gameId:      r.display[r.index].gamesId,
	}

	r.index++
	if r.index >= len(r.display) {
		r.index = 0
	}

	return i
}
