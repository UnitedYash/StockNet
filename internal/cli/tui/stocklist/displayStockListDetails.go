package stocklist

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/auth"
	"fmt"
	"StockNet/internal/database"

)

// Model for a displaying stocklist
type DisplayStockListModel struct {
	StockList		StockList
    Selected    	int
    BackPressed 	bool
    Error       	string
	SuccessMessage	string
	User        	*auth.User	
	OwnerUserID 	uint32
	Options			[]string
	Confirmed		bool
	DeleteList		bool
}

// returns initial displaying  list model
func NewDisplayStockListPage(stockList StockList, user *auth.User, OwnerID uint32) *DisplayStockListModel {
	model := &DisplayStockListModel{
		StockList:		stockList,
		Selected:		0,
		BackPressed:	false,
		Error:			"",
		SuccessMessage:	"",
		User:			user,
		OwnerUserID:	OwnerID,
		DeleteList:		false,
	}
	// if the current user is the owner of the stocklist, give edit and delete options
	if user != nil && user.UserID == OwnerID {
		model.Options = []string{"View Stocks", "Edit List", "Reviews", "Share", "Toggle Visibility", "Delete List"}
	} else {
		model.Options = []string{"View Stocks", "Reviews"}
	}

	return model
}

// returns initial command for displaying stock list page to run (nothing)
func (m *DisplayStockListModel) Init() tea.Cmd {
	return nil
}

// Handles incoming events and updates the model accordingly
func (m *DisplayStockListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			option := m.Options[m.Selected]
			switch option {
			case "Toggle Visibility":
				newVis, err := ToggleVisibility(m.StockList)
				if err != nil {
					m.Error = fmt.Sprintf("Failed to toggle visibility: %v", err)
				} else {
					// Update the model so View() shows the new visibility
					m.StockList.Visibility = newVis
					m.SuccessMessage = fmt.Sprintf("Visibility changed to %s", newVis)
				}
			}
		}
	}
	return m, nil
}

func (m *DisplayStockListModel) View() string {

	s := "\n"
	s += styles.TitleStyle.Render("📃 " + m.StockList.Name + " Stock List") + "\n\n"

	// Stock list details
	s += fmt.Sprintf("Visibility: %s\n", m.StockList.Visibility)
	if m.User != nil && m.User.UserID == m.OwnerUserID {
		s += "Owner: You\n"
	} else {
		s += fmt.Sprintf("Owner UserID: %d\n", m.OwnerUserID)
	}
	s += "\n"

	// highlight the selected option with a →
	for i, option := range m.Options {
		if i == m.Selected {
			s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+option))
		} else {
			s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+option))
		}
	}
	if m.SuccessMessage != "" {
    	s += styles.SuccessStyle.Render(m.SuccessMessage) + "\n"
	}
	if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n"
	}
	s += styles.FooterStyle.Render("Enter to select • ↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}


// a function to delete the current stocklist
func DeleteStockList(stocklistID int) error {
	db := database.New().GetDB()
    _, err := db.Exec(`DELETE FROM stocklist WHERE stocklist_id=$1`, stocklistID)
    return err
}

func ToggleVisibility(stockList StockList) (newVisibility string, err error) {
	db := database.New().GetDB()
	currentVis := stockList.Visibility
	// new vis is opposite
	if currentVis == "private" {
        newVisibility = "public"
    } else {
        newVisibility = "private"
    }

	// update the database with new vis
	_, err = db.Exec(`
        UPDATE stocklist
        SET visibility = $1
        WHERE stocklist_id = $2
    `, newVisibility, stockList.StockListID)
    if err != nil {
        return "", fmt.Errorf("failed to update visibility: %v", err)
    }
	return newVisibility, nil

}
