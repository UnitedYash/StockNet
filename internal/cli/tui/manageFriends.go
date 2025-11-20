package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	"StockNet/internal/database"
	"fmt"

)

// struct to hold a friend
type Friend struct {
	friendEmail		string
	friendName		string
}
// holds friends for update function
type FriendsLoadedMsg struct {
	friends []Friend
}
// holds error for update function
type FriendsLoadError struct {
    err error
}

// Model for the manage friend page
type ManageFriendsModel struct {
	backPressed bool
	user        *auth.User
	selected	int
	friends		[]Friend
	loading		bool
	error		string

}

// returns initial stock list model
func newManageFriendsPage(user *auth.User) *ManageFriendsModel {
	return &ManageFriendsModel{
		user: 		user,
		friends:	[]Friend{},
		selected:	0,
		loading:	true,
	}
}
// returns initial command for the manage page to run
func (m *ManageFriendsModel) Init() tea.Cmd {
	return func() tea.Msg {
		// get all friends
		db := database.New().GetDB()
		
		rows, err := db.Query(`
			SELECT 
				a.email AS friend_email,
				a.name  AS friend_name
			FROM friendstatus f
			JOIN accounts a 
				ON a.email = (
					CASE 
						WHEN f.sender = $1 THEN f.receiver
						ELSE f.sender
					END
				)
			WHERE 
				(f.sender = $1 OR f.receiver = $1)
				AND f.status = 'accepted';
		`, m.user.Email)

		if err != nil {
			return FriendsLoadError{err: err}
		}
		defer rows.Close()

		// store results within struct
		var friends []Friend
		for rows.Next() {
			var fr Friend
			if err := rows.Scan(&fr.friendEmail, &fr.friendName); err != nil {
				return FriendsLoadError{err: err} 
			}
			friends = append(friends, fr)

		}
		return FriendsLoadedMsg{friends: friends}
	}
}
// Handles incoming events and updates the model accordingly
func (m *ManageFriendsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FriendsLoadedMsg:
		m.friends = msg.friends
		m.loading = false
		m.error = ""
	case FriendsLoadError:
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
				m.selected = len(m.friends) - 1
			}
		case "down", "j":
			if m.selected < len(m.friends) - 1 {
				m.selected++
			} else {
				// at last option so wrap around to the top
				m.selected = 0
			}
		case "r":
			//TODO: press r to remove a friend
			if len(m.friends) == 0 {
				break
			}
			// get the highlighted friend 
			fri := m.friends[m.selected]
			db := database.New().GetDB()

			// change the relationship to status deleted
			_, err := db.Exec(`
				UPDATE friendstatus
				SET status = 'deleted',
					last_modified = NOW()
				WHERE 
					(
						(sender = $1 AND receiver = $2)
						OR
						(sender = $2 AND receiver = $1)
					)
					AND status = 'accepted'
			`, m.user.Email, fri.friendEmail)
			if err != nil {
				m.error = fmt.Sprintf("Error removing friend: %v", err)
				break
			}
			// remove friend from UI friends list
			m.friends = append(m.friends[:m.selected], m.friends[m.selected+1:]...)

			// fix the indexing
			if m.selected >= len(m.friends) && len(m.friends) > 0 {
				m.selected = len(m.friends) - 1
			}
			m.error = ""


		}
	}
	return m, nil
}

func (m *ManageFriendsModel) View() string {
	s := "\n"
	s += TitleStyle.Render("🫂  Manage Friends") + "\n\n"

	if m.loading {
		s += "Loading friends...\n"
	} else if m.error != "" {
			s += ErrorStyle.Render(m.error) + "\n"
	} else if len(m.friends) == 0 {
		s += "No Friends.\n"
	} else {
		for i, fri := range m.friends {
			line := fmt.Sprintf("%s (%s)", fri.friendEmail, fri.friendName)
			if i == m.selected {
			s += fmt.Sprintf("%s\n", SelectedStyle.Render("→ "+ line))
			} else {
				s += fmt.Sprintf("%s\n", UnselectedStyle.Render("  "+ line))
			}
		}
	}

	s += FooterStyle.Render("'r' to remove friends • ↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}
