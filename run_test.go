package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name       string
		version    string
		wantOut    string
		wantErrSub string
		args       []string
		wantCode   int
	}{
		{
			name:    "single operand counts from one",
			args:    []string{"seq", "3"},
			wantOut: "1\n2\n3\n",
		},
		{
			name:    "two operands set first and last",
			args:    []string{"seq", "2", "5"},
			wantOut: "2\n3\n4\n5\n",
		},
		{
			name:    "three operands set first increment last",
			args:    []string{"seq", "1", "2", "7"},
			wantOut: "1\n3\n5\n7\n",
		},
		{
			name:    "separator joins into a single line",
			args:    []string{"seq", "-s", ", ", "1", "4"},
			wantOut: "1, 2, 3, 4\n",
		},
		{
			name:    "format applies printf style",
			args:    []string{"seq", "-f", "%.0f", "1", "3"},
			wantOut: "1\n2\n3\n",
		},
		{
			name:    "equal width zero pads",
			args:    []string{"seq", "-w", "8", "1", "11"},
			wantOut: "08\n09\n10\n11\n",
		},
		{
			name:    "version flag reports injected version",
			version: "1.2.3",
			args:    []string{"seq", "--version"},
			wantOut: "seq version 1.2.3\n",
		},
		{
			name:       "no operands errors",
			args:       []string{"seq"},
			wantCode:   1,
			wantErrSub: "seq:",
		},
		{
			name:       "too many operands errors",
			args:       []string{"seq", "1", "2", "3", "4"},
			wantCode:   1,
			wantErrSub: "seq:",
		},
		{
			name:       "non-numeric operand errors",
			args:       []string{"seq", "abc"},
			wantCode:   1,
			wantErrSub: "seq:",
		},
		{
			name:       "unknown flag errors",
			args:       []string{"seq", "--nope"},
			wantCode:   1,
			wantErrSub: "seq:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			code := run(tc.version, tc.args, strings.NewReader(""), &out, &errOut, afero.NewMemMapFs())

			if code != tc.wantCode {
				t.Fatalf("exit code = %d, want %d (stderr=%q)", code, tc.wantCode, errOut.String())
			}
			if tc.wantErrSub == "" && out.String() != tc.wantOut {
				t.Fatalf("stdout = %q, want %q", out.String(), tc.wantOut)
			}
			if tc.wantErrSub != "" && !strings.Contains(errOut.String(), tc.wantErrSub) {
				t.Fatalf("stderr = %q, want substring %q", errOut.String(), tc.wantErrSub)
			}
		})
	}
}

func Test_main(t *testing.T) {
	origExit, origRun := osExit, runCLI
	t.Cleanup(func() { osExit, runCLI = origExit, origRun })

	gotCode := -1
	osExit = func(code int) { gotCode = code }
	runCLI = func(string, []string, io.Reader, io.Writer, io.Writer, afero.Fs) int { return 7 }

	main()

	if gotCode != 7 {
		t.Fatalf("main propagated exit code %d, want 7", gotCode)
	}
}
