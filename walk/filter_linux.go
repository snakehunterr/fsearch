//go:build linux

package walk

import (
	"runtime"
	"sync"

	"github.com/snakehunterr/fsearch/args"
)

func filter(de_chan <-chan dirent, r_chan chan<- dirent, args *args.Args) chan struct{} {
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
				if args.RE_name != nil {
					if !args.RE_name.Match(de.d_name[:de.d_namlen]) {
						continue
					}
				}

				if args.RE_iname != nil {
					if args.RE_iname.Match(de.d_name[:de.d_namlen]) {
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
