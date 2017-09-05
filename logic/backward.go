package logic

import (
	"context"
	"fmt"
	"interview_questions/segment/wikiracing/util/wiki"
	"sync"
	"time"
)

type SearchBackwards struct {
	//ForwardPath keeps track of "parent" of a link
	BackwardPath *ConcurrentMap

	start string
	end   string
}

type SearchBackwardsInput struct {
	batchTitles <-chan []string
	linkAgg     chan<- string
	done        chan *DonePath
	sb          *SearchBackwards
	sf          *SearchForward
	wiki        *wiki.Wikipedia
	cancel      context.CancelFunc
	ctx         context.Context
	wg          *sync.WaitGroup
}

func SearchBackwardsFrom(from, to string, wiki *wiki.Wikipedia) []string {
	results := []string{}
	bwdMap := NewConcurrentMap()
	bwdMap.Set(to, "")
	sb := &SearchBackwards{
		BackwardPath: bwdMap,
		start:        from,
		end:          to,
	}
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan *DonePath)
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

	si := &SearchBackwardsInput{
		sb:          sb,
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
		BackwardsQuerier(si)
	}
	firstBackwardsQuerier(si, to)
	go func() {
		ticker := time.NewTicker(time.Second * 1)
		for range ticker.C {
			fmt.Println("len batch: ", len(batch))
			fmt.Println("len linkAgg: ", len(linkAgg))

		}
	}()
	dn := <-done

	//parse results
	path := []string{}
	var crntWord string
	switch dn.Path {
	case Midpoint:
	case Full:
		path = append(path, dn.Word)
		crntWord = dn.Word
		for crntWord != to {
			nextWord, _ := si.sb.BackwardPath.Get(crntWord)
			path = append(path, nextWord)
			crntWord = nextWord
		}
	}

	results = path

	resultsAsc := []string{}
	for i := len(results); i > 0; i-- {
		resultsAsc = append(resultsAsc, results[i-1])
	}
	return resultsAsc
}

func BackwardsQuerier(si *SearchBackwardsInput) {

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

				qi := wiki.NewLinksHereQuery(batch)

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
						to := page.Title
						for _, from := range page.Linkshere {
							// if path already exists to this node, ignore
							if si.sb.BackwardPath.Exists(from.Title) {
								continue
							}

							si.sb.BackwardPath.Set(from.Title, to)
							// found end page!

							if from.Title == si.sb.start || (si.sf != nil && si.sf.ForwardPath.Exists(from.Title)) {
								fmt.Println("found page!")
								si.cancel()
								var path PathType
								if from.Title == si.sb.start {
									path = Full
								} else {
									path = Midpoint
								}
								si.done <- &DonePath{
									Path: path,
									Word: from.Title,
								}
								return
							}
							select {
							case <-si.ctx.Done():
								return
							default:
							}
							//send title
							si.linkAgg <- from.Title
						}
					}
					//determine if you need to continue
					if !resp.ShouldContinue(qi.Prefix()) {
						break
					}
					qi.Cont = resp.ContinueVal(qi.Prefix())
				}

				//reset timer
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(time.Second * 2)
			}
		}

	}()

	return
}

func firstBackwardsQuerier(si *SearchBackwardsInput, end string) {

	go func() {
		defer si.wg.Done()

		qi := wiki.NewLinksHereQuery([]string{end})

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
				to := page.Title
				for _, from := range page.Linkshere {
					// ignore if exists already
					if si.sb.BackwardPath.Exists(from.Title) {
						continue
					}

					si.sb.BackwardPath.Set(from.Title, to)

					if from.Title == si.sb.start /*|| si.sf.ForwardPath.Exists(from.Title)*/ {
						fmt.Println("found page!")
						si.cancel()
						var path PathType
						if from.Title == si.sb.start {
							path = Full
						} else {
							path = Midpoint
						}
						si.done <- &DonePath{
							Path: path,
							Word: from.Title,
						}
						return
					}
					select {
					case <-si.ctx.Done():
						return
					default:
					}
					//send title
					si.linkAgg <- from.Title
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
