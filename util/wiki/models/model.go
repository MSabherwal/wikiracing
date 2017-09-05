package models

//all of these derived from the http response using: https://mholt.github.io/json-to-go/

//QueryResponse is just that
type QueryResponse struct {
	Continue struct {
		Lhcontinue string `json:"lhcontinue"`
		Plcontinue string `json:"plcontinue"`
		Continue   string `json:"continue"`
	} `json:"continue"`
	Query struct {
		Pages map[int]Page
	} `json:"query"`
	Limits struct {
		Links     int `json:"links"`
		Linkshere int `json:"linkshere"`
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
	Linkshere []struct {
		Pageid   int    `json:"pageid"`
		Ns       int    `json:"ns"`
		Title    string `json:"title"`
		Redirect string `json:"redirect,omitempty"`
	} `json:"linkshere"`
}

func (qr *QueryResponse) ShouldContinue(prefix string) bool {
	switch prefix {
	case "pl":
		return len(qr.Continue.Plcontinue) > 0
	case "lh":
		return len(qr.Continue.Lhcontinue) > 0
	}
	return false
}

func (qr *QueryResponse) ContinueVal(prefix string) string {
	switch prefix {
	case "pl":
		return qr.Continue.Plcontinue
	case "lh":
		return qr.Continue.Lhcontinue
	}
	return ""
}
