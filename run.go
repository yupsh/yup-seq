package main

import (
	"context"
	"fmt"
	"io"
	"strconv"

	command "github.com/gloo-foo/cmd-seq"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v3"
)

const (
	flagSeparator  = "separator"
	flagFormat     = "format"
	flagEqualWidth = "equal-width"
)

// usageText is the command's multi-line usage synopsis, shown in --help.
// cli/v3 indents the whole block by 3 spaces, so these lines are flush-left to
// stay aligned in the rendered output.
const usageText = `seq [OPTIONS] LAST
seq [OPTIONS] FIRST LAST
seq [OPTIONS] FIRST INCREMENT LAST

print numbers from FIRST to LAST, in steps of INCREMENT.`

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

// init replaces urfave/cli's default --version/-v flag with a --version-only
// flag, freeing the single-letter -v for command flags while still exposing
// the injected build version.
func init() {
	cli.VersionFlag = &cli.BoolFlag{Name: "version", Usage: "print version information and exit"}
}

// run builds and executes the seq CLI against the injected version, I/O, and
// filesystem, returning the process exit code. seq does not read stdin or the
// filesystem; both are injected for a uniform, testable wiring shape.
func run(version string, args []string, _ io.Reader, stdout, stderr io.Writer, _ afero.Fs) int {
	cmd := newCommand(version, stdout)
	cmd.Writer = stdout
	cmd.ErrWriter = stderr
	if err := cmd.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(stderr, "seq: %v\n", err)
		return 1
	}
	return 0
}

func newCommand(version string, stdout io.Writer) *cli.Command {
	return &cli.Command{
		Name:            "seq",
		Version:         version,
		Usage:           "print a sequence of numbers",
		UsageText:       usageText,
		HideHelpCommand: true,
		// Keep exit handling in run() rather than letting urfave/cli call
		// os.Exit, so the exit code stays testable.
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: flagSeparator, Aliases: []string{"s"}, Usage: "use STRING to separate numbers (default: \\n)"},
			&cli.StringFlag{Name: flagFormat, Aliases: []string{"f"}, Usage: "use printf style floating-point FORMAT"},
			&cli.BoolFlag{Name: flagEqualWidth, Aliases: []string{"w"}, Usage: "equalize width by padding with leading zeroes"},
		},
		Action: action(stdout),
	}
}

func action(stdout io.Writer) cli.ActionFunc {
	return func(_ context.Context, cmd *cli.Command) error {
		args, err := arguments(cmd)
		if err != nil {
			return err
		}
		_, err = gloo.Run(command.Seq(args...), gloo.ByteWriteTo(stdout))
		return err
	}
}

// arguments assembles the constructor arguments: the numeric operands followed
// by any option values selected via flags.
func arguments(cmd *cli.Command) ([]any, error) {
	nums, err := operands(cmd)
	if err != nil {
		return nil, err
	}
	return append(nums, options(cmd)...), nil
}

// operands parses the 1–3 numeric operands into constructor arguments.
func operands(cmd *cli.Command) ([]any, error) {
	if cmd.NArg() < 1 || cmd.NArg() > 3 {
		return nil, ErrOperandCount
	}
	nums := make([]any, cmd.NArg())
	for i := range nums {
		val, err := strconv.ParseFloat(cmd.Args().Get(i), 64)
		if err != nil {
			return nil, ErrInvalidOperand
		}
		nums[i] = val
	}
	return nums, nil
}

// options translates the selected flags into constructor option values.
func options(cmd *cli.Command) []any {
	var opts []any
	if cmd.IsSet(flagSeparator) {
		opts = append(opts, command.SeqSeparator(cmd.String(flagSeparator)))
	}
	if cmd.IsSet(flagFormat) {
		opts = append(opts, command.SeqFormat(cmd.String(flagFormat)))
	}
	if cmd.Bool(flagEqualWidth) {
		opts = append(opts, command.SeqEqualWidth)
	}
	return opts
}
