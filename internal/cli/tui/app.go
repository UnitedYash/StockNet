package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	tuiauth "StockNet/internal/cli/tui/auth"
	"StockNet/internal/cli/tui/friends"
	"StockNet/internal/cli/tui/portfolio"
	"StockNet/internal/cli/tui/shared"
	"StockNet/internal/cli/tui/stock"
	"StockNet/internal/cli/tui/stocklist"
	"StockNet/internal/database"
)

// AppState represents different screens in the app
type AppState int

var DefaultStockList = stocklist.StockList{
	StockListID: -1,
	Name:        "Loading...",
	Visibility:  "Private",
}

const (
	MainMenuState AppState = iota 	// 0
	LoginState						// 1
	RegisterState					// 2
	HomePageState					// 3
	PortfolioState					// 4
	StockListState					// 5
	StockAnalysisState				// 6
	SocialState						// 7
	CurrentStocksState				// 8
	SearchStockState				// 9
	StockDetailsState				// 10
	SendFriReqState					// 11
	ViewFriReqState					// 12
	IncFriReqState					// 13
	OutFriReqState					// 14
	ViewPortfoliosState				// 15
	CreatePortfolioState			// 16
	ViewSpecPortfolioState			// 17
	ViewHoldingsState				// 18
	ViewTransactionsState			// 19
	BuyStockSearchState				// 20
	BuyStockState					// 21
	SellStockSearchState			// 22
	SellStockState					// 23
	DepositCashState				// 24
	WithdrawCashState				// 25
	ViewNetWorthState				// 26
	ManageFriendsState				// 27
	CreateStockListState			// 28
	ViewStockListsState				// 29
	SearchStockForDataState			// 30
	AddStockDataState				// 31
	ViewPortfolioStatisticsState	// 32
	DisplayStockListState			// 33
	EditStockListState				// 34
	HoldingDetailsState				// 35
	HoldingForecastState			// 36
	AddStockToListState				// 37
	DeleteStockFromListState		// 38
	ViewListHoldingsState			// 39
	ViewStockListStatisticsState	// 40
	ViewPublicListsState			// 41
	ShareStockListState				// 42
	ViewSharedListsState			// 43
	MainReviewState					// 44
	WriteReviewState				// 45
	ViewReviewsState				// 46
)

// AppModel is the root model for the entire app
type AppModel struct {
	state        			AppState
	currentUser 			auth.User // currently logged-in user
	mainMenu    			*shared.MainMenuModel
	login       			*tuiauth.LoginModel
	register    			*tuiauth.RegisterModel
	homepage    			*shared.HomePageModel
	portfolio				*portfolio.PortfolioModel
	viewPortfolios			*portfolio.ViewPortfoliosModel
	createPortfolio			*portfolio.CreatePortfolioModel
	stockList				*stocklist.StockListModel
	stockAnalysis 			*stock.StockAnalysisModel
	social					*shared.SocialModel
	currentStocks			*stock.CurrentStocksModel
	searchStock				*stock.SearchStockModel
	stockDetails			*stock.StockDetailsModel
	sendFriReq				*friends.SendFriReqModel
	viewFriReq				*friends.ViewFriReqModel
	incFriReq				*friends.IncFriReqModel
	outFriReq 				*friends.OutFriReqModel
	viewSpecPortfolio		*portfolio.ViewSpecPortfolioModel
	viewHoldings			*stock.ViewHoldingsModel
	viewTransactions		*portfolio.ViewTransactionsModel
	buyStockSearch			*stock.BuyStockSearchModel
	buyStock				*stock.BuyStockModel
	sellStockSearch			*stock.SellStockSearchModel
	sellStock				*stock.SellStockModel
	depositCash				*portfolio.DepositCashModel
	withdrawCash			*portfolio.WithdrawCashModel
	viewNetWorth			*portfolio.ViewNetWorthModel
	manageFriends			*friends.ManageFriendsModel
	createStockList			*stocklist.CreateStockListModel
	viewMyStockList			*stocklist.ViewStockListsModel
	displayStockList		*stocklist.DisplayStockListModel
	searchStockForData		*stock.SearchStockForDataModel
	addStockData			*stock.AddStockDataModel
	viewPortfolioStats		*portfolio.ViewPortfolioStatisticsModel
	editStockList			*stocklist.EditStockListModel
	addStockToList			*stocklist.AddStockToListModel
	deleteStockFromList		*stocklist.DeleteStockFromList
	viewListHoldings		*stocklist.ViewListHoldingsModel
	holdingDetails			*portfolio.HoldingDetailsModel
	holdingForecast			*portfolio.HoldingForecastModel
	stockListStatistics		*stocklist.ViewStockListStatisticsModel
	viewPublicLists			*stocklist.ViewPublicListsModel	
	shareStockList			*stocklist.ShareStockListModel
	viewSharedLists			*stocklist.ViewSharedListsModel
	mainReview   			*stocklist.MainReviewModel
	writeReview				*stocklist.WriteReviewModel
	viewReviews				*stocklist.ViewReviewsModel
}

