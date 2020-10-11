package submission_controller

import (
	"clx/browser"
	"clx/cli"
	commentparser "clx/comment-parser"
	"encoding/json"
	text "github.com/MichaelMure/go-term-text"
	"github.com/gdamore/tcell"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
	"gitlab.com/tslocum/cview"
	"strconv"
)

const (
	maximumStoriesToDisplay = 30
)

type submissionHandler struct {
	Submissions                 []Submission
	MappedSubmissions           int
	MappedPages                 int
	StoriesListed               int
	Pages                       *cview.Pages
	Application                 *cview.Application
	PageToFetchFromAPI          int
	CurrentPage                 int
	ScreenHeight                int
	ViewableStoriesOnSinglePage int
	MaxPages                    int
}

func NewSubmissionHandler() *submissionHandler {
	sh := new(submissionHandler)
	sh.Application = cview.NewApplication()
	sh.setShortcuts()
	sh.Pages = cview.NewPages()
	sh.MaxPages = 2
	sh.ScreenHeight = getTerminalHeight()
	sh.ViewableStoriesOnSinglePage = min(sh.ScreenHeight/2, maximumStoriesToDisplay)
	sh.fetchSubmissions()

	sh.Pages.SwitchToPage("0")


	helpScreen := getHelpScreen()

	sh.Pages.AddPage("help", helpScreen, true, false)


	return sh
}

func getHelpScreen() *cview.TextView {
	helpScreen := cview.NewTextView()
	helpScreen.SetBackgroundColor(tcell.ColorDefault)
	helpScreen.SetTextColor(tcell.ColorDefault)
	helpScreen.SetTextAlign(cview.AlignLeft)
	helpScreen.SetTitle("circumflex")
	helpScreen.SetTitleColor(tcell.ColorDefault)
	helpScreen.SetBorderColor(tcell.ColorDefault)
	helpScreen.SetTextColor(tcell.ColorDefault)
	helpScreen.Box.SetBorderPadding(10, 10, 10, 10)
	helpScreen.Box.SetBorder(true)
	helpScreen.Box.SetBorderAttributes(tcell.AttrDim)

	newLine := "\n"
	t := ""
	t += "j, ↓:          down" + newLine
	t += "h, ↑:          up" + newLine
	t += newLine
	t += "Enter:         read comments" + newLine
	t += "o:             open submission in browser" + newLine
	t += "q:             quit" + newLine
	t += "h:             bring up this screen" + newLine
	t += newLine
	t += "Ctrl + n:      next page" + newLine
	t += "Ctrl + p:      previous page" + newLine

	x, _ := terminal.Width()
	width := int(x)


	helpScreen.SetText(text.LeftPadLines(t, width / 4))

	return helpScreen
}

func (sh *submissionHandler) setShortcuts() {
	app := sh.Application
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentPage, _ := sh.Pages.GetFrontPage()
		if currentPage == "help" {
			//TODO: set to last viewed page
			sh.Pages.SwitchToPage("0")
			return event
		}

		if event.Key() == tcell.KeyCtrlN {
			sh.nextPage()
		} else if event.Key() == tcell.KeyCtrlP {
			sh.previousPage()
		} else if event.Rune() == 'q' {
			app.Stop()
		} else if event.Rune() == 'h' {
			sh.Pages.SwitchToPage("help")
		}
		return event
	})
}

