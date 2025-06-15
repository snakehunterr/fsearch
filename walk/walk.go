package walk

import (
	"runtime"
	"sync"

	args "github.com/snakehunterr/fsearch/args"
)

const (
	// channel size of parsed dirent{} structs from syscall
	SCAN_CHAN_SIZE = 1 << 10
)

// walk through directory with parallel power
func Walk(args *args.Args) (err error) {
	var (
		// sync.WaitGroup for walk workers group
		wg = &sync.WaitGroup{}
		// channel of parsed, but unfiltered dirents
		de_chan  = make(chan dirent, SCAN_CHAN_SIZE)
		r_chan   = make(chan dirent, SCAN_CHAN_SIZE)
		err_chan = make(chan error, 10)
		// MAX_WORKERS
		//
		// WARN: no need to close because runtime will auto close it on exit
		sem = make(chan struct{}, runtime.NumCPU()*8)
	)

	f_done := filter(de_chan, r_chan, &filters{
		REname: args.RE_name,
	})
	p_done := printer(r_chan)
	ep_done := errprinter(err_chan)

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = walk(args.Path, de_chan, err_chan, wg, sem)
	}()

	if err != nil {
		return err
	}

	go func() {
		wg.Wait()
		close(de_chan)
		close(err_chan)
	}()

	<-f_done
	<-p_done
	<-ep_done

	return nil
}
