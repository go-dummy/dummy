package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/cristalhq/acmd"

	"github.com/go-dummy/dummy/internal/config"
	"github.com/go-dummy/dummy/internal/logger"
	"github.com/go-dummy/dummy/internal/parse"
	"github.com/go-dummy/dummy/internal/server"
)

const version = "0.2.1"

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "server run error: %v\n", err)
	}
}

func run() error {
	cmds := []acmd.Command{
		{
			Name:        "server",
			Description: "run mock server",
			Do: func(ctx context.Context, args []string) error {
				cfg := config.NewConfig()

				cfg.Server.Path = args[0]

				fs := flag.NewFlagSet("dummy", flag.ContinueOnError)
				fs.StringVar(&cfg.Server.Port, "port", "8080", "")
				fs.StringVar(&cfg.Logger.Level, "logger-level", "INFO", "")
				if err := fs.Parse(args[1:]); err != nil {
					return err
				}

				api, err := parse.Parse(cfg.Server.Path)
				if err != nil {
					return fmt.Errorf("specification parse error: %w", err)
				}

				l := logger.NewLogger(cfg.Logger.Level)
				h := server.NewHandlers(api, l)
				s := server.NewServer(cfg.Server, l, h)

				return s.Run()
			},
		},
	}

	r := acmd.RunnerOf(cmds, acmd.Config{
		AppName: "Dummy",
		Version: version,
	})

	return r.Run()
}