func getTerminalHeight() int {
	y, _ := terminal.Height()
	return int(y)
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func (sh *submissionHandler) nextPage() {
	nextPage := sh.CurrentPage + 1

	if nextPage > sh.MaxPages {
		return
	}

	_, primitive := sh.Pages.GetFrontPage()
	list := primitive.(*cview.List)
	currentlySelectedItem := list.GetCurrentItem()

	if nextPage < sh.MappedPages {
		sh.Pages.SwitchToPage(strconv.Itoa(nextPage))
		_, p := sh.Pages.GetFrontPage()
		l := p.(*cview.List)
		l.SetCurrentItem(currentlySelectedItem)
	} else {
		sh.fetchSubmissions()
		sh.Pages.SwitchToPage(strconv.Itoa(nextPage))
	}

	sh.CurrentPage++
}

func (sh *submissionHandler) previousPage() {
	previousPage := sh.CurrentPage - 1

	if previousPage < 0 {
		return
	}

	_, primitive := sh.Pages.GetFrontPage()
	list := primitive.(*cview.List)
	currentlySelectedItem := list.GetCurrentItem()

	sh.CurrentPage--
	sh.Pages.SwitchToPage(strconv.Itoa(sh.CurrentPage))

	_, p := sh.Pages.GetFrontPage()
	l := p.(*cview.List)
	l.SetCurrentItem(currentlySelectedItem)
}

func (sh *submissionHandler) getStoriesToDisplay() int {
	return sh.ViewableStoriesOnSinglePage
}

func setSelectedFunction(app *cview.Application, list *cview.List, sh *submissionHandler) {
	list.SetSelectedFunc(func(i int, a string, b string, c rune) {
		app.Suspend(func() {
			for index := range sh.Submissions {
				if index == i {
					id := getSubmissionID(i, sh)
					JSON, _ := get("http://node-hnapi.herokuapp.com/item/" + id)
					jComments := new(commentparser.Comments)
					_ = json.Unmarshal(JSON, jComments)

					commentTree := commentparser.PrintCommentTree(*jComments, 4, 70)
					cli.Less(commentTree)
				}
			}
		})
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'o' {
			item := list.GetCurrentItem()
			url := sh.Submissions[item].URL
			browser.Open(url)
		}
		return event
	})
}

func getSubmissionID(i int, sh *submissionHandler) string {
	storyIndex := (sh.CurrentPage)*sh.ViewableStoriesOnSinglePage + i
	s := sh.Submissions[storyIndex]
	return strconv.Itoa(s.ID)
}

func (sh *submissionHandler) getSubmission(i int) Submission {
	return sh.Submissions[i]
}

type Submission struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Points        int    `json:"points"`
	Author        string `json:"user"`
	Time          string `json:"time_ago"`
	CommentsCount int    `json:"comments_count"`
	URL           string `json:"url"`
	Domain        string `json:"domain"`
	Type          string `json:"type"`
}

func (sh *submissionHandler) fetchSubmissions() {
	sh.PageToFetchFromAPI++
	p := strconv.Itoa(sh.PageToFetchFromAPI)
	submissions := getSubmissions("http://node-hnapi.herokuapp.com/news?page=" + p)
	sh.Submissions = append(sh.Submissions, submissions...)
	sh.mapSubmissionsToListItems()
}

func (sh *submissionHandler) mapSubmissionsToListItems() {
	for sh.hasStoriesToMap() {
		sub := sh.Submissions[sh.MappedSubmissions : sh.MappedSubmissions+sh.ViewableStoriesOnSinglePage]
		list := createNewList(sh)
		addSubmissionsToList(list, sub, sh)

		sh.Pages.AddPage(strconv.Itoa(sh.MappedPages), list, true, true)
		sh.MappedPages++
	}
}

func (sh *submissionHandler) hasStoriesToMap() bool {
	return len(sh.Submissions)-sh.MappedSubmissions > sh.ViewableStoriesOnSinglePage
}

func createNewList(sh *submissionHandler) *cview.List {
	list := cview.NewList()
	list.SetBackgroundTransparent(false)
	list.SetBackgroundColor(tcell.ColorDefault)
	list.SetMainTextColor(tcell.ColorDefault)
	list.SetSecondaryTextColor(tcell.ColorDefault)
	list.ShowSecondaryText(true)
	setSelectedFunction(sh.Application, list, sh)

	return list
}

func addSubmissionsToList(list *cview.List, submissions []Submission, sh *submissionHandler) {
	for _, submission := range submissions {
		list.AddItem(
			submission.getMainText(sh.MappedSubmissions),
			submission.getSecondaryText(),
			0,
			nil,
		)
		sh.MappedSubmissions++
	}
}

func (s Submission) getMainText(i int) string {
	rank := i + 1
	return strconv.Itoa(rank) + "." + getRankIndentBlock(rank) + s.Title + s.GetDomain()
}

func (s Submission) getSecondaryText() string {
	return "[::d]" + "    " + s.getPoints() + " points by " + s.Author + " " +
		s.Time + " | " + s.getComments() + " comments" + "[-:-:-]"
}

func (s Submission) GetDomain() string {
	domain := s.Domain
	if domain == "" {
		return ""
	}
	return "[::d]" + " " + paren(domain) + "[-:-:-]"
}

func (s Submission) getComments() string {
	return strconv.Itoa(s.CommentsCount)
}

func (s Submission) getPoints() string {
	return strconv.Itoa(s.Points)
}

func paren(text string) string {
	return "(" + text + ")"
}

func getRankIndentBlock(rank int) string {
	largeIndent := "  "
	smallIndent := " "
	if rank > 9 {
		return smallIndent
	}
	return largeIndent
}
