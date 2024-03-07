package interactive

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type model struct {
	list list.Model
	quit bool

	mode  string // pod2pod, pod2remote, etc
	state string // pickingMode, pickingPodTo, pickingNamespaceTo, typingRemoteURI, pickingPodFrom, pickingNamespaceFrom, typingPort

	namespaceFrom string
	podFrom       string

	remoteURI string

	namespaceTo string
	podTo       string
	portTo      string
}

type item struct {
	title string
}

func (i item) FilterValue() string {
	return i.title
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.title)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("🫧 " + strings.Join(s, " "))
		}
	} else {
		str = " " + str
	}

	fmt.Fprint(w, fn(str))
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil

}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case "pickingMode":
			switch keypress := msg.String(); keypress {
			case "q", "ctrl+c":
				m.quit = true
				return m, tea.Quit

			case "enter":
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.mode = string(i.title)
				}
				return m, tea.Quit
			}
		case "pod2pod":
			switch keypress := msg.String(); keypress {
			case "q", "ctrl+c":
				m.quit = true
				return m, tea.Quit

			case "enter":
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.mode = string(i.title)
				}
				return m, tea.Quit

			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {

	// Iterate over our choices
	if m.mode != "mode" {
		return quitTextStyle.Render(fmt.Sprintf("%s? Sounds good to me.", m.mode))
	}
	if m.quit {
		return quitTextStyle.Render("Not hungry? That’s cool.")
	}
	return "\n" + m.list.View()
}

func startingModel() model {
	modes := []list.Item{
		item{title: "pod2pod"},
		item{title: "pod2remote"},
	}

	const defaultWidth = 20

	l := list.New(modes, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Pick a mode:"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return model{list: l, state: "pickingMode"}
}

func Start() *tea.Program {
	p := tea.NewProgram(startingModel())
	return p
}

func Run(p *tea.Program) error {
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		return err
	}
	return nil
}
