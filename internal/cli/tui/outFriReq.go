package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
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
	backPressed bool
	user        *auth.User
	selected	int
	requests	[]OutFriendRequest
	loading		bool
	error		string
}

// returns initial outgoing friend request page model
func newOutFriReqPage(user *auth.User) *OutFriReqModel {
	return &OutFriReqModel{
		user: 		user,
		requests: 	[]OutFriendRequest{},
		selected:	0,
		loading:	true,
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
		`, m.user.Email)

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
		m.requests = msg.requests
		m.loading = false
		m.error = ""

	case OutRequestsLoadError:
		m.loading = false
		m.error = fmt.Sprintf("Error loading requests: %v", msg.err)
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
				m.selected = len(m.requests) - 1
			}
		case "down", "j":
			if m.selected < len(m.requests) - 1 {
				m.selected++
			} else {
				// at last option so wrap around to the top
				m.selected = 0
			}
		case "c":
			// we cancel 
			if len(m.requests) == 0 {
				break
			}
			// get the selected user
			req := m.requests[m.selected]

			db := database.New().GetDB()

			// Run DELETE query
			_, err := db.Exec(`
				DELETE FROM friendstatus
				WHERE sender = $1 AND receiver = $2 AND status = 'pending'
			`, m.user.Email, req.ReceiverEmail)


			if err != nil {
				m.error = fmt.Sprintf("Failed to cancel request: %v", err)
				break
			}
			// we also want to delete it from the UI
			m.requests = append(m.requests[:m.selected], m.requests[m.selected+1:]...)
			// fix index if needed
			if m.selected >= len(m.requests) && m.selected > 0 {
				m.selected--
			}
			m.error = ""
		}
	}
	return m, nil
}

func (m *OutFriReqModel) View() string {
	s := "\n"
	s += TitleStyle.Render("📤 Outgoing Friend Requests") + "\n\n"

	if m.loading {
		s += "Loading outgoing requests...\n"
	} else if m.error != "" {
			s += ErrorStyle.Render(m.error) + "\n"
	} else if len(m.requests) == 0 {
		s += "No pending requests.\n"
	} else {
		for i, req := range m.requests {
			line := fmt.Sprintf("%s (%s)", req.ReceiverName, req.ReceiverEmail)
			if i == m.selected {
			s += fmt.Sprintf("%s\n", SelectedStyle.Render("→ "+ line))
			} else {
				s += fmt.Sprintf("%s\n", UnselectedStyle.Render("  "+ line))
			}
		}
	}

	s += FooterStyle.Render("'c' to cancel request • ↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}








