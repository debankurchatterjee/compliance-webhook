package main

import (
	"context"
	"github.com/compliance-webhook/pkg/app"
)

func main() {
	ctx := context.Background()
	app.Run(ctx)
}
