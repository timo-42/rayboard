package main

import (
	"context"
	"os"

	"github.com/timo-42/rayboard/internal/app"
)

func main() {
	os.Exit(app.Main(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}
