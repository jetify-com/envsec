// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package main

import (
	"context"
	"os"

	"go.jetpack.io/envsec/internal/envcli"
)

func main() {
	os.Exit(envcli.Execute(context.Background()))
}
