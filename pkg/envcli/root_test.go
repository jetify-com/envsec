// Copyright 2024 Jetify Inc. and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"context"
)

func ExampleExecute() {
	ctx := context.Background()
	Execute(ctx)
}
