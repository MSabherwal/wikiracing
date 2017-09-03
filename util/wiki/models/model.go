package models

//QueryResponse is just that
type QueryResponse struct {
	Continue struct {
		Plcontinue string `json:"plcontinue"`
		Continue   string `json:"continue"`
	} `json:"continue"`
	Query struct {
		Pages map[int]Page
	} `json:"query"`
	Limits struct {
		Links int `json:"links"`
	} `json:"limits"`
}

//Page represents all links from a page(title)
type Page struct {
	Pageid int    `json:"pageid"`
	Ns     int    `json:"ns"`
	Title  string `json:"title"`
	Links  []struct {
		Ns    int    `json:"ns"`
		Title string `json:"title"`
	} `json:"links"`
}

func (qr *QueryResponse) ShouldContinue(prefix string) bool {
	switch prefix {
	case "pl":
		return len(qr.Continue.Plcontinue) > 0
	}
	return false
}

func (qr *QueryResponse) ContinueVal(prefix string) string {
	switch prefix {
	case "pl":
		return qr.Continue.Plcontinue
	}
	return ""
}
