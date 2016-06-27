package acmetool

import kingpin "gopkg.in/alecthomas/kingpin.v2"

type App struct {
	CommandLine *kingpin.Application
	Commands    map[string]func(Ctx)
}
