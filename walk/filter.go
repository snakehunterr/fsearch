package walk

import (
	"regexp"
	"sync"
)

type filters struct {
	REname  *regexp.Regexp
	REiname *regexp.Regexp
}

func filter(de_chan <-chan dirent, r_chan chan<- dirent, filters *filters) chan struct{} {
	var (
		done    = make(chan struct{})
		wg      = &sync.WaitGroup{}
		workers = 8
	)

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				de, ok := <-de_chan
				if !ok {
					return
				}

				// process this entry...
				if filters.REname != nil {
					if !filters.REname.Match(de.d_name[:de.d_namlen]) {
						continue
					}
				}

				if filters.REname != nil {
					if filters.REiname.Match(de.d_name[:de.d_namlen]) {
						continue
					}
				}

				r_chan <- de
			}
		}()
	}

	go func() {
		wg.Wait()
		close(r_chan)
		close(done)
	}()

	return done
}
