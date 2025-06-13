//go:build linux

package walk

import (
	"errors"
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

const (
	// size of []byte, given to syscall.Syscall(syscall.SYS_GETDENTS64)
	SCAN_BUF_SIZE = 4 * (1 << 10) // n * (1 << 10) = n KB

	// max retries if getting syscall.EINTR with syscall's
	EINTR_MAX_TRIES = 5
)

// from "man dirent":
//
//	struct linux_dirent64 {
//	    ino64_t         d_ino;       // 8 bytes
//	    off64_t         d_off;       // 8 bytes
//	    unsigned short  d_reclen;    // 2 bytes
//	    unsigned char   d_type;      // 1 bytes
//	    char            d_name[]; // null-terminated
//	};

const (
	NAME_MAX    = syscall.NAME_MAX
	DIRENT_SIZE = 8 + 8 + 2 + 1
)

type dirent struct {
	d_ino    uint64
	d_seek   int64
	d_reclen uint16
	d_type   uint8
	d_namlen uint8
	d_name   [NAME_MAX]byte
	d_path   string
}

// walk through directory and parse all entryes within it
func walk(path string, de_chan chan<- dirent, err_chan chan<- error, wg *sync.WaitGroup, sem chan struct{}) (err error) {
	// NOTE: prevent panic
	defer func() {
		e := recover()
		if e != nil {
			err = errors.New(fmt.Sprint(e))
		}
	}()

	var (
		r1    uintptr
		fd    int
		errno syscall.Errno
	)

	for range EINTR_MAX_TRIES {
		fd, err = syscall.Open(path, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_CLOEXEC, 0)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			return fmt.Errorf("open '%s': %w", path, err)
		}
		break
	}

	defer syscall.Close(fd)

	var (
		buf [SCAN_BUF_SIZE]byte
		n   int
	)

	for {
		r1, _, errno = syscall.Syscall(
			syscall.SYS_GETDENTS64,
			uintptr(fd),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(SCAN_BUF_SIZE),
		)

		if errno != 0 {
			return fmt.Errorf("getdirents64 '%s': %v", path, errno)
		}

		n = int(r1)

		// no entries, quit
		if n == 0 {
			return nil
		}

		for offset := 0; offset < n; {
			var de dirent

			de.d_reclen = *(*uint16)(unsafe.Pointer(&buf[offset+16]))

			if de.d_reclen == 0 {
				continue
			}

			de.d_ino = *(*uint64)(unsafe.Pointer(&buf[offset]))
			if de.d_ino == 0 {
				goto skip
			}

			de.d_name = *(*[NAME_MAX]byte)(unsafe.Pointer(&buf[offset+DIRENT_SIZE]))

			switch de.d_name[0] {
			case 0:
				// empty filename?
				goto skip
			case '.':
				switch de.d_name[1] {
				case 0:
					goto skip
				case '.':
					if de.d_name[2] == 0 {
						goto skip
					}
				}
			}

			de.d_type = *(*uint8)(unsafe.Pointer(&buf[offset+18]))

			for i, b := range de.d_name {
				if b == 0 {
					de.d_namlen = uint8(i)
					break
				}
			}
			if de.d_namlen == 0 {
				de.d_namlen = NAME_MAX
			}

			de.d_path = path + "/" + string(de.d_name[:de.d_namlen])
			de_chan <- de

			if de.d_type == syscall.DT_DIR {
				wg.Add(1)
				go func(path string) {
					sem <- struct{}{}
					defer func() {
						wg.Done()
						<-sem
					}()
					if err := walk(path, de_chan, err_chan, wg, sem); err != nil {
						err_chan <- err
					}
				}(de.d_path)
			}

		skip:
			offset += int(de.d_reclen)
		}
	}
}
