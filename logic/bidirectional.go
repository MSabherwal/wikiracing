package logic

// WIP
// func BidirectionalSearch(start, end string, wiki *wiki.Wikipedia) {
// 	results := []string{}
// 	fwdMap := NewConcurrentMap()
// 	bwdMap := NewConcurrentMap()

// 	fwdMap.Set(start, "")
// 	bwdMap.Set(end, "")

// 	sf := &SearchForward{
// 		ForwardPath: fwdMap,
// 		start:       start,
// 		end:         end,
// 	}

// 	sb := &SearchBackwards{
// 		BackwardPath: bwdMap,
// 		start:        start,
// 		end:          end,
// 	}

// 	ctx, cancel := context.WithCancel(context.Background())

// 	done := make(chan struct{})
// 	searchWG := &sync.WaitGroup{}

// 	// create chans
// 	linkAgg := make(chan string, 1000)
// 	batch := make(chan []string, 10000)

// 	aggStrings := NewConcurrentSet()

// 	numAggregators := 1
// 	searchWG.Add(numAggregators)

// 	for i := 0; i < numAggregators; i++ {
// 		LinkAggregator(ctx, linkAgg, aggStrings, searchWG, batch)
// 	}

// 	si := &SearchForwardInput{
// 		sf:          sf,
// 		wiki:        wiki,
// 		done:        done,
// 		ctx:         ctx,
// 		cancel:      cancel,
// 		wg:          searchWG,
// 		linkAgg:     linkAgg,
// 		batchTitles: batch,
// 	}

// 	var numQueriers = 4
// 	searchWG.Add(numQueriers)
// 	for i := 0; i < numQueriers-1; i++ {
// 		ForwardQuerier(si)
// 	}
// 	firstForwardQuerier(si, from)
// 	go func() {
// 		ticker := time.NewTicker(time.Second * 1)
// 		for range ticker.C {
// 			fmt.Println("len batch: ", len(batch))
// 			fmt.Println("len linkAgg: ", len(linkAgg))

// 		}
// 	}()
// 	<-done

// 	crntSite := to
// 	fp := si.sf.ForwardPath
// 	results = append(results, crntSite)
// 	var next string
// 	for {
// 		next, _ = fp.Get(crntSite)
// 		results = append(results, next)
// 		if next == from {
// 			break
// 		}
// 		crntSite = next
// 	}
// 	resultsAsc := []string{}
// 	for i := len(results); i > 0; i-- {
// 		resultsAsc = append(resultsAsc, results[i-1])
// 	}
// 	return resultsAsc
// }
