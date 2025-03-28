package model

type Webhook struct {
	Avatar         string
	RepoName       string
	PrUrl          string
	SrcPrUrl       string
	Action         string
	HRef           string
	BRef           string
	ContribUname   string
	ContribUrl     string
	IsMerged       bool
	MergedBy       string
	ApproveComment string
	ApproveUrl     string
	Comment        string
	CommentUrl     string
	DiffHunk       string
}
