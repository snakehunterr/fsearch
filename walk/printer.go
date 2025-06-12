package walk

import (
	"fmt"
	"os"
)

func printer(r_chan <-chan dirent) chan struct{} {
	done_ch := make(chan struct{})

	go func() {
		for {
			de, ok := <-r_chan
			if !ok {
				// channel is closed: close done_ch and quit
				close(done_ch)
				return
			}

			fmt.Println(de.d_path)
		}
	}()

	return done_ch
}

func errprinter(err_chan <-chan error) chan struct{} {
	done_ch := make(chan struct{})

	go func() {
		for {
			err, ok := <-err_chan
			if !ok {
				// channel is closed: close done_ch and quit
				close(done_ch)
				return
			}

			fmt.Fprintln(os.Stderr, err)
		}
	}()

	return done_ch
}
