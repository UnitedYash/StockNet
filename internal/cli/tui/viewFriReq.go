package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	"fmt"
)

// Note this is not for a specifc direction of friend requests 
// ie just a screen to select to view incoming or outgoing

// Model for the view friend reqests page
type ViewFriReqModel struct {
	options		[]string
	selected  	int
	backPressed bool
	confirmed 	bool
	user        *auth.User
}

// returns initial view friend request page model
func newViewFriReqPage(user *auth.User) *ViewFriReqModel {
	return &ViewFriReqModel{
		user: user,
		options: []string{
			"Incoming Requests (Accept / Reject)",
			"Outgoing Requests (Cancel)",
		},
		selected: 0,
	}
}
// returns initial command for the view friend request page to run (nothing)
func (m *ViewFriReqModel) Init() tea.Cmd {
	return nil
}
// Handles incoming events and updates the model accordingly
func (m *ViewFriReqModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.backPressed = true
		case "up", "k":
			if m.selected > 0 {
				// go up an option
				m.selected--
			} else {
				// at the top so wrap around to bottom
				m.selected = len(m.options) - 1
			}
		case "down", "j":
			if m.selected < len(m.options) - 1 {
				m.selected++
			} else {
				// at last option so wrap around to the top
				m.selected = 0
			}
		case "enter":
			m.confirmed = true
		}
	}
	return m, nil
}

func (m *ViewFriReqModel) View() string {
	s := "\n"
	s += TitleStyle.Render("⛹️  View Friend Requests") + "\n\n"

	// highlight the selected option with a →
	for i, option := range m.options {
		if i == m.selected {
			s += fmt.Sprintf("%s\n", SelectedStyle.Render("→ "+option))
		} else {
			s += fmt.Sprintf("%s\n", UnselectedStyle.Render("  "+option))
		}
	}
	s += FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}

// returns (logged in) user 
func (m *ViewFriReqModel) GetUser() *auth.User {
	return m.user
}







