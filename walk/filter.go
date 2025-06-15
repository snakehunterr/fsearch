//go:build linux

package walk

import (
	"regexp"
	"runtime"
	"sync"
)

type filters struct {
	name  *regexp.Regexp
	iname *regexp.Regexp
}

func filter(de_chan <-chan dirent, r_chan chan<- dirent, filters *filters) chan struct{} {
	var (
		done    = make(chan struct{})
		wg      = &sync.WaitGroup{}
		workers = min(runtime.NumCPU()/2, 1)
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
				if filters.name != nil {
					if !filters.name.Match(de.d_name[:de.d_namlen]) {
						continue
					}
				}

				if filters.iname != nil {
					if filters.iname.Match(de.d_name[:de.d_namlen]) {
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
