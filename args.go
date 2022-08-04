package main

import (
	"strings"
)

type Args struct {
	Self          string
	GamescopeArgs []string
	GameAndArgs   []string

	ClientSocket    string
	ShowDummyWindow bool
}

func ParseArgs(args []string) Args {
	var result Args
	if len(args) < 2 {
		return result
	}

	result.Self = args[0]

	args = args[1:]

	if len(args) >= 3 && args[0] == "--client" {
		result.ClientSocket = args[1]

		i := 2
	loop:
		for i < len(args) {
			switch args[i] {
			case "--dummy-window":
				result.ShowDummyWindow = true
				i++
			default:
				break loop
			}
		}

		result.GameAndArgs = args[i:]
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

	for i := 0; i < len(result.GamescopeArgs); {
		switch result.GamescopeArgs[i] {
		case "--dummy-window":
			result.ShowDummyWindow = true
		default:
			i++
			continue
		}

		result.GamescopeArgs = append(result.GamescopeArgs[:i], result.GamescopeArgs[i+1:]...)
	}

	return result
}
