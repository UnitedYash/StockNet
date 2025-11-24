package friends

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
	"fmt"
)


// struct to hold a outgoing request
type OutFriendRequest struct {
	ReceiverEmail	string
	ReceiverName	string
}
// holds requests for update function
type OutRequestsLoadedMsg struct {
	requests []OutFriendRequest
}
// holds error for update function
type OutRequestsLoadError struct {
    err error
}

// Model for the outgoing friend request page
type OutFriReqModel struct {
	BackPressed bool
	User        *auth.User
	Selected	int
	Requests	[]OutFriendRequest
	Loading		bool
	Error		string
}

// returns initial outgoing friend request page model
func NewOutFriReqPage(user *auth.User) *OutFriReqModel {
	return &OutFriReqModel{
		User: 		user,
		Requests: 	[]OutFriendRequest{},
		Selected:	0,
		Loading:	true,
	}
}
// returns initial command for the outgoing friend request page to run (nothing)
func (m *OutFriReqModel) Init() tea.Cmd {
	// we get outgoing requests from DB
	return func() tea.Msg {

		db := database.New().GetDB()
		// get receivers email and their name
		rows, err := db.Query(`
			SELECT receiver, name
			FROM friendstatus
			INNER JOIN accounts ON email = receiver
			WHERE sender = $1 AND status = 'pending'
		`, m.User.Email)

		if err != nil {
			return OutRequestsLoadError{err: err} 
		}
		defer rows.Close()

		// store within requests 
		var requests []OutFriendRequest
		for rows.Next() {
			var fr OutFriendRequest
			if err := rows.Scan(&fr.ReceiverEmail, &fr.ReceiverName); err != nil {
				return OutRequestsLoadError{err: err} 
			}
			requests = append(requests, fr)
		}
		return OutRequestsLoadedMsg{requests: requests}
	}
}
// Handles outgoing events and updates the model accordingly
func (m *OutFriReqModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case OutRequestsLoadedMsg:
		m.Requests = msg.requests
		m.Loading = false
		m.Error = ""

	case OutRequestsLoadError:
		m.Loading = false
		m.Error = fmt.Sprintf("Error loading requests: %v", msg.err)
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
				m.Selected = len(m.Requests) - 1
			}
		case "down", "j":
			if m.Selected < len(m.Requests) - 1 {
				m.Selected++
			} else {
				// at last option so wrap around to the top
				m.Selected = 0
			}
		case "c":
			// we cancel 
			if len(m.Requests) == 0 {
				break
			}
			// get the selected user
			req := m.Requests[m.Selected]

			db := database.New().GetDB()

			// Run DELETE query
			_, err := db.Exec(`
				DELETE FROM friendstatus
				WHERE sender = $1 AND receiver = $2 AND status = 'pending'
			`, m.User.Email, req.ReceiverEmail)


			if err != nil {
				m.Error = fmt.Sprintf("Failed to cancel request: %v", err)
				break
			}
			// we also want to delete it from the UI
			m.Requests = append(m.Requests[:m.Selected], m.Requests[m.Selected+1:]...)
			// fix index if needed
			if m.Selected >= len(m.Requests) && m.Selected > 0 {
				m.Selected--
			}
			m.Error = ""
		}
	}
	return m, nil
}

func (m *OutFriReqModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📤 Outgoing Friend Requests") + "\n\n"

	if m.Loading {
		s += "Loading outgoing requests...\n"
	} else if m.Error != "" {
			s += styles.ErrorStyle.Render(m.Error) + "\n"
	} else if len(m.Requests) == 0 {
		s += "No pending requests.\n"
	} else {
		for i, req := range m.Requests {
			line := fmt.Sprintf("%s (%s)", req.ReceiverName, req.ReceiverEmail)
			if i == m.Selected {
			s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+ line))
			} else {
				s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+ line))
			}
		}
	}

	s += styles.FooterStyle.Render("'c' to cancel request • ↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}








