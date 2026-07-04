#!/bin/sh
# Integration checks for yup-seq, run inside a Debian (GNU coreutils) container.
#
# parity CASE  — yup-seq must produce byte-identical output to GNU `seq`.
# assert WANT  — yup-seq must produce WANT exactly (used where yup-seq diverges
#                from GNU by design; see cmd-seq COMPATIBILITY.md).
set -eu

fails=0

parity() {
	ours=$(yup-seq "$@" 2>/dev/null || true)
	gnu=$(seq "$@" 2>/dev/null || true)
	if [ "$ours" = "$gnu" ]; then
		printf 'ok    parity  seq %s\n' "$*"
	else
		printf 'FAIL  parity  seq %s\n        gnu:  %s\n        ours: %s\n' "$*" "$gnu" "$ours"
		fails=$((fails + 1))
	fi
}

assert() {
	want=$1
	shift
	got=$(yup-seq "$@" 2>/dev/null || true)
	if [ "$got" = "$want" ]; then
		printf 'ok    assert  seq %s\n' "$*"
	else
		printf 'FAIL  assert  seq %s\n        want: %s\n        got:  %s\n' "$*" "$want" "$got"
		fails=$((fails + 1))
	fi
}

# Operand forms: LAST / FIRST LAST / FIRST INCREMENT LAST.
parity 5
parity 2 5
parity 1 2 10
parity 10 -2 1

# --separator (-s): GNU places the separator only between numbers.
parity -s , 1 5
parity -s ' ' 1 3

# --equal-width (-w): zero-pad to a common field width (matches GNU).
parity -w 8 11
parity -w 8 0.5 10

# --format (-f): printf-style float format (matches GNU).
parity -f %.2f 1 3

# Documented divergence: default rendering uses Go %g, so a fractional step
# does not promote whole numbers to a trailing-zero form (GNU prints 1.0 2.0).
assert "$(printf '1\n1.5\n2\n2.5')" 1 0.5 2.5

if [ "$fails" -ne 0 ]; then
	printf '\n%s check(s) failed\n' "$fails"
	exit 1
fi
printf '\nall checks passed\n'
