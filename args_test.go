package main

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		Name     string
		Args     []string
		Expected Args
	}{
		{
			Name:     "No args given",
			Args:     nil,
			Expected: Args{},
		},
		{
			Name:     "Only executable given.",
			Args:     []string{"foobar"},
			Expected: Args{},
		},
		{
			Name: "Only gamescope args given",
			Args: []string{"foobar", "--help", "-x", "1", "-y", "2"},
			Expected: Args{
				Self:          "foobar",
				GamescopeArgs: []string{"--help", "-x", "1", "-y", "2"},
			},
		},
		{
			Name: "Only game args given",
			Args: []string{"foobar", "thegame", "--launch", "--steam"},
			Expected: Args{
				Self:        "foobar",
				GameAndArgs: []string{"thegame", "--launch", "--steam"},
			},
		},
		{
			Name: "Gamescope and game args given",
			Args: []string{"foobar", "-w", "1920", "-f", "--", "thegame", "--launch", "--steam"},
			Expected: Args{
				Self:          "foobar",
				GamescopeArgs: []string{"-w", "1920", "-f"},
				GameAndArgs:   []string{"thegame", "--launch", "--steam"},
			},
		},

		{
			Name: "Run as client",
			Args: []string{"foobar", "-client", "xyz", "thegame", "--launch", "--steam"},
			Expected: Args{
				Self:         "foobar",
				GameAndArgs:  []string{"thegame", "--launch", "--steam"},
				ClientSocket: "xyz",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			args := ParseArgs(tt.Args)
			td.Cmp(t, args, tt.Expected)
		})
	}
}
