package args

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
)

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

type arggroup int8

const (
	filterGroup arggroup = iota
	optionsGroup
	positionalGroup
)

type Arg[T any] struct {
	Group arggroup
	Name  string
	Desc  string
	parse func()
	Value T
}

type Args struct {
	REName  Arg[*regexp.Regexp]
	REIname Arg[*regexp.Regexp]
	Path    Arg[string]
	Help    Arg[struct{}]
}

//go:embed version
var VERSION string

func Argparse() (*Args, error) {
	var (
		argc  = len(os.Args)
		argv  = os.Args
		index int
		arg   string
		value string
		err   error
	)

	// first, define default args and it's values

	REName := Arg[*regexp.Regexp]{
		Group: filterGroup,
		Name:  "-name",
		Desc:  "filter file names by regex matches",
		Value: nil,
	}

	REName.parse = func() {
		index++
		if index >= argc {
			err = error_missing_arg_value(REName.Name)
		}
		value = argv[index]

		REName.Value, err = regexp.Compile(value)
		fmt.Println("in REName.Parse(): REName.Value: ", REName.Value)
		if err != nil {
			err = error_invalid_arg_value(REName.Name, value, err.Error())
		}
	}

	REIname := Arg[*regexp.Regexp]{
		Group: filterGroup,
		Name:  "-iname",
		Desc:  "exclude files which names matches this regex",
		Value: nil,
	}
	REIname.parse = func() {
		index++
		if index >= argc {
			err = error_missing_arg_value(REIname.Name)
		}
		value = argv[index]

		REIname.Value, err = regexp.Compile(value)
		if err != nil {
			err = error_invalid_arg_value(REIname.Name, value, err.Error())
		}
	}

	Help := Arg[struct{}]{
		Group: optionsGroup,
		Name:  "-help",
		Desc:  "show this help and quit",
	}

	Path := Arg[string]{
		Group: positionalGroup,
		Name:  "PATH",
		Desc:  "start walk from this path",
		Value: ".",
	}
	Path.parse = func() {
		Path.Value = argv[index]
	}

	// make a args struct
	args := &Args{
		REName:  REName,
		REIname: REIname,
		Path:    Path,
		Help:    Help,
	}

	for index = 0; index < argc; index++ {
		arg = argv[index]

		switch arg {
		case REName.Name:
			fmt.Println("parsing rename")
			REName.parse()
			if err != nil {
				return nil, err
			}
			if REName.Value == nil {
				fmt.Println("both rename.Value and err is nil")
			}

		case REIname.Name:
			REIname.parse()
			if err != nil {
				return nil, err
			}

		case Help.Name:
			print_usage(argv[0], args)
			os.Exit(0)

		default:
			if arg[0] == '-' {
				return nil, error_unknown_arg(arg)
			}

			// then this is <path>
			args.Path.Value = arg
		}
	}

	fmt.Println(args.REName.Value)
	fmt.Println(args.REIname.Value)
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

// print usage and exit
func print_usage(prog_name string, args *Args) {
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

	fmt.Println(LOGO)
	fmt.Println()
	fmt.Printf(C_USAGE+"Usage: %s "+args.Path.Name+" [...OPTIONS] [...FILTERS]"+C_RESET+"\n", prog_name)

	posArgs := []Arg[any]{}
	filterArgs := []Arg[any]{}
	optionArgs := []Arg[any]{}
	maxlen := 0

	rvargs := reflect.ValueOf(*args)

	fmt.Println("rvargs.NumField(): ", rvargs.NumField())
	for i := range rvargs.NumField() {
		field := rvargs.Field(i)

		arg := Arg[any]{
			Group: field.FieldByName("Group").Interface().(arggroup),
			Name:  field.FieldByName("Name").Interface().(string),
			Desc:  field.FieldByName("Desc").Interface().(string),
		}

		maxlen = max(maxlen, len(arg.Name))

		switch arg.Group {
		case positionalGroup:
			posArgs = append(posArgs, arg)
		case filterGroup:
			filterArgs = append(filterArgs, arg)
		case optionsGroup:
			optionArgs = append(optionArgs, arg)
		}
	}

	maxlen += 4 // padding

	argtable := [3]struct {
		g  string
		as []Arg[any]
	}{
		{g: "POSITIONAL", as: posArgs},
		{g: "FILTERS", as: filterArgs},
		{g: "OPTIONS", as: optionArgs},
	}

	for _, argg := range argtable { // group, args
		fmt.Println()
		fmt.Println(argg.g)
		for _, arg := range argg.as {
			fmt.Println(DESC_OFFSET + colorize_flag(rwidth(arg.Name, maxlen)) + arg.Desc)
		}
	}
}
