package args

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Args struct {
	RE_name  *regexp.Regexp
	RE_iname *regexp.Regexp
	Path     string
}

const (
	C_RED     = "\x1B[31m"
	C_GREEN   = "\x1B[32m"
	C_YELLOW  = "\x1B[33m"
	C_BLUE    = "\x1B[34m"
	C_MAGENTA = "\x1B[35m"
	C_CYAN    = "\x1B[36m"
	C_RESET   = "\x1B[0m"
	C_LINE1   = "\x1B[38;2;203;166;247m"
	C_LINE2   = "\x1B[38;2;213;159;226m"
	C_LINE3   = "\x1B[38;2;223;152;207m"
	C_LINE4   = "\x1B[38;2;232;146;189m"
	C_LINE5   = "\x1B[38;2;243;139;167m"
	C_USAGE   = "\x1B[38;2;190;190;190m"
)

const (
	ARG_NAME      = "-name"
	ARG_INAME     = "-iname"
	ARG_HELP      = "-h"
	ARG_HELP_LONG = "--help"
)

func Argparse() (*Args, error) {
	var (
		argc  = len(os.Args)
		argv  = os.Args
		arg   string
		value string
		err   error
	)

	args := &Args{
		Path: ".", // set default value
	}

	for i := 1; i < argc; i++ {
		arg = argv[i]

		switch arg {
		case ARG_NAME:
			i++
			if i == argc {
				return nil, error_missing_arg_value(arg)
			}
			value = argv[i]

			args.RE_name, err = regexp.Compile(value)
			if err != nil {
				return nil, error_invalid_arg_value(arg, value, err.Error())
			}

		case ARG_INAME:
			i++
			if i == argc {
				return nil, error_missing_arg_value(arg)
			}
			value = argv[i]

			args.RE_iname, err = regexp.Compile(value)
			if err != nil {
				return nil, error_invalid_arg_value(arg, value, err.Error())
			}

		case ARG_HELP, ARG_HELP_LONG:
			print_usage(argv[0])
			os.Exit(0)

		default:
			if arg[0] == '-' {
				return nil, error_unknown_arg(arg)
			}

			// then this is <path>
			args.Path = arg
		}
	}

	return args, nil
}

func colorize_flag(arg string) string {
	return C_GREEN + arg + C_RESET
}

func error_wrapper(s string) error {
	return errors.New(C_RED + "error" + C_RESET + ": " + s)
}

func error_missing_arg_value(arg string) error {
	return error_wrapper("value missing for argument: " + colorize_flag(arg))
}

func error_invalid_arg_value(arg, value, desc string) error {
	return error_wrapper(fmt.Sprintf("incorrect value '%s' for argument %s: %s", value, colorize_flag(arg), desc))
}

func error_unknown_arg(arg string) error {
	return error_wrapper("unknown flag: " + colorize_flag(arg))
}

var VERSION string

// print usage and exit
func print_usage(prog_name string) {
	rwidth := func(s string, n int) string {
		if n <= len(s) {
			return s
		}

		return s + strings.Repeat(" ", n-len(s))
	}

	LOGO := C_LINE1 + "    dMMMMMP .dMMMb " + C_RESET +
		"  dMMMMMP .aMMMb  dMMMMb  .aMMMb  dMP dMP\n" + C_LINE2 +
		"   dMP     dMP\" VP" + C_RESET +
		" dMP     dMP\"dMP dMP.dMP dMP\"VMP dMP dMP\n" + C_LINE3 +
		"  dMMMP    VMMMb" + C_RESET +
		"  dMMMP   dMMMMMP dMMMMK\" dMP     dMMMMMP\n" + C_LINE4 +
		" dMP     dP .dMP" + C_RESET +
		" dMP     dMP dMP dMP\"AMF dMP.aMP dMP dMP\n" + C_LINE5 +
		"dMP      VMMMP\"" + C_RESET +
		" dMMMMMP dMP dMP dMP dMP  VMMMP\" dMP dMP    " + C_LINE5 +
		VERSION + C_RESET

	const DESC_OFFSET = "  "

	argtable := map[string]map[string]string{
		"POSITIONAL": {
			"PATH": "init path for walk",
		},
		"FILTERS": {
			ARG_NAME:  "filter files by regex matches",
			ARG_INAME: "exclude files by regex matches",
		},
		"OPTIONS": {
			ARG_HELP + " " + ARG_HELP_LONG: "show this help and quit",
		},
	}

	fmt.Println(LOGO)
	fmt.Println()
	fmt.Printf(C_USAGE+"Usage: %s [...OPTIONS] [...FILTERS] [PATH]"+C_RESET+"\n", prog_name)

	var max_len int
	for _, args := range argtable {
		for k := range args {
			if len(k) > max_len {
				max_len = len(k)
			}
		}
	}
	max_len += 4

	for g, args := range argtable { // group, args
		fmt.Println()
		fmt.Println(g)
		for n, h := range args { // name, help
			fmt.Println(DESC_OFFSET + colorize_flag(rwidth(n, max_len)) + h)
		}
	}
}
