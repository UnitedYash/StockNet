package stocklist

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
	//"fmt"
)

// Model for the write review page
type WriteReviewModel struct {
	StockList		StockList
	Error       	string
	BackPressed 	bool
	SuccessMessage	string
	Confirmed		bool
	User        	*auth.User
	UserInput		string
	Cursor			int			// when editing a review, a cursor is useful 

}


// returns initial share stock list model
func NewWriteReviewPage(stockList StockList, user *auth.User) *WriteReviewModel {
	// if they have an existing review we going to put that in textbox
	userInput := ""
	if user != nil {
		existingText, err := GetReview(int(user.UserID), stockList.StockListID)
		if err == nil {
			userInput = existingText
		}
	}

	return &WriteReviewModel{
		User: 			user,
		StockList:		stockList,
		Error:			"",
		SuccessMessage:	"",
		UserInput:		userInput,
		Cursor:			0,
		
	}
}

// returns initial command for write review page to run (nothing)
func (m *WriteReviewModel) Init() tea.Cmd {
	return nil
}

// Handles incoming events and updates the model accordingly
func (m *WriteReviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.BackPressed = true
		case "enter":
			if m.UserInput == "" {
				break
			}
			err := AddOrUpdateReview(int(m.User.UserID), m.StockList.StockListID, m.UserInput)
			if err != nil {
				m.Error = "Failed to save review: " + err.Error()
			} else {
				m.SuccessMessage = "Review saved successfully!"
			}
		case "backspace":
			// delete at cursor
			if m.Cursor > 0 {
				m.UserInput = m.UserInput[:m.Cursor-1] + m.UserInput[m.Cursor:]
				m.Cursor--
			}
		case "left":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "right":
			if m.Cursor < len(m.UserInput) {
				m.Cursor++
			}
		default:
			// any other key would indicate user typing in the text field
			// also block any non-printable letters
			if len(msg.String()) == 1 {
				r := rune(msg.String()[0])
				if r >= 32 && r != 127 {
					// insert at cursor
					m.UserInput = m.UserInput[:m.Cursor] + string(r) + m.UserInput[m.Cursor:]
					m.Cursor++
				}
			}
		}
	}
	return m, nil

}

func (m *WriteReviewModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("✍️  Write a Review") + "\n\n"
	// display the cursor
	inputDisplay := m.UserInput
	if m.Cursor <= len(inputDisplay) {
		if m.Cursor == len(inputDisplay) {
			inputDisplay += "|"
		} else {
			inputDisplay = inputDisplay[:m.Cursor] + "|" + inputDisplay[m.Cursor:]
		}
	}
	s += styles.InputStyle.Render("Enter Review: "+ inputDisplay) + "\n"

	if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n"
	} else if m.SuccessMessage != "" {
		s += styles.SuccessStyle.Render(m.SuccessMessage) + "\n"
	}
	s += styles.FooterStyle.Render("'Enter' to Confirm • 'Ctrl + b' or 'Esc' to go back") + "\n\n"

	return s
}

// a function to get the preexisting review so we can load it in for user editing
func GetReview(userID int, stocklistID int) (string, error) {
    db := database.New().GetDB()
    var text string

    err := db.QueryRow(`SELECT text FROM reviews WHERE user_id=$1 AND stocklist_id=$2;`,
        userID, stocklistID).Scan(&text)

    if err != nil {
        return "", err // no review found OR query error
    }

    return text, nil
}
// update or insert the new revire for a given stocklist
func AddOrUpdateReview(userID int, stocklistID int, text string) error {
    db := database.New().GetDB()

    _, err := db.Exec(`
        INSERT INTO reviews (user_id, stocklist_id, text)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id, stocklist_id)
        DO UPDATE SET text = EXCLUDED.text;`, 
        userID, stocklistID, text)

    return err
}

