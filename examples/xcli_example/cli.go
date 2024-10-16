package main

import (
	"context"
	"os"

	"github.com/seefs001/xox/xcli"
	"github.com/seefs001/xox/xlog"
)

func _main() {
	app := xcli.NewApp("test", "test application", "1.0.0")

	// Enable debug mode if needed
	xcli.EnableDebug(true)

	// Add a sample command
	app.AddCommand(&xcli.Command{
		Name:        "greet",
		Description: "Greet the user",
		Run: func(ctx context.Context, cmd *xcli.Command, args []string) error {
			if len(args) > 0 {
				xlog.Infof("Hello, %s!", args[0])
			} else {
				xlog.Info("Hello, World!")
			}
			return nil
		},
	})

	// Set the default run function
	app.SetDefaultRun(func(ctx context.Context, app *xcli.App) error {
		app.PrintHelp()
		return nil
	})

	// Run the application
	if err := app.Run(context.Background(), os.Args); err != nil {
		xlog.Error(err.Error())
		os.Exit(1)
	}
}
