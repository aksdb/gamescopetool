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

	if len(args) >= 3 && args[0] == "-client" {
		result.ClientSocket = args[1]
		result.GameAndArgs = args[2:]
		return result
	}

	if strings.HasPrefix(args[0], "-") {
		result.GamescopeArgs = args
		for i := range args {
			if args[i] == "--" {
				if i+1 < len(args) {
					result.GamescopeArgs = args[:i]
					result.GameAndArgs = args[i+1:]
				}
				break
			}
		}
	} else {
		result.GameAndArgs = args
	}

	return result
}
