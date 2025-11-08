package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AppState represents different screens in the app
type AppState int

const (
	MainMenuState AppState = iota
	LoginState
	RegisterState
)

// AppModel is the root model for the entire app
type AppModel struct {
	state        AppState
	mainMenu     *MainMenuModel
	login        *LoginModel
	register     *RegisterModel
}

// NewAppModel creates a new app model
func NewAppModel() *AppModel {
	return &AppModel{
		state:    MainMenuState,
		mainMenu: NewMainMenu(),
		login:    NewLogin(),
		register: NewRegister(),
	}
}

func (m *AppModel) Init() tea.Cmd {
	return nil
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global quit key
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// Route to the appropriate state handler
	switch m.state {
	case MainMenuState:
		menu, cmd := m.mainMenu.Update(msg)
		m.mainMenu = menu.(*MainMenuModel)

		if m.mainMenu.selected >= 0 && m.mainMenu.selected < len(m.mainMenu.options) {
			option := m.mainMenu.options[m.mainMenu.selected]
			if m.mainMenu.confirmed {
				m.mainMenu.confirmed = false
				switch option {
				case "Login":
					m.state = LoginState
					m.login = NewLogin()
				case "Register":
					m.state = RegisterState
					m.register = NewRegister()
				case "Quit":
					return m, tea.Quit
				}
			}
		}
		return m, cmd

	case LoginState:
		login, cmd := m.login.Update(msg)
		m.login = login.(*LoginModel)
		if m.login.backPressed {
			m.login.backPressed = false
			m.state = MainMenuState
		}
		return m, cmd

	case RegisterState:
		register, cmd := m.register.Update(msg)
		m.register = register.(*RegisterModel)
		if m.register.backPressed {
			m.register.backPressed = false
			m.state = MainMenuState
		}
		return m, cmd
	}

	return m, nil
}

func (m *AppModel) View() string {
	switch m.state {
	case MainMenuState:
		return m.mainMenu.View()
	case LoginState:
		return m.login.View()
	case RegisterState:
		return m.register.View()
	}
	return ""
}

// StartApp starts the interactive CLI app
func StartApp() error {
	p := tea.NewProgram(NewAppModel())
	_, err := p.Run()
	return err
}

// Common styles
var (
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("3")).
		PaddingLeft(2).
		PaddingRight(2)

	SelectedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("2")).
		Background(lipgloss.Color("8")).
		PaddingLeft(2)

	UnselectedStyle = lipgloss.NewStyle().
		PaddingLeft(2)

	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6")).
		PaddingLeft(2)

	FooterStyle = lipgloss.NewStyle().
		Faint(true).
		PaddingTop(1).
		PaddingLeft(2)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("1")).
		PaddingLeft(2)

	SuccessStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")).
		PaddingLeft(2)

	InputStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("6"))
)
