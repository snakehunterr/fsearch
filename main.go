// package main

// import (
// 	"fmt"
// 	"os"
//
// 	"github.com/snakehunterr/fsearch/args"
// 	"github.com/snakehunterr/fsearch/walk"
// )
//
// func main() {
// 	args, err := args.Argparse()
// 	if err != nil {
// 		fmt.Fprintln(os.Stderr, err)
// 		os.Exit(1)
// 	}
//
// 	if err := walk.Walk(args); err != nil {
// 		fmt.Fprintln(os.Stderr, err)
// 		os.Exit(1)
// 	}
// }

package main

import (
	"fmt"
	"runtime/debug"
)

func main() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Println("no version")
		return
	}

	fmt.Println(info)
}
