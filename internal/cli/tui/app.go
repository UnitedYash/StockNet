package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	"github.com/charmbracelet/lipgloss"
)

// AppState represents different screens in the app
type AppState int

const (
	MainMenuState AppState = iota 	// 0
	LoginState						// 1
	RegisterState					// 2
	ConfigureState					// 3
	HomePageState					// 4
	PortfolioState					// 5
	StockListState					// 6
	StockAnalysisState				// 7
	SocialState						// 8
)

// AppModel is the root model for the entire app
type AppModel struct {
	state        	AppState
	currentUser 	auth.User // currently logged-in user
	mainMenu    	*MainMenuModel
	login       	*LoginModel
	register    	*RegisterModel
	configure   	*ConfigureModel
	homepage    	*HomePageModel
	portfolio		*PortfolioModel
	stockList		*StockListModel
	stockAnalysis 	*StockAnalysisModel
	social			*SocialModel

}

// NewAppModel creates a new app model
func NewAppModel() *AppModel {
	return &AppModel{
		state:     		MainMenuState,
		mainMenu:  		NewMainMenu(),
		login:     		NewLogin(),
		register:  		NewRegister(),
		configure: 		NewConfigure(),
		homepage:  		NewHomePage(nil),
		portfolio: 		newPortfolioPage(),
		stockList: 		newStockListPage(),
		stockAnalysis: 	newStockAnalysisPage(),
		social:			newSocialPage(),
	}
}
// returns intial command for the application to run
func (m *AppModel) Init() tea.Cmd {
	// Note: For now, we don't have any initial I/O commands to do. nil = "no command"
	return nil
}
// handles incoming events and updates root model and runs corresponding commands
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global quit key
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// Route to the appropriate state handler
	switch m.state {
	// give msg input to Update method of the currently active model
	// and changes app model state if a button/option (that is meant for switching views) gets triggered

	case MainMenuState:
		menu, cmd := m.mainMenu.Update(msg)
		m.mainMenu = menu.(*MainMenuModel)

		// get the selected option and switch to that model if enter key pressed (confirmed)
		if m.mainMenu.selected >= 0 && m.mainMenu.selected < len(m.mainMenu.options) {
			option := m.mainMenu.options[m.mainMenu.selected]
			if m.mainMenu.confirmed {
				m.mainMenu.confirmed = false
				switch option {
				case "Login":
					m.state = LoginState
					m.login = NewLogin()
				case "Register":
					m.state = RegisterState
					m.register = NewRegister()
				case "Configure":
					m.state = ConfigureState
					m.configure = NewConfigure()
				case "Quit":
					return m, tea.Quit
				}
			}
		}
		return m, cmd

	case LoginState:
		login, cmd := m.login.Update(msg)
		m.login = login.(*LoginModel)
		// Check if user successfully logged in
		if m.login.GetUser() != nil {
			m.currentUser = *m.login.GetUser()
			m.state = HomePageState
			m.homepage = NewHomePage(m.login.GetUser()) // reset homepage
			m.login = NewLogin() // reset login form
		}
		// Go back to main view from Login view
		if m.login.backPressed {
			m.login.backPressed = false
			m.state = MainMenuState
		}
		return m, cmd

	case RegisterState:
		register, cmd := m.register.Update(msg)
		m.register = register.(*RegisterModel)
		// Check if user successfully registered
		if m.register.GetUser() != nil {
			m.currentUser = *m.register.GetUser()
			m.state = HomePageState
			m.homepage = NewHomePage(m.register.GetUser()) // reset homepage
			m.register = NewRegister() // reset register form
		}
		// Go back to main view from register view
		if m.register.backPressed {
			m.register.backPressed = false
			m.state = MainMenuState
		}
		return m, cmd

	case ConfigureState:
		configure, cmd := m.configure.Update(msg)
		m.configure = configure.(*ConfigureModel)

		// Go back to home view from configure view
		if m.configure.backPressed {
			m.configure.backPressed = false
			m.state = MainMenuState
		}
		return m, cmd

	case HomePageState:
		homepage, cmd := m.homepage.Update(msg)
		m.homepage = homepage.(*HomePageModel)

		// get the selected option and switch to that model if enter key pressed (confirmed)
		if m.homepage.selected >= 0 && m.homepage.selected < len(m.homepage.options) {
			option := m.homepage.options[m.homepage.selected]
			if m.homepage.confirmed {
				m.homepage.confirmed = false
				switch option {
				case "My Portfolios":
					m.state = PortfolioState
					m.portfolio = newPortfolioPage()
				case "My Stock Lists":
					m.state = StockListState
					m.stockList = newStockListPage()
				case "Stock Data & Analysis":
					m.state = StockAnalysisState
					m.stockAnalysis = newStockAnalysisPage()
				case "Friends & Social":
					m.state = SocialState
					m.social = newSocialPage()
				}
			}
		}

		// Go back to main menu (logout) from HomePage
		if m.homepage.backPressed {
			m.homepage.backPressed = false
			m.state = MainMenuState
			m.currentUser = auth.User{} // clear current user
		}
		return m, cmd
	case PortfolioState:
		portfolio, cmd := m.portfolio.Update(msg)
		m.portfolio = portfolio.(*PortfolioModel)
		// Go back to homepage from portfolio page
		if m.portfolio.backPressed {
			m.portfolio.backPressed = false
			m.state = HomePageState
		}
		return m, cmd
	case StockListState:
		stockList, cmd := m.stockList.Update(msg)
		m.stockList = stockList.(*StockListModel)
		// Go back to homepage from stock list page
		if m.stockList.backPressed {
			m.stockList.backPressed = false
			m.state = HomePageState
		}
		return m, cmd
	case StockAnalysisState:
		stockAnalysis, cmd := m.stockAnalysis.Update(msg)
		m.stockAnalysis = stockAnalysis.(*StockAnalysisModel)
		// Go back to homepage from Stock Data & Analysis page
		if m.stockAnalysis.backPressed {
			m.stockAnalysis.backPressed = false
			m.state = HomePageState
		}
		return m, cmd
	case SocialState:
		social, cmd := m.social.Update(msg)
		m.social = social.(*SocialModel)
		// Go back to homepage from social page
		if m.social.backPressed {
			m.social.backPressed = false
			m.state = HomePageState
		}
		return m, cmd
	}

	return m, nil
}

// Render the current model view
func (m *AppModel) View() string {
	switch m.state {
	case MainMenuState:
		return m.mainMenu.View()
	case LoginState:
		return m.login.View()
	case RegisterState:
		return m.register.View()
	case ConfigureState:
		return m.configure.View()
	case HomePageState:
		return m.homepage.View()
	case PortfolioState:
		return m.portfolio.View()
	case StockListState:
		return m.stockList.View()
	case StockAnalysisState:
		return m.stockAnalysis.View()
	case SocialState:
		return m.social.View()
	}
	return ""
}

// StartApp starts the interactive CLI app
func StartApp() error {
	p := tea.NewProgram(NewAppModel())
	_, err := p.Run()
	return err
}

// Common styles
var (
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("3")).
		PaddingLeft(2).
		PaddingRight(2)

	SelectedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("2")).
		Background(lipgloss.Color("8")).
		PaddingLeft(2)

	UnselectedStyle = lipgloss.NewStyle().
		PaddingLeft(2)

	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6")).
		PaddingLeft(2)

	FooterStyle = lipgloss.NewStyle().
		Faint(true).
		PaddingTop(1).
		PaddingLeft(2)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("1")).
		PaddingLeft(2)

	SuccessStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")).
		PaddingLeft(2)

	InputStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("6"))
)
