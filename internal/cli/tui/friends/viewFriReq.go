package friends

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	"StockNet/internal/cli/tui/styles"
	"fmt"
)

// Note this is not for a specifc direction of friend requests 
// ie just a screen to select to view incoming or outgoing

// Model for the view friend reqests page
type ViewFriReqModel struct {
	Options		[]string
	Selected  	int
	BackPressed bool
	Confirmed 	bool
	User        *auth.User
}

// returns initial view friend request page model
func NewViewFriReqPage(user *auth.User) *ViewFriReqModel {
	return &ViewFriReqModel{
		User: user,
		Options: []string{
			"Incoming Requests (Accept / Reject)",
			"Outgoing Requests (Cancel)",
		},
		Selected: 0,
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
			m.BackPressed = true
		case "up", "k":
			if m.Selected > 0 {
				// go up an option
				m.Selected--
			} else {
				// at the top so wrap around to bottom
				m.Selected = len(m.Options) - 1
			}
		case "down", "j":
			if m.Selected < len(m.Options) - 1 {
				m.Selected++
			} else {
				// at last option so wrap around to the top
				m.Selected = 0
			}
		case "enter":
			m.Confirmed = true
		}
	}
	return m, nil
}

func (m *ViewFriReqModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("⛹️  View Friend Requests") + "\n\n"

	// highlight the selected option with a →
	for i, option := range m.Options {
		if i == m.Selected {
			s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+option))
		} else {
			s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+option))
		}
	}
	s += styles.FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}

// returns (logged in) user 
func (m *ViewFriReqModel) GetUser() *auth.User {
	return m.User
}







