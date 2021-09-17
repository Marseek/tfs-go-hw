package main

import (
	"fmt"
)

type Mod = func()

func main() {
	var d int
	var charact rune
	var n int

	color := func(col string) func() {
		return func() {
			switch {
			case col == "black":
				d = 30
			case col == "red":
				d = 31
			case col == "green":
				d = 32
			case col == "brown":
				d = 33
			case col == "blue":
				d = 34
			case col == "purple":
				d = 35
			case col == "cyan":
				d = 36
			}
		}
	}
	char := func(c rune) func() {
		return func() {
			charact = c
		}
	}

	size := func(i int) func() {
		return func() {
			n = i
		}
	}
	sandglass := func(args ...Mod) {
		// seting default params
		d = 37
		charact = 'X'
		n = 15

		// handling options
		for _, mods := range args {
			mods()
		}

		// printing the sandglass
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				if (j == i) || (j == n-i-1) || (i == 0) || (i == n-1) {
					fmt.Printf("\033[%dm%c\033[0m", d, charact)
				} else {
					fmt.Printf(" ")
				}
			}
			fmt.Printf("\n")
		}
	}

	sandglass()
	sandglass(size(14), char('Q'), color("red"))
}
