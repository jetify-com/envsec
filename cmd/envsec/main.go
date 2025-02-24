// Copyright 2024 Jetify Inc. and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package main

import (
	"context"
	"os"

	"go.jetify.com/envsec/pkg/envcli"
)

func main() {
	os.Exit(envcli.Execute(context.Background()))
}
