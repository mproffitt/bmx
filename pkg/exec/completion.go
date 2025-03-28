// Copyright (c) 2025 Martin Proffitt <mprooffitt@choclab.net>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package exec

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

// MissingZshError is returned when Zsh is not found in the users environment
//
// Applications implementing this package can use this todisable the completion
// system
type MissingZshError struct{}

func (e MissingZshError) Error() string {
	return "missing zsh"
}

// ExecTimeoutError is returned when the completion takes longer than the
// anticipated execution time of `maxDuration`
//
// Errors of this type would normally be transient as the completion system
// generally returns long before the deadline is reached.
type ExecTimeoutError struct{}

func (e ExecTimeoutError) Error() string {
	return "deadline exceeded"
}

// Completion represents a completion object
type Completion struct {
	Option      string
	Description string
}

// We use a context timeout for executing shell completions by default this is
// set to 5 seconds to allow commands which rely on remote executions such as
// kubectl time to complete.
const maxDuration = 5 * time.Second

// The Capture script is a modified version of the `zsh-capture-completion`
// script from https://github.com/Valodim/zsh-capture-completion.
//
// This script can be bound to an input field and will execute the completion
// commands for a given input.
//
// It should be noted that it is not worth binding to every keypress, but should
// be bound to the space character, and optionally hyphen `-` and slash `/`
// characters.
//
// The script will only work in environments where ZSH is installed although
// this does not have to be the users primary shell as long as the completions
// are also installed alongside
//
// Modifications are made to ensure custom completions defined by zsh.completion
// and oh-my-zsh are loaded and an additional set of kubernetes specific command
// completions are also sourced.
const capture = `
#!/bin/zsh
zmodload zsh/zpty || { echo 'error: missing module zsh/zpty' >&2; exit 1 }

# spawn shell
zpty z zsh -f -i

# line buffer for pty output
local line

setopt rcquotes
() {
    zpty -w z source $1
    repeat 4; do
        zpty -r z line
        [[ $line == ok* ]] && return
    done
    echo 'error initializing.' >&2
    exit 2
} =( <<< '
# no prompt!
PROMPT=

# Check for custom completions for common frameworks
if [[ -d "${HOME}/.oh-my-zsh/custom/completions" ]]; then
    fpath+=("${HOME}/.oh-my-zsh/custom/completions")
fi

if [[ -d "$HOME/.zsh-completions" ]]; then
    fpath+=("$HOME/.zsh-completions/src")
fi

# load completion system
autoload compinit
compinit -d ~/.zcompdump_capture

# Load in any custom commands we want to ensure are
# part of the completion system
typeset -A cmds=(
  [kubectl]=''source <(kubectl completion zsh)''
  [stern]=''source <(stern --completion=zsh)''
  [helm]=''source <(helm completion zsh)''
  [tsh]=''source <(tsh --completion-script-zsh)''
  [flux]=''source <(flux completion zsh)''
)

# Iterate over each command to source its completion script
for key value in ${(kv)cmds}; do
  if command -v $key >/dev/null 2>&1; then
    eval ${value}
  fi
done

# disable any option to execute the command within
# this script by disabling carriage return and
# line feed characters
bindkey ''^M'' undefined
bindkey ''^J'' undefined

# Ensure TAB is bound to complete-word
bindkey ''^I'' complete-word

# send a line with null-byte at the end before and after completions are output
null-line () {
    echo -E - $''\0''
}
compprefuncs=( null-line )
comppostfuncs=( null-line exit )

# never group stuff!
zstyle '':completion:*'' list-grouped false
# don''t insert tab when attempting completion on empty line
# zstyle '':completion:*'' insert-tab false
# no list separator, this saves some stripping later on
zstyle '':completion:*'' list-separator ''''
zstyle '':completion:*'' menu no

# we use zparseopts
zmodload zsh/zutil

# override compadd (this our hook)
compadd () {

    # check if any of -O, -A or -D are given
    if [[ ${@[1,(i)(-|--)]} == *-(O|A|D)\ * ]]; then
        # if that is the case, just delegate and leave
        builtin compadd "$@"
        return $?
    fi

    # ok, this concerns us!
    # echo -E - got this: "$@"

    # be careful with namespacing here, we don''t want to mess with stuff that
    # should be passed to compadd!
    typeset -a __hits __dscr __tmp

    # do we have a description parameter?
    # note we don''t use zparseopts here because of combined option parameters
    # with arguments like -default- confuse it.
    if (( $@[(I)-d] )); then # kind of a hack, $+@[(r)-d] doesn''t work because of line noise overload
        # next param after -d
        __tmp=${@[$[${@[(i)-d]}+1]]}
        # description can be given as an array parameter name, or inline () array
        if [[ $__tmp == \(* ]]; then
            eval "__dscr=$__tmp"
        else
            __dscr=( "${(@P)__tmp}" )
        fi
    fi

    # capture completions by injecting -A parameter into the compadd call.
    # this takes care of matching for us.
    builtin compadd -A __hits -D __dscr "$@"

    # set additional required options
    setopt localoptions norcexpandparam extendedglob

    # extract prefixes and suffixes from compadd call. we can''t do zsh''s cool
    # -r remove-func magic, but it''s better than nothing.
    typeset -A apre hpre hsuf asuf
    zparseopts -E P:=apre p:=hpre S:=asuf s:=hsuf

    # append / to directories? we are only emulating -f in a half-assed way
    # here, but it''s better than nothing.
    integer dirsuf=0
    # don''t be fooled by -default- >.>
    if [[ -z $hsuf && "${${@//-default-/}% -# *}" == *-[[:alnum:]]#f* ]]; then
        dirsuf=1
    fi

    # just drop
    [[ -n $__hits ]] || return

    # this is the point where we have all matches in $__hits and all
    # descriptions in $__dscr!

    # display all matches
    local dsuf dscr
    for i in {1..$#__hits}; do
        # add a dir suffix?
        (( dirsuf )) && [[ -d $__hits[$i] ]] && dsuf=/ || dsuf=
        # description to be displayed afterwards
        (( $#__dscr >= $i )) && dscr=" -- ${${__dscr[$i]}##$__hits[$i] #}" || dscr=

        echo -E - $IPREFIX$apre$hpre$__hits[$i]$dsuf$hsuf$asuf$dscr
    done
}

# signal success!
echo ok')

zpty -w z "$1"$'\t'
integer tog=0
# read from the pty, and parse linewise
while zpty -r z; do :; done | while IFS= read -r line; do
    if [[ $line == *$'\0\r' ]]; then
        (( tog++ )) && return 0 || continue
    fi
    # display between toggles
    (( tog )) && echo -E - $line
done

return 2
`