// NewAppModel creates a new app model
func NewAppModel() *AppModel {
	return &AppModel{
		state:     				MainMenuState,
		mainMenu:  				shared.NewMainMenu(),
		login:     				tuiauth.NewLogin(),
		register:  				tuiauth.NewRegister(),
		homepage:  				shared.NewHomePage(nil),
		portfolio: 				portfolio.NewPortfolioPage(),
		viewPortfolios:			portfolio.NewViewPortfoliosPageWithUserID(0), // Will be set with actual user ID
		createPortfolio:		portfolio.NewCreatePortfolioPageWithUserID(0), // Will be set with actual user ID
		stockList: 				stocklist.NewStockListPage(nil),
		stockAnalysis: 			stock.NewStockAnalysisPage(),
		social:					shared.NewSocialPage(nil),
		currentStocks:			stock.NewCurrentStocksPage(),
		searchStock:			stock.NewSearchStockPage(),
		stockDetails:			nil,
		sendFriReq:				friends.NewSendFriReqPage(nil),
		viewFriReq:				friends.NewViewFriReqPage(nil),
		incFriReq:				friends.NewIncFriReqPage(nil),
		outFriReq:				friends.NewOutFriReqPage(nil),
		viewSpecPortfolio:		nil,
		viewHoldings:			nil,
		viewTransactions:		nil,
		buyStockSearch:			nil,
		buyStock:				nil,
		sellStockSearch:		nil,
		sellStock:				nil,
		depositCash:			nil,
		withdrawCash:			nil,
		viewNetWorth:			nil,
		manageFriends:			friends.NewManageFriendsPage(nil),
		createStockList:		stocklist.NewCreateStockListPage(nil),
		viewMyStockList: 		stocklist.NewViewStockLists(nil),
		displayStockList: 		stocklist.NewDisplayStockListPage(DefaultStockList, nil, 0),
		searchStockForData:		stock.NewSearchStockForDataPage(),
		addStockData:			nil,
		viewPortfolioStats:		nil,
		editStockList:			stocklist.NewEditStockListPage(DefaultStockList, nil),
		addStockToList:			stocklist.NewAddStockToListPage(DefaultStockList, nil),
		deleteStockFromList:	stocklist.NewDeleteStockFromListPage(DefaultStockList, nil),
		viewListHoldings:		stocklist.NewViewListHoldingsPage(DefaultStockList, nil),
		holdingDetails:			nil,
		holdingForecast:		nil,
		stockListStatistics:	nil,
		viewPublicLists:		stocklist.NewViewPublicListsPage(nil),
		shareStockList:			stocklist.NewShareStockListPage(DefaultStockList, nil),
		viewSharedLists:		stocklist.NewViewSharedListsPage(nil),
		mainReview:				stocklist.NewMainReviewPage(DefaultStockList, nil),
		writeReview:			stocklist.NewWriteReviewPage(DefaultStockList, nil),
		viewReviews:			stocklist.NewViewReviewsPage(DefaultStockList, nil),
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
		m.mainMenu = menu.(*shared.MainMenuModel)

		// get the selected option and switch to that model if enter key pressed (confirmed)
		if m.mainMenu.Selected >= 0 && m.mainMenu.Selected < len(m.mainMenu.Options) {
			option := m.mainMenu.Options[m.mainMenu.Selected]
			if m.mainMenu.Confirmed {
				m.mainMenu.Confirmed = false
				switch option {
				case "Login":
					m.state = LoginState
					m.login = tuiauth.NewLogin()
				case "Register":
					m.state = RegisterState
					m.register = tuiauth.NewRegister()
				case "Add Stock Data":
					m.state = SearchStockForDataState
					m.searchStockForData = stock.NewSearchStockForDataPage()
					return m, m.searchStockForData.Init()
				case "Quit":
					return m, tea.Quit
				}
			}
		}
		return m, cmd

	case LoginState:
		login, cmd := m.login.Update(msg)
		m.login = login.(*tuiauth.LoginModel)
		// Check if user successfully logged in
		if m.login.GetUser() != nil {
			m.currentUser = *m.login.GetUser()
			m.state = HomePageState
			m.homepage = shared.NewHomePage(m.login.GetUser()) // reset homepage
			m.login = tuiauth.NewLogin() // reset login form
		}
		// Go back to main view from Login view
		if m.login.BackPressed {
			m.login.BackPressed = false
			m.state = MainMenuState
		}
		return m, cmd

	case RegisterState:
		register, cmd := m.register.Update(msg)
		m.register = register.(*tuiauth.RegisterModel)
		// Check if user successfully registered
		if m.register.GetUser() != nil {
			m.currentUser = *m.register.GetUser()
			m.state = HomePageState
			m.homepage = shared.NewHomePage(m.register.GetUser()) // reset homepage
			m.register = tuiauth.NewRegister() // reset register form
		}
		// Go back to main view from register view
		if m.register.BackPressed {
			m.register.BackPressed = false
			m.state = MainMenuState
		}
		return m, cmd

	case HomePageState:
		homepage, cmd := m.homepage.Update(msg)
		m.homepage = homepage.(*shared.HomePageModel)

		// get the selected option and switch to that model if enter key pressed (confirmed)
		if m.homepage.Selected >= 0 && m.homepage.Selected < len(m.homepage.Options) {
			option := m.homepage.Options[m.homepage.Selected]
			if m.homepage.Confirmed {
				m.homepage.Confirmed = false
				switch option {
				case "My Portfolios":
					m.state = PortfolioState
					m.portfolio = portfolio.NewPortfolioPage()
				case "Stock Lists":
					m.state = StockListState
					m.stockList = stocklist.NewStockListPage(m.homepage.GetUser())
				case "Stock Data & Analysis":
					m.state = StockAnalysisState
					m.stockAnalysis = stock.NewStockAnalysisPage()
				case "Friends & Social":
					m.state = SocialState
					m.social = shared.NewSocialPage(m.homepage.GetUser())
				}
			}
		}

		// Go back to main menu (logout) from HomePage
		if m.homepage.BackPressed {
			m.homepage.BackPressed = false
			m.state = MainMenuState
			m.currentUser = auth.User{} // clear current user
		}
		return m, cmd
	case PortfolioState:
		portfolioModel, cmd := m.portfolio.Update(msg)
		m.portfolio = portfolioModel.(*portfolio.PortfolioModel)

		// Check if user pressed Enter to select an option
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
			option := m.portfolio.GetSelectedOption()
			userID := int(m.currentUser.UserID)
			if option == 0 {
				// View Portfolios selected
				m.state = ViewPortfoliosState
				m.viewPortfolios = portfolio.NewViewPortfoliosPageWithUserID(userID)
				cmd = m.viewPortfolios.Init()
			} else if option == 1 {
				// Create Portfolio selected
				m.state = CreatePortfolioState
				m.createPortfolio = portfolio.NewCreatePortfolioPageWithUserID(userID)
			}
		}

		// Go back to homepage from portfolio page
		if m.portfolio.BackPressed {
			m.portfolio.BackPressed = false
			m.state = HomePageState
		}
		return m, cmd
	case ViewPortfoliosState:
		viewPortfolios, cmd := m.viewPortfolios.Update(msg)
		m.viewPortfolios = viewPortfolios.(*portfolio.ViewPortfoliosModel)

		// Check if a portfolio was selected
		if portfolioMsg, ok := msg.(portfolio.PortfolioSelectedMsg); ok {
			// View Specific Portfolio selected
			m.state = ViewSpecPortfolioState
			m.viewSpecPortfolio = portfolio.NewViewSpecPortfolioPageWithUserID(int(m.currentUser.UserID), portfolioMsg.Portfolio.PortfolioID)
			// Pass the message to the new model to populate it
			return m, func() tea.Msg {
				return portfolioMsg
			}
		}

		// Go back to portfolio page from view portfolios page
		if m.viewPortfolios.BackPressed {
			m.viewPortfolios.BackPressed = false
			m.state = PortfolioState
			m.portfolio = portfolio.NewPortfolioPage()
		}
		return m, cmd

	case ViewSpecPortfolioState:
		viewSpecPortfolio, cmd := m.viewSpecPortfolio.Update(msg)
		m.viewSpecPortfolio = viewSpecPortfolio.(*portfolio.ViewSpecPortfolioModel)

		// Handle option selection from specific portfolio view
		if m.viewSpecPortfolio.OptionSelected != "" {
			option := m.viewSpecPortfolio.OptionSelected
			m.viewSpecPortfolio.OptionSelected = "" // Reset

			switch option {
			case "View Holdings":
				m.state = ViewHoldingsState
				portfolioID := m.viewSpecPortfolio.PortfolioID
				m.viewHoldings = stock.NewViewHoldingsPageWithPortfolioID(portfolioID)
				return m, m.viewHoldings.Init()
			case "View Transactions":
				m.state = ViewTransactionsState
				portfolioID := m.viewSpecPortfolio.PortfolioID
				m.viewTransactions = portfolio.NewViewTransactionsPage(portfolioID)
				return m, m.viewTransactions.Init()
			case "Buy Stock":
				m.state = BuyStockSearchState
				userID := int(m.currentUser.UserID)
				portfolioID := m.viewSpecPortfolio.PortfolioID
				cashAccount := m.viewSpecPortfolio.Portfolio.CashAccount
				m.buyStockSearch = stock.NewBuyStockSearchPageWithPortfolio(userID, portfolioID, cashAccount)
				return m, m.buyStockSearch.Init()
			case "Sell Stock":
				m.state = SellStockSearchState
				userID := int(m.currentUser.UserID)
				portfolioID := m.viewSpecPortfolio.PortfolioID
				m.sellStockSearch = stock.NewSellStockSearchPageWithPortfolio(userID, portfolioID)
				return m, m.sellStockSearch.Init()
			case "Deposit Cash":
				m.state = DepositCashState
				userID := int(m.currentUser.UserID)
				portfolioID := m.viewSpecPortfolio.PortfolioID
				m.depositCash = portfolio.NewDepositCashPageWithPortfolioID(userID, portfolioID)
			case "Withdraw Cash":
				m.state = WithdrawCashState
				userID := int(m.currentUser.UserID)
				portfolioID := m.viewSpecPortfolio.PortfolioID
				cashAccount := m.viewSpecPortfolio.Portfolio.CashAccount
				m.withdrawCash = portfolio.NewWithdrawCashPageWithPortfolioID(userID, portfolioID, cashAccount)
			case "View Net Worth":
				m.state = ViewNetWorthState
				userID := int(m.currentUser.UserID)
				portfolioID := m.viewSpecPortfolio.PortfolioID
				m.viewNetWorth = portfolio.NewViewNetWorthPageWithPortfolioID(userID, portfolioID)
				return m, m.viewNetWorth.Init()
			case "View Statistics":
				m.state = ViewPortfolioStatisticsState
				userID := int(m.currentUser.UserID)
				portfolioID := m.viewSpecPortfolio.PortfolioID
				m.viewPortfolioStats = portfolio.NewViewPortfolioStatisticsPageWithPortfolioID(userID, portfolioID)
				return m, m.viewPortfolioStats.Init()
			}
		}

		// Go back to view portfolios page from specific portfolio view page
		if m.viewSpecPortfolio.BackPressed {
			m.viewSpecPortfolio.BackPressed = false
			m.state = ViewPortfoliosState
			userID := int(m.currentUser.UserID)
			m.viewPortfolios = portfolio.NewViewPortfoliosPageWithUserID(userID)
			cmd = m.viewPortfolios.Init()
		}
		return m, cmd
	case CreatePortfolioState:
		createPortfolio, cmd := m.createPortfolio.Update(msg)
		m.createPortfolio = createPortfolio.(*portfolio.CreatePortfolioModel)
		// Go back to portfolio page from create portfolio page
		if m.createPortfolio.BackPressed {
			m.createPortfolio.BackPressed = false
			m.state = PortfolioState
			m.portfolio = portfolio.NewPortfolioPage()
		}
		return m, cmd
	case StockListState:
		stockList, cmd := m.stockList.Update(msg)
		m.stockList = stockList.(*stocklist.StockListModel)

		if m.stockList.Selected >= 0 && m.stockList.Selected < len(m.stockList.Options) {
			option := m.stockList.Options[m.stockList.Selected]
			if m.stockList.Confirmed {
				m.stockList.Confirmed = false
				switch option {
				case "Create New Stock List":
					m.state = CreateStockListState
					m.createStockList = stocklist.NewCreateStockListPage(m.stockList.GetUser())
				case "My Stock Lists":
					m.state = ViewStockListsState
					m.viewMyStockList = stocklist.NewViewStockLists(m.stockList.GetUser())
					cmd = m.viewMyStockList.Init()
				case "Public Stock Lists":
					m.state = ViewPublicListsState
					m.viewPublicLists = stocklist.NewViewPublicListsPage(m.stockList.GetUser())
					cmd = m.viewPublicLists.Init()
				case "Shared With Me":
					m.state = ViewSharedListsState
					m.viewSharedLists = stocklist.NewViewSharedListsPage(m.stockList.GetUser())
					cmd = m.viewSharedLists.Init()
				}
			}
		}
		// Go back to homepage from stock list page
		if m.stockList.BackPressed {
			m.stockList.BackPressed = false
			m.state = HomePageState
		}
		return m, cmd
	case ViewSharedListsState:
		viewSharedLists, cmd := m.viewSharedLists.Update(msg)
		m.viewSharedLists = viewSharedLists.(*stocklist.ViewSharedListsModel)
		// Go back to stocklist page from shared stocklist page
		if m.viewSharedLists.BackPressed {
			m.viewSharedLists.BackPressed = false
			m.state = StockListState
		}
		if m.viewSharedLists.Selected >= 0 && m.viewSharedLists.Selected < len(m.viewSharedLists.StockLists) {
			selectedStockList := m.viewSharedLists.StockLists[m.viewSharedLists.Selected]
			if m.viewSharedLists.Confirmed {
				m.viewSharedLists.Confirmed = false
				currentUser := m.stockList.GetUser()
				m.displayStockList = stocklist.NewDisplayStockListPage(
            		selectedStockList,
            		currentUser,
            		selectedStockList.UserID,   
       			)
				m.state = DisplayStockListState
				return m, nil
			}
		}
		return m, cmd
	case ViewPublicListsState:
		viewPublicLists, cmd := m.viewPublicLists.Update(msg)
		m.viewPublicLists = viewPublicLists.(*stocklist.ViewPublicListsModel)
		// Go back to stocklist page from viewing public stocking list page
		if m.viewPublicLists.BackPressed {
			m.viewPublicLists.BackPressed = false
			m.state = StockListState
		}
		if m.viewPublicLists.Selected >= 0 && m.viewPublicLists.Selected < len(m.viewPublicLists.StockLists) {
			selectedStockList := m.viewPublicLists.StockLists[m.viewPublicLists.Selected]
			if m.viewPublicLists.Confirmed {
				m.viewPublicLists.Confirmed = false
				currentUser := m.stockList.GetUser()
				m.displayStockList = stocklist.NewDisplayStockListPage(
            		selectedStockList,
            		currentUser,
            		selectedStockList.UserID,   
       			)
				m.state = DisplayStockListState
				return m, nil
			}
		}
		return m, cmd
	case CreateStockListState:
		createStockList, cmd := m.createStockList.Update(msg)
		m.createStockList = createStockList.(*stocklist.CreateStockListModel)
		// Go back to main stock list page from create stock list page
		if m.createStockList.BackPressed {
			m.createStockList.BackPressed = false
			m.state = StockListState
		}
		return m, cmd

	case ViewStockListsState:
		viewMyStockList, cmd := m.viewMyStockList.Update(msg)
		m.viewMyStockList = viewMyStockList.(*stocklist.ViewStockListsModel)
		// Go back to main stock list page from create stock list page
		if m.viewMyStockList.BackPressed {
			m.viewMyStockList.BackPressed = false
			m.state = StockListState
		}
		if m.viewMyStockList.Selected >= 0 && m.viewMyStockList.Selected < len(m.viewMyStockList.StockLists) {
			selectedStockList := m.viewMyStockList.StockLists[m.viewMyStockList.Selected]
			if m.viewMyStockList.Confirmed {
				m.viewMyStockList.Confirmed = false
				currentUser := m.viewMyStockList.GetUser()
				m.displayStockList = stocklist.NewDisplayStockListPage(
            		selectedStockList,
            		currentUser,
            		currentUser.UserID,   
       			)
				m.state = DisplayStockListState
				return m, nil
			}
		}

		return m, cmd
	
	case DisplayStockListState:
		displayStockList, cmd := m.displayStockList.Update(msg)
		m.displayStockList = displayStockList.(*stocklist.DisplayStockListModel)
		// Go back to main stock list page from display stock list page
		if m.displayStockList.BackPressed {
			m.displayStockList.BackPressed = false
			m.state = StockListState
		}
		if m.displayStockList.Selected >= 0 && m.displayStockList.Selected < len(m.displayStockList.Options) {
			option := m.displayStockList.Options[m.displayStockList.Selected]
			if m.displayStockList.Confirmed {
				m.displayStockList.Confirmed = false
				switch option {
				case "Edit List":
					m.state = EditStockListState
					m.editStockList = stocklist.NewEditStockListPage(m.displayStockList.StockList, m.stockList.GetUser())
				case "View Stocks":
					m.state = ViewListHoldingsState
					m.viewListHoldings = stocklist.NewViewListHoldingsPage(m.displayStockList.StockList, m.stockList.GetUser())
					cmd = m.viewListHoldings.Init()
				case "Delete List":
					selectedID := m.displayStockList.StockList.StockListID
					err := stocklist.DeleteStockList(selectedID)
					if err != nil {
						m.displayStockList.Error = fmt.Sprintf("Failed to delete '%s': %v", m.displayStockList.StockList.Name, err)
					} else {
						m.displayStockList.SuccessMessage = fmt.Sprintf("'%s' deleted successfully", m.displayStockList.StockList.Name)
						m.displayStockList.DeleteList = true
						m.state = StockListState
						m.displayStockList.SuccessMessage = ""
						m.displayStockList.DeleteList  = false
					}
					return m, cmd
				case "Share":
					m.state = ShareStockListState
					m.shareStockList = stocklist.NewShareStockListPage(m.displayStockList.StockList, m.stockList.GetUser())
				case "Reviews":
					m.state = MainReviewState
					m.mainReview = stocklist.NewMainReviewPage(m.displayStockList.StockList, m.stockList.GetUser())
				}
			}
		}
		return m, cmd
	case MainReviewState:
		mainReview, cmd := m.mainReview.Update(msg)
		m.mainReview = mainReview.(*stocklist.MainReviewModel)
		// Go back to display stock list page from main review page
		if m.mainReview.BackPressed {
			m.mainReview.BackPressed = false
			m.state = DisplayStockListState
		}
		if m.mainReview.Selected >= 0 && m.mainReview.Selected < len(m.mainReview.Options) {
			option := m.mainReview.Options[m.mainReview.Selected]
			if m.mainReview.Confirmed {
				m.mainReview.Confirmed = false
				switch option {
				case "Write/Edit Review":
					m.state = WriteReviewState
					m.writeReview = stocklist.NewWriteReviewPage(m.mainReview.StockList, m.mainReview.User)
				case "View Reviews":
					m.state = ViewReviewsState
					m.viewReviews = stocklist.NewViewReviewsPage(m.mainReview.StockList, m.mainReview.User)
					cmd = m.viewReviews.Init()
				}
			}
		}
		return m, cmd
	case ViewReviewsState:
		viewReviews, cmd := m.viewReviews.Update(msg)
		m.viewReviews = viewReviews.(*stocklist.ViewReviewsModel)
		// Go back to main review page from write reviews
		if m.viewReviews.BackPressed {
			m.viewReviews.BackPressed = false
			m.state = MainReviewState
		}
		return m, cmd

	case WriteReviewState:
		writeReview, cmd := m.writeReview.Update(msg)
		m.writeReview = writeReview.(*stocklist.WriteReviewModel)
		// Go back to main review page from write reviews
		if m.writeReview.BackPressed {
			m.writeReview.BackPressed = false
			m.state = MainReviewState
		}
		return m, cmd

	case ShareStockListState:
		shareStockList, cmd := m.shareStockList.Update(msg)
		m.shareStockList = shareStockList.(*stocklist.ShareStockListModel)
		// Go back to display stock list page from share page
		if m.shareStockList.BackPressed {
			m.shareStockList.BackPressed = false
			m.state = DisplayStockListState
		}
		return m, cmd

	case ViewListHoldingsState:
		viewListHoldings, cmd := m.viewListHoldings.Update(msg)
		m.viewListHoldings = viewListHoldings.(*stocklist.ViewListHoldingsModel)
		// Go back to display stock list page from view stock holdings page
		if m.viewListHoldings.BackPressed {
			m.viewListHoldings.BackPressed = false
			m.state = DisplayStockListState
		}
		// Go to stock list statistics
		if m.viewListHoldings.ViewStatsPressed {
			m.viewListHoldings.ViewStatsPressed = false
			m.state = ViewStockListStatisticsState
			m.stockListStatistics = stocklist.NewViewStockListStatisticsPage(m.viewListHoldings.StockList.StockListID)
			cmd = m.stockListStatistics.Init()
		}
		return m, cmd
	case EditStockListState:
		editStockList, cmd := m.editStockList.Update(msg)
		m.editStockList = editStockList.(*stocklist.EditStockListModel)
		// Go back to display stock list page from edit stock list page
		if m.editStockList.BackPressed {
			m.editStockList.BackPressed = false
			m.state = DisplayStockListState
		}
		if m.editStockList.Selected >= 0 && m.editStockList.Selected < len(m.editStockList.Options) {
			option := m.editStockList.Options[m.editStockList.Selected]
			if m.editStockList.Confirmed {
				m.editStockList.Confirmed = false
				switch option {
				case "Add/Update Stock":
					m.state = AddStockToListState
					m.addStockToList = stocklist.NewAddStockToListPage(m.editStockList.StockList, m.stockList.GetUser())
				case "Delete Stock":
					m.state = DeleteStockFromListState
					m.deleteStockFromList = stocklist.NewDeleteStockFromListPage(m.editStockList.StockList, m.stockList.GetUser())
					cmd = m.deleteStockFromList.Init()
				}
			}
		}
		return m, cmd
	case AddStockToListState:
		addStockToList, cmd := m.addStockToList.Update(msg)
		m.addStockToList = addStockToList.(*stocklist.AddStockToListModel)
		// Go back to edit stock list page from add stock 
		if m.addStockToList.BackPressed {
			m.addStockToList.BackPressed = false
			m.state = EditStockListState
		}
		return m, cmd

	case DeleteStockFromListState:
		deleteStockFromList, cmd := m.deleteStockFromList.Update(msg)
		m.deleteStockFromList = deleteStockFromList.(*stocklist.DeleteStockFromList)
		// Go back to edit stock list page from delete stock 
		if m.deleteStockFromList.BackPressed {
			m.deleteStockFromList.BackPressed = false
			m.state = EditStockListState
		}
		return m, cmd

	case SearchStockForDataState:
		searchStockForData, cmd := m.searchStockForData.Update(msg)
		m.searchStockForData = searchStockForData.(*stock.SearchStockForDataModel)

		// handle stock selection
		if m.searchStockForData.BackPressed {
			m.searchStockForData.BackPressed = false
			m.state = MainMenuState
			m.mainMenu = shared.NewMainMenu()
		} else if msg, ok := msg.(stock.StockSelectedForDataMsg); ok {
			m.state = AddStockDataState
			m.addStockData = stock.NewAddStockDataPageWithSymbol(msg.Stock.Symbol)
		}
		return m, cmd

	case AddStockDataState:
		// Handle completion messages first
		if completedMsg, ok := msg.(stock.AddStockDataCompletedMsg); ok {
			if completedMsg.Success {
				m.addStockData.Error = ""
				// Show success and go back to stock selection
				m.state = SearchStockForDataState
				m.searchStockForData = stock.NewSearchStockForDataPage()
				return m, m.searchStockForData.Init()
			} else {
				m.addStockData.Error = completedMsg.Message
			}
			return m, nil
		}

		addStockData, cmd := m.addStockData.Update(msg)
		m.addStockData = addStockData.(*stock.AddStockDataModel)

		// Handle form submission
		if m.addStockData.BackPressed {
			m.addStockData.BackPressed = false
			m.state = SearchStockForDataState
			m.searchStockForData = stock.NewSearchStockForDataPage()
			return m, m.searchStockForData.Init()
		} else if m.addStockData.Confirmed {
			m.addStockData.Confirmed = false
			// Parse and submit the data
			date, open, high, low, close, volume, err := m.addStockData.GetData()
			if err != nil {
				m.addStockData.Error = fmt.Sprintf("Invalid data: %v", err)
				return m, nil
			}

			// Call database function asynchronously
			return m, func() tea.Msg {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				dbService := database.New()
				err := dbService.AddStockData(ctx, m.addStockData.Symbol, date, open, high, low, close, volume)
				if err != nil {
					return stock.AddStockDataCompletedMsg{
						Success: false,
						Message: fmt.Sprintf("Error adding stock data: %v", err),
					}
				}

				return stock.AddStockDataCompletedMsg{
					Symbol:  m.addStockData.Symbol,
					Date:    date,
					Open:    open,
					High:    high,
					Low:     low,
					Close:   close,
					Volume:  volume,
					Success: true,
					Message: fmt.Sprintf("Successfully added data for %s on %s", m.addStockData.Symbol, date),
				}
			}
		}
		return m, cmd

	case ViewPortfolioStatisticsState:
		viewPortfolioStats, cmd := m.viewPortfolioStats.Update(msg)
		m.viewPortfolioStats = viewPortfolioStats.(*portfolio.ViewPortfolioStatisticsModel)

		// Go back to specific portfolio view
		if m.viewPortfolioStats.BackPressed {
			m.viewPortfolioStats.BackPressed = false
			m.state = ViewSpecPortfolioState
		}
		return m, cmd

	case HoldingDetailsState:
		holdingDetails, cmd := m.holdingDetails.Update(msg)
		m.holdingDetails = holdingDetails.(*portfolio.HoldingDetailsModel)

		// View forecast of selected holding
		if m.holdingDetails.ViewForecastPressed {
			m.holdingDetails.ViewForecastPressed = false
			m.state = HoldingForecastState
			m.holdingForecast = portfolio.NewHoldingForecastPage(
				m.holdingDetails.Symbol,
				m.holdingDetails.CurrentPrice,
			)
			return m, m.holdingForecast.Init()
		}

		// Go back to view holdings
		if m.holdingDetails.BackPressed {
			m.holdingDetails.BackPressed = false
			m.state = ViewHoldingsState
		}
		return m, cmd

	case HoldingForecastState:
		holdingForecast, cmd := m.holdingForecast.Update(msg)
		m.holdingForecast = holdingForecast.(*portfolio.HoldingForecastModel)

		// Go back to holding details
		if m.holdingForecast.BackPressed {
			m.holdingForecast.BackPressed = false
			m.state = HoldingDetailsState
		}
		return m, cmd

	case ViewStockListStatisticsState:
		stockListStats, cmd := m.stockListStatistics.Update(msg)
		m.stockListStatistics = stockListStats.(*stocklist.ViewStockListStatisticsModel)

		// Go back to view list holdings
		if m.stockListStatistics.BackPressed {
			m.stockListStatistics.BackPressed = false
			m.state = ViewListHoldingsState
		}
		return m, cmd

	case StockAnalysisState:
		stockAnalysis, cmd := m.stockAnalysis.Update(msg)
		m.stockAnalysis = stockAnalysis.(*stock.StockAnalysisModel)

		// Handle menu option selection
		if m.stockAnalysis.Selected >= 0 && m.stockAnalysis.Selected < len(m.stockAnalysis.Options) {
			option := m.stockAnalysis.Options[m.stockAnalysis.Selected]
			if m.stockAnalysis.Confirmed {
				m.stockAnalysis.Confirmed = false
				switch option {
				case "View Current Stocks":
					m.state = CurrentStocksState
					m.currentStocks = stock.NewCurrentStocksPage()
					// Return Init command to fetch stocks data
					cmd = m.currentStocks.Init()
				case "Search Stock":
					m.state = SearchStockState
					m.searchStock = stock.NewSearchStockPage()
				}
			}
		}

		// Go back to homepage from Stock Data & Analysis page
		if m.stockAnalysis.BackPressed {
			m.stockAnalysis.BackPressed = false
			m.state = HomePageState
		}
		return m, cmd
	case SocialState:
		social, cmd := m.social.Update(msg)
		m.social = social.(*shared.SocialModel)

		// get the selected option and switch to that model if enter key pressed (confirmed)
		if m.social.Selected >= 0 && m.social.Selected < len(m.social.Options) {
			option := m.social.Options[m.social.Selected]
			if m.social.Confirmed {
				m.social.Confirmed = false
				switch option {
				case "Send Friend Request":
					m.state = SendFriReqState
					m.sendFriReq = friends.NewSendFriReqPage(m.social.GetUser())
				case "View Friends Requests":
					m.state = ViewFriReqState
					m.viewFriReq = friends.NewViewFriReqPage(m.social.GetUser())
				case "Manage Friends":
					m.state = ManageFriendsState
					m.manageFriends = friends.NewManageFriendsPage(m.social.GetUser())
					cmd = m.manageFriends.Init()
				}
			}
		}
		// Go back to homepage from social page
		if m.social.BackPressed {
			m.social.BackPressed = false
			m.state = HomePageState
		}
		return m, cmd

	case CurrentStocksState:
		currentStocks, cmd := m.currentStocks.Update(msg)
		m.currentStocks = currentStocks.(*stock.CurrentStocksModel)
		// Go back to stock analysis page from current stocks page
		if m.currentStocks.BackPressed {
			m.currentStocks.BackPressed = false
			m.state = StockAnalysisState
			m.stockAnalysis = stock.NewStockAnalysisPage()
			return m, m.stockAnalysis.Init()
		}
		return m, cmd
	case SearchStockState:
		searchStock, cmd := m.searchStock.Update(msg)
		m.searchStock = searchStock.(*stock.SearchStockModel)
		// Check if user confirmed search
		if m.searchStock.Confirmed {
			m.searchStock.Confirmed = false
			symbol := m.searchStock.GetSymbol()
			m.state = StockDetailsState
			m.stockDetails = stock.NewStockDetailsPage(symbol)
			// Return Init command to fetch historical data
			cmd = m.stockDetails.Init()
		}
		// Go back to stock analysis page
		if m.searchStock.BackPressed {
			m.searchStock.BackPressed = false
			m.state = StockAnalysisState
			m.searchStock = stock.NewSearchStockPage()
		}
		return m, cmd
	case StockDetailsState:
		stockDetails, cmd := m.stockDetails.Update(msg)
		m.stockDetails = stockDetails.(*stock.StockDetailsModel)
		// Go back to search stock page
		if m.stockDetails.BackPressed {
			m.stockDetails.BackPressed = false
			m.state = SearchStockState
			m.searchStock = stock.NewSearchStockPage()
		}
		return m, cmd
	case SendFriReqState:
		sendFriReq, cmd := m.sendFriReq.Update(msg)
		m.sendFriReq = sendFriReq.(*friends.SendFriReqModel)
		// Go back to friend and social from send friend request page
		if m.sendFriReq.BackPressed {
			m.sendFriReq.BackPressed = false
			m.state = SocialState
		}
		return m, cmd
	case ViewFriReqState:
		viewFriReq, cmd := m.viewFriReq.Update(msg)
		m.viewFriReq = viewFriReq.(*friends.ViewFriReqModel)
		// Go back to friend and social from view friend request page
		if m.viewFriReq.BackPressed {
			m.viewFriReq.BackPressed = false
			m.state = SocialState
		}
		// get the selected option and switch to that model if enter key pressed (confirmed)
		if m.viewFriReq.Selected >= 0 && m.viewFriReq.Selected < len(m.viewFriReq.Options) {
			option := m.viewFriReq.Options[m.viewFriReq.Selected]
			if m.viewFriReq.Confirmed {
				m.viewFriReq.Confirmed = false
				switch option {
				case "Incoming Requests (Accept / Reject)":
					m.state = IncFriReqState
					m.incFriReq = friends.NewIncFriReqPage(m.viewFriReq.GetUser())
					cmd = m.incFriReq.Init()
				case "Outgoing Requests (Cancel)":
					m.state = OutFriReqState
					m.outFriReq = friends.NewOutFriReqPage(m.viewFriReq.GetUser())
					cmd = m.outFriReq.Init()
				}
			}
		}
		return m, cmd
	case IncFriReqState:
		incFriReq, cmd := m.incFriReq.Update(msg)
		m.incFriReq = incFriReq.(*friends.IncFriReqModel)
		// Go back to view friend request page from incoming view page
		if m.incFriReq.BackPressed {
			m.incFriReq.BackPressed = false
			m.state = ViewFriReqState
		}
		return m, cmd
	case OutFriReqState:
		outFriReq, cmd := m.outFriReq.Update(msg)
		m.outFriReq = outFriReq.(*friends.OutFriReqModel)
		// Go back to view friend request page from outgoing view page
		if m.outFriReq.BackPressed {
			m.outFriReq.BackPressed = false
			m.state = ViewFriReqState
		}
		return m, cmd
	case ViewHoldingsState:
		viewHoldings, cmd := m.viewHoldings.Update(msg)
		m.viewHoldings = viewHoldings.(*stock.ViewHoldingsModel)

		// View details of selected holding
		if m.viewHoldings.ViewDetailsPressed {
			m.viewHoldings.ViewDetailsPressed = false
			holding := m.viewHoldings.GetSelectedHolding()
			m.state = HoldingDetailsState
			m.holdingDetails = portfolio.NewHoldingDetailsPage(
				m.viewHoldings.PortfolioID,
				holding.Symbol,
				holding.Shares,
				holding.Price,
			)
			return m, m.holdingDetails.Init()
		}

		// Go back to specific portfolio view from view holdings
		if m.viewHoldings.BackPressed {
			m.viewHoldings.BackPressed = false
			m.state = ViewSpecPortfolioState
		}
		return m, cmd

	case ViewTransactionsState:
		viewTransactions, cmd := m.viewTransactions.Update(msg)
		m.viewTransactions = viewTransactions.(*portfolio.ViewTransactionsModel)

		// Go back to specific portfolio view from view transactions
		if m.viewTransactions.BackPressed {
			m.viewTransactions.BackPressed = false
			m.state = ViewSpecPortfolioState
		}
		return m, cmd

	case BuyStockSearchState:
		buyStockSearch, cmd := m.buyStockSearch.Update(msg)
		m.buyStockSearch = buyStockSearch.(*stock.BuyStockSearchModel)

		// Check if a stock was selected
		if stockMsg, ok := msg.(stock.StockSelectedForBuyMsg); ok {
			m.state = BuyStockState
			userID := int(m.currentUser.UserID)
			portfolioID := m.buyStockSearch.PortfolioID
			cashAccount := m.buyStockSearch.CashAccount
			m.buyStock = stock.NewBuyStockPageWithStock(userID, portfolioID, stockMsg.Stock, cashAccount)
		}

		// Go back to specific portfolio view from buy stock search
		if m.buyStockSearch.BackPressed {
			m.buyStockSearch.BackPressed = false
			m.state = ViewSpecPortfolioState
		}
		return m, cmd
	case BuyStockState:
		buyStock, cmd := m.buyStock.Update(msg)
		m.buyStock = buyStock.(*stock.BuyStockModel)

		// Handle confirmed purchase
		if m.buyStock.Confirmed {
			m.buyStock.Confirmed = false
			// Execute the buy transaction asynchronously
			return m, func() tea.Msg {
				db := database.New()
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				// Convert portfolioID from string to int
				portfolioID := 0
				fmt.Sscanf(m.buyStock.PortfolioID, "%d", &portfolioID)

				err := db.BuyStock(
					ctx,
					portfolioID,
					m.buyStock.Stock.Symbol,
					m.buyStock.GetQuantity(),
					m.buyStock.Stock.Price,
				)

				if err != nil {
					return stock.BuyStockCompletedMsg{
						Success: false,
						Message: "✗ " + err.Error(),
					}
				}

				return stock.BuyStockCompletedMsg{
					Success: true,
					Message: "✓ Stock purchase completed!",
				}
			}
		}

		// Handle buy completion message
		if completedMsg, ok := msg.(stock.BuyStockCompletedMsg); ok {
			if completedMsg.Success {
				// Purchase successful, go back to portfolio view and refresh data
				m.state = ViewSpecPortfolioState
				return m, m.viewSpecPortfolio.RefreshPortfolio()
			} else {
				// Show error message in the buy stock page
				m.buyStock.Error = completedMsg.Message
			}
		}

		// Go back to stock search from buy stock page
		if m.buyStock.BackPressed {
			m.buyStock.BackPressed = false
			m.state = BuyStockSearchState
		}

		return m, cmd
	case SellStockSearchState:
		sellStockSearch, cmd := m.sellStockSearch.Update(msg)
		m.sellStockSearch = sellStockSearch.(*stock.SellStockSearchModel)

		// Check if a stock was selected
		if stockMsg, ok := msg.(stock.StockSelectedForSellMsg); ok {
			m.state = SellStockState
			userID := int(m.currentUser.UserID)
			portfolioID := m.sellStockSearch.PortfolioID
			m.sellStock = stock.NewSellStockPageWithHolding(userID, portfolioID, stockMsg.Holding)
		}

		// Go back to specific portfolio view from sell stock search
		if m.sellStockSearch.BackPressed {
			m.sellStockSearch.BackPressed = false
			m.state = ViewSpecPortfolioState
		}
		return m, cmd
	case SellStockState:
		sellStock, cmd := m.sellStock.Update(msg)
		m.sellStock = sellStock.(*stock.SellStockModel)

		// Handle confirmed sale
		if m.sellStock.Confirmed {
			m.sellStock.Confirmed = false
			// Execute the sell transaction asynchronously
			return m, func() tea.Msg {
				db := database.New()
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				// Convert portfolioID from string to int
				portfolioID := 0
				fmt.Sscanf(m.sellStock.PortfolioID, "%d", &portfolioID)

				err := db.SellStock(
					ctx,
					portfolioID,
					m.sellStock.GetHolding().Symbol,
					m.sellStock.GetQuantity(),
					m.sellStock.GetHolding().Price,
				)

				if err != nil {
					return stock.SellStockCompletedMsg{
						Success: false,
						Message: "✗ " + err.Error(),
					}
				}

				return stock.SellStockCompletedMsg{
					Success: true,
					Message: "✓ Stock sale completed!",
				}
			}
		}

		// Handle sale completion message
		if completedMsg, ok := msg.(stock.SellStockCompletedMsg); ok {
			if completedMsg.Success {
				// Sale successful, go back to portfolio view and refresh data
				m.state = ViewSpecPortfolioState
				return m, m.viewSpecPortfolio.RefreshPortfolio()
			} else {
				// Show error message in the sell stock page
				m.sellStock.Error = completedMsg.Message
			}
		}

		// Go back to stock search from sell stock page
		if m.sellStock.BackPressed {
			m.sellStock.BackPressed = false
			m.state = SellStockSearchState
		}

		return m, cmd
	case DepositCashState:
		depositCash, cmd := m.depositCash.Update(msg)
		m.depositCash = depositCash.(*portfolio.DepositCashModel)

		// Handle confirmed deposit
		if m.depositCash.Confirmed {
			m.depositCash.Confirmed = false
			// Execute the deposit asynchronously
			return m, func() tea.Msg {
				db := database.New()
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				// Convert portfolioID from string to int
				portfolioID := 0
				fmt.Sscanf(m.depositCash.PortfolioID, "%d", &portfolioID)

				err := db.DepositCash(ctx, portfolioID, m.depositCash.GetAmount())

				if err != nil {
					return portfolio.DepositCashCompletedMsg{
						Success: false,
						Message: "✗ " + err.Error(),
					}
				}

				return portfolio.DepositCashCompletedMsg{
					Success: true,
					Message: "✓ Deposit completed!",
				}
			}
		}

		// Handle deposit completion message
		if completedMsg, ok := msg.(portfolio.DepositCashCompletedMsg); ok {
			if completedMsg.Success {
				// Deposit successful, go back to portfolio view and refresh data
				m.state = ViewSpecPortfolioState
				return m, m.viewSpecPortfolio.RefreshPortfolio()
			} else {
				// Show error message in the deposit page
				m.depositCash.Error = completedMsg.Message
			}
		}

		// Go back to portfolio view from deposit page
		if m.depositCash.BackPressed {
			m.depositCash.BackPressed = false
			m.state = ViewSpecPortfolioState
		}

		return m, cmd
	case WithdrawCashState:
		withdrawCash, cmd := m.withdrawCash.Update(msg)
		m.withdrawCash = withdrawCash.(*portfolio.WithdrawCashModel)

		// Handle confirmed withdrawal
		if m.withdrawCash.Confirmed {
			m.withdrawCash.Confirmed = false
			// Execute the withdrawal asynchronously
			return m, func() tea.Msg {
				db := database.New()
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				// Convert portfolioID from string to int
				portfolioID := 0
				fmt.Sscanf(m.withdrawCash.PortfolioID, "%d", &portfolioID)

				err := db.WithdrawCash(ctx, portfolioID, m.withdrawCash.GetAmount())

				if err != nil {
					return portfolio.WithdrawCashCompletedMsg{
						Success: false,
						Message: "✗ " + err.Error(),
					}
				}

				return portfolio.WithdrawCashCompletedMsg{
					Success: true,
					Message: "✓ Withdrawal completed!",
				}
			}
		}

		// Handle withdrawal completion message
		if completedMsg, ok := msg.(portfolio.WithdrawCashCompletedMsg); ok {
			if completedMsg.Success {
				// Withdrawal successful, go back to portfolio view and refresh data
				m.state = ViewSpecPortfolioState
				return m, m.viewSpecPortfolio.RefreshPortfolio()
			} else {
				// Show error message in the withdraw page
				m.withdrawCash.Error = completedMsg.Message
			}
		}

		// Go back to portfolio view from withdraw page
		if m.withdrawCash.BackPressed {
			m.withdrawCash.BackPressed = false
			m.state = ViewSpecPortfolioState
		}

		return m, cmd
	case ViewNetWorthState:
		viewNetWorth, cmd := m.viewNetWorth.Update(msg)
		m.viewNetWorth = viewNetWorth.(*portfolio.ViewNetWorthModel)

		// Go back to portfolio view from net worth page
		if m.viewNetWorth.BackPressed {
			m.viewNetWorth.BackPressed = false
			m.state = ViewSpecPortfolioState
		}

		return m, cmd
	case ManageFriendsState:
		manageFriends, cmd := m.manageFriends.Update(msg)
		m.manageFriends = manageFriends.(*friends.ManageFriendsModel)
		// Go back to friend & social page from manage friends page
		if m.manageFriends.BackPressed {
			m.manageFriends.BackPressed = false
			m.state = SocialState
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
	case HomePageState:
		return m.homepage.View()
	case PortfolioState:
		return m.portfolio.View()
	case ViewPortfoliosState:
		return m.viewPortfolios.View()
	case CreatePortfolioState:
		return m.createPortfolio.View()
	case StockListState:
		return m.stockList.View()
	case StockAnalysisState:
		return m.stockAnalysis.View()
	case SocialState:
		return m.social.View()
	case CurrentStocksState:
		return m.currentStocks.View()
	case SearchStockState:
		return m.searchStock.View()
	case StockDetailsState:
		return m.stockDetails.View()
	case SendFriReqState:
		return m.sendFriReq.View()
	case ViewFriReqState:
		return m.viewFriReq.View()
	case IncFriReqState:
		return m.incFriReq.View()
	case OutFriReqState:
		return m.outFriReq.View()
	case ViewSpecPortfolioState:
		return m.viewSpecPortfolio.View()
	case ViewHoldingsState:
		return m.viewHoldings.View()
	case ViewTransactionsState:
		return m.viewTransactions.View()
	case BuyStockSearchState:
		return m.buyStockSearch.View()
	case BuyStockState:
		return m.buyStock.View()
	case SellStockSearchState:
		return m.sellStockSearch.View()
	case SellStockState:
		return m.sellStock.View()
	case DepositCashState:
		return m.depositCash.View()
	case WithdrawCashState:
		return m.withdrawCash.View()
	case ViewNetWorthState:
		return m.viewNetWorth.View()
	case ManageFriendsState:
		return m.manageFriends.View()
	case CreateStockListState:
		return m.createStockList.View()
	case ViewStockListsState:
		return m.viewMyStockList.View()
	case SearchStockForDataState:
		return m.searchStockForData.View()
	case AddStockDataState:
		return m.addStockData.View()
	case ViewPortfolioStatisticsState:
		return m.viewPortfolioStats.View()
	case DisplayStockListState:
		return m.displayStockList.View()
	case EditStockListState:
		return m.editStockList.View()
	case AddStockToListState:
		return m.addStockToList.View()
	case DeleteStockFromListState:
		return m.deleteStockFromList.View()
	case ViewListHoldingsState:
		return m.viewListHoldings.View()
	case HoldingDetailsState:
		return m.holdingDetails.View()
	case HoldingForecastState:
		return m.holdingForecast.View()
	case ViewStockListStatisticsState:
		return m.stockListStatistics.View()
	case ViewPublicListsState:
		return m.viewPublicLists.View()
	case ShareStockListState:
		return m.shareStockList.View()
	case ViewSharedListsState:
		return m.viewSharedLists.View()
	case MainReviewState:
		return m.mainReview.View()
	case WriteReviewState:
		return m.writeReview.View()
	case ViewReviewsState:
		return m.viewReviews.View()
	}
	return ""
}

// StartApp starts the interactive CLI app
func StartApp() error {
	p := tea.NewProgram(NewAppModel())
	_, err := p.Run()
	return err
}

