package message

import "clx/item"

type EditorFinishedMsg struct {
	Err error
}

type EnteringCommentSection struct {
	Id           int
	CommentCount int
}

type EnteringReaderMode struct {
	Url    string
	Title  string
	Domain string
}

type StatusMessageTimeout struct{}

type FetchingFinished struct {
	Message string
}

type ChangeCategory struct {
	Category int
	Cursor   int
}

type CategoryFetchingFinished struct {
	Category int
	Cursor   int
	Message  string
}

type AddToFavorites struct {
	Item *item.Item
}
