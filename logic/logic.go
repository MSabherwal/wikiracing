package logic

import (
	"context"
	"fmt"
	"interview_questions/segment/wikiracing/util/wiki"
	"sync"
)

type SearchForward struct {
	//ForwardPath keeps track of "parent" of a link
	ForwardPath *ConcurrentMap

	start string
	end   string
}

func LinkAggregator(ctx context.Context, links <-chan string, aggregatedStrings *ConcurrentSet, wg *sync.WaitGroup, aggregate chan []string) {

	go func() {
		defer wg.Done()
		var toBeSearched []string
		for {
			select {
			case <-ctx.Done():
				return
			case link := <-links:
				//if already aggregated, ignore
				// maybe redundant?
				if aggregatedStrings.Exists(link) {
					continue
				}
				toBeSearched = append(toBeSearched, link)
				if len(toBeSearched) == 50 {
					aggregate <- toBeSearched
					toBeSearched = []string{}
				}

			}
		}
	}()
	return
}

type SearchInput struct {
	batchTitles <-chan []string
	sf          *SearchForward
	wiki        *wiki.Wikipedia
	linkAgg     chan<- string
	done        chan struct{}
	cancel      context.CancelFunc
	ctx         context.Context
	wg          *sync.WaitGroup
}

func Search(from, to string, wiki *wiki.Wikipedia) []string {
	results := []string{}
	fwdMap := NewConcurrentMap()
	fwdMap.Set(from, "")
	sf := &SearchForward{
		ForwardPath: fwdMap,
		start:       from,
		end:         to,
	}
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	searchWG := &sync.WaitGroup{}

	// create chans
	linkAgg := make(chan string, 1000)
	batch := make(chan []string, 1000)

	aggStrings := NewConcurrentSet()

	numAggregators := 2
	searchWG.Add(numAggregators)

	for i := 0; i < numAggregators; i++ {
		LinkAggregator(ctx, linkAgg, aggStrings, searchWG, batch)
	}

	si := &SearchInput{
		sf:          sf,
		wiki:        wiki,
		done:        done,
		ctx:         ctx,
		cancel:      cancel,
		wg:          searchWG,
		linkAgg:     linkAgg,
		batchTitles: batch,
	}

	var numQueriers = 5
	searchWG.Add(numQueriers)
	for i := 0; i < numQueriers-1; i++ {
		querier(si)
	}
	firstQuerier(si, from)

	<-done

	crntSite := to
	fp := si.sf.ForwardPath
	var next string
	for {
		next, _ = fp.Get(crntSite)
		results = append(results, next)
		if next == from {
			break
		}
		crntSite = next
	}

	return results
}

// to start populating the loop
func firstQuerier(si *SearchInput, start string) {

	go func() {
		defer si.wg.Done()
		for {

			qi := &wiki.QueryInput{
				Prop:   "links",
				Titles: []string{start},
				Prefix: "pl",
				Cont:   "",
			}

			for {
				resp, err := si.wiki.Query(qi)
				if err != nil {
					panic(err)
				}

				for _, page := range resp.Query.Pages {
					select {
					case <-si.ctx.Done():
						return
					default:
					}
					from := page.Title
					for _, to := range page.Links {
						// if path already exists to this node, ignore
						if si.sf.ForwardPath.Exists(to.Title) {
							continue
						}

						si.sf.ForwardPath.Set(to.Title, from)
						// found end page!
						if to.Title == si.sf.end {
							si.cancel()
							si.done <- struct{}{}
							return
						}
						select {
						case <-si.ctx.Done():
							return
						default:

						}

						//send title
						si.linkAgg <- to.Title

					}
				}
				//determine if you need to continue
				if !resp.ShouldContinue(qi.Prefix) {
					break
				}

				qi.Cont = resp.ContinueVal(qi.Prefix)
			}

		}
	}()

	return
}

func querier(si *SearchInput) {

	go func() {
		defer si.wg.Done()
		for {
			for batch := range si.batchTitles {
				qi := &wiki.QueryInput{
					Prop:   "links",
					Titles: batch,
					Prefix: "pl",
					Cont:   "",
				}
				fmt.Println("titles:", batch)

				for {
					fmt.Println("qi cont:", qi.Cont)
					resp, err := si.wiki.Query(qi)
					if err != nil {
						panic(err)
					}

					for _, page := range resp.Query.Pages {
						select {
						case <-si.ctx.Done():
							return
						default:
						}
						from := page.Title
						for _, to := range page.Links {
							// if path already exists to this node, ignore
							if si.sf.ForwardPath.Exists(to.Title) {
								continue
							}

							si.sf.ForwardPath.Set(to.Title, from)
							// found end page!

							if to.Title == si.sf.end {
								fmt.Println("found page!")
								si.cancel()
								si.done <- struct{}{}
								return
							}
							select {
							case <-si.ctx.Done():
								return
							default:
							}
							//send title
							si.linkAgg <- to.Title
						}
					}
					//determine if you need to continue
					if !resp.ShouldContinue(qi.Prefix) {
						break
					}
					qi.Cont = resp.ContinueVal(qi.Prefix)
				}
			}
		}
	}()

	return
}

// func Race(startPage, endPage string) []string {
// 	//
// }

func NewConcurrentMap() *ConcurrentMap {
	return &ConcurrentMap{
		data:    make(map[string]string),
		RWMutex: &sync.RWMutex{},
	}
}

type ConcurrentMap struct {
	data map[string]string
	*sync.RWMutex
}

func (cm *ConcurrentMap) Get(in string) (string, bool) {
	cm.RLock()
	defer cm.RUnlock()
	val, exists := cm.data[in]
	return val, exists
}

func (cm *ConcurrentMap) Exists(in string) bool {
	cm.RLock()
	defer cm.RUnlock()
	_, exists := cm.data[in]
	return exists
}

func (cm *ConcurrentMap) Set(in, val string) {
	cm.Lock()
	defer cm.Unlock()
	cm.data[in] = val
	return
}

type ConcurrentSet struct {
	data map[string]struct{}
	*sync.RWMutex
}

func NewConcurrentSet() *ConcurrentSet {
	return &ConcurrentSet{
		data:    make(map[string]struct{}),
		RWMutex: &sync.RWMutex{},
	}
}

func (cs *ConcurrentSet) Exists(in string) bool {
	cs.RLock()
	defer cs.RUnlock()
	_, exists := cs.data[in]
	return exists
}

func (cs *ConcurrentSet) Add(in string) {
	cs.Lock()
	defer cs.Unlock()
	cs.data[in] = struct{}{}
}
