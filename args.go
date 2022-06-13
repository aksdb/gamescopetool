package main

import (
	"strings"
)

type Args struct {
	Self          string
	GamescopeArgs []string
	GameAndArgs   []string

	ClientSocket string
}

func ParseArgs(args []string) Args {
	var result Args
	if len(args) < 2 {
		return result
	}

	result.Self = args[0]

	args = args[1:]
	if strings.HasPrefix(args[0], "-") {
	loop:
		for i := 0; i < len(args); i++ {
			switch {
			case args[i] == "-client" && i+1 < len(args):
				result.ClientSocket = args[i+1]
				i++
			case args[i] == "--":
				if i+1 < len(args) {
					result.GameAndArgs = args[i+1:]
				}
				break loop
			default:
				result.GamescopeArgs = append(result.GamescopeArgs, args[i])
			}
		}
	} else {
		result.GameAndArgs = args
	}

	return result
}
