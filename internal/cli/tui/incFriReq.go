package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"

)

// Model for the incoming friend request page
type IncFriReqModel struct {
	backPressed bool
	user        *auth.User
	message		string
}

// returns initial incoming friend request page model
func newIncFriReqPage(user *auth.User) *IncFriReqModel {
	return &IncFriReqModel{
		user: user,
	}
}
// returns initial command for the incoming friend request page to run (nothing)
func (m *IncFriReqModel) Init() tea.Cmd {
	return nil
}
// Handles incoming events and updates the model accordingly
func (m *IncFriReqModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.backPressed = true
		}
	}
	return m, nil
}

func (m *IncFriReqModel) View() string {
	s := "\n"
	s += TitleStyle.Render("👋 Incoming Friend Requests") + "\n\n"
	s += FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}








