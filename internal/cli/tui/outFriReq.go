package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"

)

// Model for the outgoing friend request page
type OutFriReqModel struct {
	backPressed bool
	user        *auth.User
	message		string
}

// returns initial outgoing friend request page model
func newOutFriReqPage(user *auth.User) *OutFriReqModel {
	return &OutFriReqModel{
		user: user,
	}
}
// returns initial command for the outgoing friend request page to run (nothing)
func (m *OutFriReqModel) Init() tea.Cmd {
	return nil
}
// Handles outgoing events and updates the model accordingly
func (m *OutFriReqModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.backPressed = true
		}
	}
	return m, nil
}

func (m *OutFriReqModel) View() string {
	s := "\n"
	s += TitleStyle.Render("📤 Outgoing Friend Requests") + "\n\n"
	s += FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}








