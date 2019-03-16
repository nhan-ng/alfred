package main

import (
	"log"

	"github.com/nhan-ng/alfred/cmd/git/app"
)

func main() {
	command := app.NewGitHubCommand()
	if err := command.Execute(); err != nil {
		log.Fatalln(err)
	}
}
