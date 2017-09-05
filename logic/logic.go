package logic

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type DonePath struct {
	Path string
	Word string
}

func LinkAggregator(ctx context.Context, links <-chan string, aggregatedStrings *ConcurrentSet, wg *sync.WaitGroup, aggregate chan []string) {

	go func() {
		defer wg.Done()
		var toBeSearched []string
		timer := time.NewTimer(time.Second * 1)
		flushCnt := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				fmt.Println("linkAggregator hasn't rcvd data in 1 second! flushing.")
				if flushCnt >= 2 || len(toBeSearched) == 0 {
					fmt.Println("second attempt to flush or no links have been rcvd since last flush! rtrning")
					return
				}
				aggregate <- toBeSearched
				toBeSearched = []string{}
				flushCnt++

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

				//reset timer
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(time.Second * 1)
				flushCnt = 0

			}
		}
	}()
	return
}
