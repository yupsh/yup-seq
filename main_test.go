package main

import (
	"context"
	"errors"
	"testing"

	clix "github.com/gloo-foo/cli"
	"github.com/spf13/afero"
	urf "github.com/urfave/cli/v3"
)

// parse runs args through a bare command carrying the wrapper's flags and
// returns the parsed accessor.
func parse(t *testing.T, args ...string) *urf.Command {
	t.Helper()
	var got *urf.Command
	app := &urf.Command{
		Name:   name,
		Flags:  flags(),
		Action: func(_ context.Context, c *urf.Command) error { got = c; return nil },
	}
	if err := app.Run(context.Background(), args); err != nil {
		t.Fatalf("parse: %v", err)
	}
	return got
}

func TestOperands(t *testing.T) {
	cases := []struct {
		err  error
		name string
		args []string
		want int
	}{
		{ErrOperandCount, "none", []string{name}, 0},
		{ErrOperandCount, "four", []string{name, "1", "2", "3", "4"}, 0},
		{ErrInvalidOperand, "invalid", []string{name, "x"}, 0},
		{nil, "one", []string{name, "5"}, 1},
		{nil, "three", []string{name, "1", "2", "3"}, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nums, err := operands(parse(t, tc.args...))
			if !errors.Is(err, tc.err) {
				t.Fatalf("err=%v, want %v", err, tc.err)
			}
			if len(nums) != tc.want {
				t.Fatalf("nums len=%d, want %d", len(nums), tc.want)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want int
	}{
		{"none", []string{name, "1"}, 0},
		{"separator", []string{name, "-s", ",", "1"}, 1},
		{"format", []string{name, "-f", "%g", "1"}, 1},
		{"equal-width", []string{name, "-w", "1"}, 1},
		{"every", []string{name, "-s", ",", "-f", "%g", "-w", "1"}, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := len(options(parse(t, tc.args...))); got != tc.want {
				t.Fatalf("options len=%d, want %d", got, tc.want)
			}
		})
	}
}

func TestBuild_Source(t *testing.T) {
	src, filter, err := build(clix.Invocation{Args: parse(t, name, "3"), Fs: afero.NewMemMapFs()})
	if err != nil || src == nil || filter != nil {
		t.Fatalf("build: src=%v filter=%v err=%v (want source, nil filter)", src, filter, err)
	}
}

func TestBuild_OperandError(t *testing.T) {
	src, filter, err := build(clix.Invocation{Args: parse(t, name), Fs: afero.NewMemMapFs()})
	if !errors.Is(err, ErrOperandCount) {
		t.Fatalf("err=%v, want ErrOperandCount", err)
	}
	if src != nil || filter != nil {
		t.Fatalf("src=%v filter=%v, want both nil on error", src, filter)
	}
	if err.Error() != string(ErrOperandCount) {
		t.Fatalf("message=%q, want %q", err.Error(), string(ErrOperandCount))
	}
}

func Test_main(t *testing.T) {
	orig := runMain
	t.Cleanup(func() { runMain = orig })
	var gotName clix.Name
	runMain = func(s clix.Spec, _ clix.Version) { gotName = s.Name }
	main()
	if gotName != name {
		t.Fatalf("main used spec %q, want %s", gotName, name)
	}
}
