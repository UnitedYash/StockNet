package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type MainMenuModel struct {
	options   []string
	selected  int
	confirmed bool
}

func NewMainMenu() *MainMenuModel {
	return &MainMenuModel{
		options: []string{
			"Login",
			"Register",
			"Configure",
			"Quit",
		},
		selected: 0,
	}
}

func (m *MainMenuModel) Init() tea.Cmd {
	return nil
}

func (m *MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			} else {
				m.selected = len(m.options) - 1
			}
		case "down", "j":
			if m.selected < len(m.options)-1 {
				m.selected++
			} else {
				m.selected = 0
			}
		case "enter":
			m.confirmed = true
		}
	}
	return m, nil
}

func (m *MainMenuModel) View() string {
	s := "\n"
	s += TitleStyle.Render("🚀 StockNet") + "\n\n"

	for i, option := range m.options {
		if i == m.selected {
			s += fmt.Sprintf("%s\n", SelectedStyle.Render("→ "+option))
		} else {
			s += fmt.Sprintf("%s\n", UnselectedStyle.Render("  "+option))
		}
	}

	s += FooterStyle.Render("↑/↓ or k/j to navigate • Enter to select • q to quit") + "\n\n"
	return s
}
