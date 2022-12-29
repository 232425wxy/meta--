package main

import "github.com/232425wxy/meta--/cmd/commands"

func main() {
	if err := commands.DockerNetCmd.Execute(); err != nil {
		panic(err)
	}
}
