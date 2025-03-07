package model

type Webhook struct {
	Avatar       string
	RepoName     string
	PrUrl        string
	Action       string
	HRef         string
	BRef         string
	ContribUname string
	ContribUrl   string
	IsMerged     bool
	MergedBy     string
}