// HasZsh returns true if ZSH is installed and accessible through $PATH
func HasZsh() bool {
	_, err := exec.LookPath("zsh")
	return err != nil
}

// ZshCompletions matches completions based on the `in` string.
//
// This triggers the completion script and captures the completions given by
// the ZSH subshell environment, returning a slice of Completion objects which
// contain the completion option and an optional description if one is provided
// for that option.
func ZshCompletions(in string) (out []Completion, err error) {
	// Check to see if zsh exists in the users environment
	var zsh string
	{
		// Check to see if zsh is installed in the system
		if zsh, err = exec.LookPath("zsh"); err != nil {
			err = MissingZshError{}
			return
		}
	}

	// Create a temporary file to hold the script contents
	var tmpFile *os.File
	{
		tmpFile, err = os.CreateTemp("/tmp", "script-*.zsh")
		if err != nil {
			err = fmt.Errorf("Error creating temp file %w", err)
			return
		}
	}
	name := tmpFile.Name()
	defer func() {
		if err := os.Remove(name); err != nil {
			log.Error("failed to clean up temporary file", "tmpfile", name, "error", err)
		}
	}()

	{
		if _, err = tmpFile.Write([]byte(capture)); err != nil {
			return
		}
		if err := tmpFile.Close(); err != nil {
			log.Error("failed to close file", "name", name, "error", err)
		}
	}

	var stdout, stderr strings.Builder
	{
		ctx, cancel := context.WithTimeout(context.Background(), maxDuration)
		defer cancel()

		cmd := exec.CommandContext(ctx, zsh, tmpFile.Name(), in)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err = cmd.Run(); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return []Completion{}, ExecTimeoutError{}
			}
			return []Completion{}, &BmxExecError{
				Stdout: stdout.String(),
				Stderr: stderr.String(),
				error:  err,
			}
		}
	}
	out = make([]Completion, 0)
	for _, line := range strings.Split(stdout.String(), "\n") {
		p := strings.Split(line, " -- ")
		option := strings.TrimSpace(p[0])

		var description string
		if len(p) == 2 {
			description = strings.TrimSpace(p[1])
		}
		if len(option) == 0 {
			continue
		}

		// When handling hyphens, we need to strip from the start of the option any
		// that have already been typed.
		//
		// This means that we cannot use strings.TrimLeft to complete it as that
		// will simply remove all leading hyphens causing a mismatch between
		// what was typed and the option being presented as a suggestion to the
		// user
		//
		// We solve this by running the string through a custom function for
		// trimming hyphens given below.
		option = trimHyphens(in, option)
		option = strings.TrimPrefix(option, in)

		out = append(out, Completion{
			Option:      in + option,
			Description: description,
		})
	}
	return
}

// This function trims matching hyphens from the
// start of `b` that are found at the end of `a`
func trimHyphens(a, b string) string {
	i, j := len(a)-1, 0
	var trim bool
	for {
		if i == 0 || j >= len(b) {
			break
		} else if a[i] != '-' || b[j] != '-' {
			trim = true
			break
		}
		i--
		j++
	}
	if trim {
		return b[j:]
	}
	return b
}
