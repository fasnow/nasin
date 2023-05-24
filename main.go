package main

import "nasin/cmd"

func main() {
	config := cmd.FlagConfig{}
	config.ParseFlags()
}
