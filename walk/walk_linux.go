//go:build darwin

package walk

import (
	"errors"
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

const (
	// size of []byte, given to syscall.Syscall6(GETDIRENT)
	SCAN_BUF_SIZE = 4 * (1 << 10) // n * (1 << 10) = n KB

	// max retries if getting syscall.EINTR with syscall's
	EINTR_MAX_TRIES = 5
)

// from "man dirent":
//
//	struct dirent {
//	    ino_t        d_ino;       // 8 bytes
//	    uint64_t     d_seekoff;   // 8 bytes
//	    uint16_t     d_reclen;    // 2 bytes
//	    uint16_t     d_namlen;    // 2 bytes
//	    uint8_t      d_type;      // 1 byte
//	    char         d_name\[1024\];
//	};

// WARN: macOS currently can't create files which have
// name greater than 255 (syscall.NAME_MAX constant)
// should then NAME_MAX const setted to syscall.NAME_MAX?
const (
	NAME_MAX    = 1024 // according to `man dirent`
	DIRENT_SIZE = 8 + 8 + 2 + 2 + 1
)

type dirent struct {
	d_ino    uint64
	d_seek   uint64
	d_reclen uint16
	d_namlen uint16
	d_type   uint8
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
		buf   [SCAN_BUF_SIZE]byte
		basep int64 // macOS off_t
		n     int
	)

	for {
		r1, _, errno = syscall.Syscall6(
			syscall.SYS_GETDIRENTRIES64,
			uintptr(fd),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(len(buf)),
			uintptr(unsafe.Pointer(&basep)),
			0, 0, // padding
		)

		if errno != 0 {
			return fmt.Errorf("getdirents64 '%s': %v", path, errno)
		}

		n = int(r1)

		// no entries, quit
		if n == 0 {
			return nil
		}

		offset := 0
		for offset < n {
			if offset+DIRENT_SIZE > n {
				break
			}

			d_reclen := *(*uint16)(unsafe.Pointer(&buf[offset+16]))
			if d_reclen == 0 || offset+int(d_reclen) > n {
				break
			}

			var (
				d_ino     uint64
				d_seekoff uint64
				d_namlen  uint16
				d_type    uint8
				d_name    [NAME_MAX]byte
			)

			d_ino = *(*uint64)(unsafe.Pointer(&buf[offset])) // 0

			// XXX: from os/dir_darwin.go `func(f *File) readdir():`
			// Darwin may return a zero inode when a directory entry has been
			// deleted but not yet removed from the directory. The man page for
			// getdirentries(2) states that programs are responsible for skipping
			// those entries:
			//
			//   Users of getdirentries() should skip entries with d_fileno = 0,
			//   as such entries represent files which have been deleted but not
			//   yet removed from the directory entry.
			if d_ino == 0 {
				goto skip
			}

			d_seekoff = *(*uint64)(unsafe.Pointer(&buf[offset+8])) // 0 + 8 = 8
			// d_recl already parsed // 8 + 8 = 16
			d_namlen = *(*uint16)(unsafe.Pointer(&buf[offset+18])) // 16 + 2 = 18
			d_type = *(*uint8)(unsafe.Pointer(&buf[offset+20]))    // 18 + 2 = 20

			// Name starts at offset DIRENT_SIZE
			if d_namlen > 0 && offset+DIRENT_SIZE+int(d_namlen) <= n {
				// zero-copy name conversion
				_tempbuf := buf[offset+DIRENT_SIZE : offset+DIRENT_SIZE+int(d_namlen)]
				d_name = *(*[NAME_MAX]byte)(unsafe.Pointer(&_tempbuf[0]))

				if (d_namlen == 1 && d_name[0] == '.') || (d_namlen == 2 && d_name[0] == '.' && d_name[1] == '.') {
					goto skip
				}

				d_path := path + "/" + string(d_name[:d_namlen])

				de_chan <- dirent{
					d_ino:    d_ino,
					d_seek:   d_seekoff,
					d_reclen: d_reclen,
					d_namlen: d_namlen,
					d_type:   d_type,
					d_name:   d_name,
					d_path:   d_path,
				}

				if d_type == syscall.DT_DIR {
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
					}(d_path)
				}
			}

		skip:
			offset += int(d_reclen)
		}
	}
}
