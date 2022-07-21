package main

import (
	"context"

	"go.jetpack.io/envsec/envcli"
)

func main() {
	envcli.Execute(context.Background())
}
