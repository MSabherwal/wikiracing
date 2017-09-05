package logic

import (
	"context"
	"fmt"
	"interview_questions/segment/wikiracing/util/wiki"
	"sync"
	"time"
)

type SearchForward struct {
	//ForwardPath keeps track of "parent" of a link
	ForwardPath *ConcurrentMap

	start string
	end   string
}

type SearchForwardInput struct {
	batchTitles <-chan []string
	sf          *SearchForward
	sb          *SearchBackwards
	wiki        *wiki.Wikipedia
	linkAgg     chan<- string
	done        chan struct{}
	cancel      context.CancelFunc
	ctx         context.Context
	wg          *sync.WaitGroup
}

func SearchForwardTo(from, to string, wiki *wiki.Wikipedia) []string {
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
	batch := make(chan []string, 10000)

	aggStrings := NewConcurrentSet()

	numAggregators := 1
	searchWG.Add(numAggregators)

	for i := 0; i < numAggregators; i++ {
		LinkAggregator(ctx, linkAgg, aggStrings, searchWG, batch)
	}

	si := &SearchForwardInput{
		sf:          sf,
		wiki:        wiki,
		done:        done,
		ctx:         ctx,
		cancel:      cancel,
		wg:          searchWG,
		linkAgg:     linkAgg,
		batchTitles: batch,
	}

	var numQueriers = 4
	searchWG.Add(numQueriers)
	for i := 0; i < numQueriers-1; i++ {
		ForwardQuerier(si)
	}
	firstForwardQuerier(si, from)
	go func() {
		ticker := time.NewTicker(time.Second * 1)
		for range ticker.C {
			fmt.Println("len batch: ", len(batch))
			fmt.Println("len linkAgg: ", len(linkAgg))

		}
	}()
	<-done

	crntSite := to
	fp := si.sf.ForwardPath
	results = append(results, crntSite)
	var next string
	for {
		next, _ = fp.Get(crntSite)
		results = append(results, next)
		if next == from {
			break
		}
		crntSite = next
	}
	resultsAsc := []string{}
	for i := len(results); i > 0; i-- {
		resultsAsc = append(resultsAsc, results[i-1])
	}
	return resultsAsc
}

// to start populating the loop
func firstForwardQuerier(si *SearchForwardInput, start string) {

	go func() {
		defer si.wg.Done()

		qi := wiki.NewLinksQuery([]string{start})

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

			if !resp.ShouldContinue(qi.Prefix()) {
				break
			}

			qi.Cont = resp.ContinueVal(qi.Prefix())
		}

	}()

	return
}

func ForwardQuerier(si *SearchForwardInput) {

	go func() {
		timer := time.NewTimer(time.Second * 2)
		defer si.wg.Done()
		for {
			select {
			case <-si.ctx.Done():
				return
			case <-timer.C:
				fmt.Println("querier returning from no data for 2 sec")
				return
			case batch := <-si.batchTitles:

				qi := wiki.NewLinksQuery(batch)

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
					if !resp.ShouldContinue(qi.Prefix()) {
						break
					}
					qi.Cont = resp.ContinueVal(qi.Prefix())
				}

				//flush timer for reset
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(time.Second * 2)
			}
		}

	}()

	return
}
