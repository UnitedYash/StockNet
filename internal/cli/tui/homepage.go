package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type HomePageModel struct {
	backPressed bool
}

func NewHomePage() *HomePageModel {
	return &HomePageModel{}
}

func (m *HomePageModel) Init() tea.Cmd {
	return nil
}

func (m *HomePageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+l", "esc":
			m.backPressed = true
		}
	}
	return m, nil
}

func (m *HomePageModel) View() string {
	s := "\n"
	s += TitleStyle.Render("📊 Home") + "\n\n"
	s += InputStyle.Render("Welcome to StockNet!") + "\n\n"
	s += FooterStyle.Render("Press 'Ctrl + l' or 'Esc' to logout") + "\n\n"
	return s
}
