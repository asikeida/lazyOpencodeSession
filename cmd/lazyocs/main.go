package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/asikeida/lazyOpencodeSession/internal/app"
)

var version = "dev"

func main() {
	var cli app.Options
	showVersion := flag.Bool("version", false, "print version")
	check := flag.Bool("check", false, "check database access without launching TUI")
	printConfig := flag.Bool("print-config", false, "print sample config")
	flag.StringVar(&cli.ConfigPath, "config", "", "path to lazyocs config file")
	flag.StringVar(&cli.DBPath, "db", "", "path to OpenCode SQLite database")
	flag.IntVar(&cli.Limit, "limit", 500, "maximum sessions to load")
	flag.StringVar(&cli.Theme, "theme", "default", "theme name: default or dark")
	flag.StringVar(&cli.Language, "language", "auto", "language: auto, en, or zh-CN")
	flag.StringVar(&cli.OpenCodeCommand, "opencode", "opencode", "opencode command path")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}
	if *printConfig {
		if err := app.PrintDefaultConfig(os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	opts, err := app.ResolveOptions(cli, visitedFlags())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *check {
		if err := app.Check(context.Background(), opts); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if err := app.Run(context.Background(), opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func visitedFlags() map[string]bool {
	set := map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		set[f.Name] = true
	})
	return set
}
