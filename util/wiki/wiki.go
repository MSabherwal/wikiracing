//Package wiki contains all logic for parsing+calling wikipedia
package wiki

import (
	"encoding/json"
	"fmt"
	"interview_questions/segment/wikiracing/util/wiki/models"
	"net/http"
	"net/url"
	"strings"
)

func newBaseURI() *url.URL {
	return &url.URL{Scheme: "https", Opaque: "", Host: "en.wikipedia.org", Path: "/w/api.php", RawPath: "", ForceQuery: false, RawQuery: "", Fragment: ""}
}

//Wikipedia wraps all basic wiki calls
type Wikipedia struct {
	client *http.Client
}

func New() *Wikipedia {
	return &Wikipedia{
		client: &http.Client{},
	}
}

type QueryInput struct {
	prefix string
	prop   string
	Titles []string
	Cont   string
}

func (qi *QueryInput) Prefix() string {
	return qi.prefix
}

func (qi *QueryInput) Prop() string {
	return qi.prop
}

func NewLinksQuery(titles []string) *QueryInput {
	return &QueryInput{
		prefix: "pl",
		prop:   "links",
		Titles: titles,
		Cont:   "",
	}
}

func NewLinksHereQuery(titles []string) *QueryInput {
	return &QueryInput{
		prefix: "lh",
		prop:   "linkshere",
		Titles: titles,
		Cont:   "",
	}
}

// Query ...
func (wp *Wikipedia) Query(in *QueryInput) (*models.QueryResponse, error) {
	uri := in.queryURI()

	resp, err := wp.client.Get(uri.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	qResp := &models.QueryResponse{}
	json.NewDecoder(resp.Body).Decode(qResp)
	return qResp, err
}

func (qi *QueryInput) queryURI() *url.URL {
	uri := newBaseURI()

	values := url.Values{}

	values.Add("prop", qi.prop)
	values.Add("format", "json")
	values.Add("action", "query")
	values.Add("titles", strings.Join(qi.Titles, "|"))

	values.Add(fmt.Sprintf("%slimit", qi.prefix), "max")
	if len(qi.Cont) > 0 {
		values.Add(fmt.Sprintf("%scontinue", qi.prefix), qi.Cont)
	}

	uri.RawQuery = values.Encode()
	return uri
}
