package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	"StockNet/internal/database"
	"fmt"
)

// struct to hold a incoming request
type IncFriendRequest struct {
	SenderEmail		string
	SenderName		string
}
// holds requests for update function
type IncRequestsLoadedMsg struct {
	requests []IncFriendRequest
}
// holds error for update function
type IncRequestsLoadError struct {
    err error
}

// Model for the incoming friend request page
type IncFriReqModel struct {
	backPressed bool
	user        *auth.User
	selected	int
	requests	[]IncFriendRequest
	loading		bool
	error		string
}

// returns initial incoming friend request page model
func newIncFriReqPage(user *auth.User) *IncFriReqModel {
	return &IncFriReqModel{
		user: 		user,
		requests: 	[]IncFriendRequest{},
		selected:	0,
		loading:	true,
	}
}
// returns initial command for the incoming friend request page to run
func (m *IncFriReqModel) Init() tea.Cmd {
	// we get incoming requests from DB
	return func() tea.Msg {

		db := database.New().GetDB()
		// get senders email and their name
		rows, err := db.Query(`
			SELECT sender, name
			FROM friendstatus
			INNER JOIN accounts ON email = sender
			WHERE receiver = $1 AND status = 'pending'
		`, m.user.Email)

		if err != nil {
			return IncRequestsLoadError{err: err} 
		}
		defer rows.Close()

		// store within requests 
		var requests []IncFriendRequest
		for rows.Next() {
			var fr IncFriendRequest
			if err := rows.Scan(&fr.SenderEmail, &fr.SenderName); err != nil {
				return IncRequestsLoadError{err: err} 
			}
			requests = append(requests, fr)
		}
		return IncRequestsLoadedMsg{requests: requests}
	}
}
// Handles incoming events and updates the model accordingly
func (m *IncFriReqModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case IncRequestsLoadedMsg:
		m.requests = msg.requests
		m.loading = false
		m.error = ""

	case IncRequestsLoadError:
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
		case "a":
			// accept the friend request
			if len(m.requests) == 0 {
        		break
    		}
			// get the highlighted request
			req := m.requests[m.selected]
    		db := database.New().GetDB()

			_, err := db.Exec(`
				UPDATE friendstatus
				SET status = 'accepted',
					last_modified = NOW()
				WHERE sender = $1 AND receiver = $2 AND status = 'pending'
			`, req.SenderEmail, m.user.Email)

			if err != nil {
				m.error = fmt.Sprintf("Failed to reject request: %v", err)
				break
			}

			// remove from UI for that request
			m.requests = append(m.requests[:m.selected], m.requests[m.selected+1:]...)
			// fix indexing of requests
			if m.selected >= len(m.requests) && m.selected > 0 {
				m.selected--
			}
			m.error = ""

			
		case "r":
			// reject the friend request
			if len(m.requests) == 0 {
        		break
    		}
			// get the highlighted request
			req := m.requests[m.selected]
    		db := database.New().GetDB()

			// mark the request as rejected and update last_modified so cooldown applies
			_, err := db.Exec(`
				UPDATE friendstatus
				SET status = 'rejected',
					last_modified = NOW()
				WHERE sender = $1 AND receiver = $2 AND status = 'pending'
			`, req.SenderEmail, m.user.Email)
			if err != nil {
				m.error = fmt.Sprintf("Failed to reject request: %v", err)
				break
			}

			// Remove from UI list (it remains in DB with status='rejected')
			m.requests = append(m.requests[:m.selected], m.requests[m.selected+1:]...)
			if m.selected >= len(m.requests) && m.selected > 0 {
				m.selected--
			}
			m.error = ""


		}
	}
	return m, nil
}

func (m *IncFriReqModel) View() string {
	s := "\n"
	s += TitleStyle.Render("👋 Incoming Friend Requests") + "\n\n"

	if m.loading {
		s += "Loading incoming requests...\n"
	} else if m.error != "" {
			s += ErrorStyle.Render(m.error) + "\n"
	} else if len(m.requests) == 0 {
		s += "No pending requests.\n"
	} else {
		for i, req := range m.requests {
			line := fmt.Sprintf("%s (%s)", req.SenderName, req.SenderEmail)
			if i == m.selected {
			s += fmt.Sprintf("%s\n", SelectedStyle.Render("→ "+ line))
			} else {
				s += fmt.Sprintf("%s\n", UnselectedStyle.Render("  "+ line))
			}
		}
	}

	s += FooterStyle.Render("'a' or 'r' to accept or reject • ↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}








