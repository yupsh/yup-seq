// Command yup-seq is the CLI wrapper around github.com/gloo-foo/cmd-seq.
package main

import (
	"strconv"

	clix "github.com/gloo-foo/cli"
	command "github.com/gloo-foo/cmd-seq"
	urf "github.com/urfave/cli/v3"
)

// version is the build version. It defaults to "dev" for local builds and is
// overridden at release time via the linker: -ldflags "-X main.version=<v>".
var version = "dev"

const (
	name           = "seq"
	flagSeparator  = "separator"
	flagFormat     = "format"
	flagEqualWidth = "equal-width"
)

// Error is the sole error type the wrapper emits, so every failure path is
// testable with errors.Is.
type Error string

func (e Error) Error() string { return string(e) }

const (
	// ErrOperandCount is returned when the number of numeric operands is not 1, 2, or 3.
	ErrOperandCount Error = "usage: seq [OPTIONS] LAST | FIRST LAST | FIRST INCREMENT LAST"
	// ErrInvalidOperand is returned when an operand is not a valid number.
	ErrInvalidOperand Error = "invalid operand"
)

// synopsis is the multi-line --help usage block; urfave/cli indents it three
// spaces, so the lines stay flush-left.
const synopsis = `seq [OPTIONS] LAST
seq [OPTIONS] FIRST LAST
seq [OPTIONS] FIRST INCREMENT LAST

print numbers from FIRST to LAST, in steps of INCREMENT.`

// spec declares the seq wrapper. seq is a source command: it produces its
// sequence directly, so build returns it as the whole pipeline (a nil filter).
var spec = clix.Spec{
	Name:     name,
	Summary:  "print a sequence of numbers",
	Synopsis: synopsis,
	Build:    build,
	Flags:    flags(),
}

// flags builds a fresh set of the wrapper's flags. It is a constructor rather
// than a package var so each parse gets independent flag values (urfave/cli
// records IsSet state on the flag itself, which would otherwise leak between
// invocations that share the pointers).
func flags() []urf.Flag {
	return []urf.Flag{
		&urf.StringFlag{
			Name:    flagSeparator,
			Aliases: []string{"s"},
			Usage:   "use STRING to separate numbers (default: \\n)",
		},
		&urf.StringFlag{Name: flagFormat, Aliases: []string{"f"}, Usage: "use printf style floating-point FORMAT"},
		&urf.BoolFlag{
			Name:    flagEqualWidth,
			Aliases: []string{"w"},
			Usage:   "equalize width by padding with leading zeroes",
		},
	}
}

// build maps the invocation to seq's pipeline: the numeric operands and flags
// produce the sequence source, with no filter. A wrong operand count or a
// non-numeric operand is a usage error.
func build(inv clix.Invocation) (clix.Source, clix.Command, error) {
	nums, err := operands(inv.Args)
	if err != nil {
		return nil, nil, err
	}
	return command.Seq(append(nums, options(inv.Args)...)...), nil, nil
}

// operands parses the 1-3 numeric operands into constructor arguments.
func operands(c *urf.Command) ([]any, error) {
	if c.NArg() < 1 || c.NArg() > 3 {
		return nil, ErrOperandCount
	}
	nums := make([]any, c.NArg())
	for i := range nums {
		val, err := strconv.ParseFloat(c.Args().Get(i), 64)
		if err != nil {
			return nil, ErrInvalidOperand
		}
		nums[i] = val
	}
	return nums, nil
}

// options folds the parsed flags into seq's option values.
func options(c *urf.Command) []any {
	var opts []any
	if c.IsSet(flagSeparator) {
		opts = append(opts, command.SeqSeparator(c.String(flagSeparator)))
	}
	if c.IsSet(flagFormat) {
		opts = append(opts, command.SeqFormat(c.String(flagFormat)))
	}
	if c.Bool(flagEqualWidth) {
		opts = append(opts, command.SeqEqualWidth)
	}
	return opts
}

// runMain is an indirection seam so main's wiring is testable without spawning
// the process; a test swaps it and restores it.
var runMain = clix.Main

func main() { runMain(spec, version) }
