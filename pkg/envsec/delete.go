package envsec

import (
	"context"
	"strings"

	"go.jetpack.io/envsec/internal/tux"
)

func (e *Envsec) DeleteAll(ctx context.Context, envID EnvID, envNames ...string) error {
	if err := e.store.DeleteAll(ctx, envID, envNames); err != nil {
		return err
	}
	return tux.WriteHeader(e.Stderr,
		"[DONE] Deleted environment %s %v in environment: %s\n",
		tux.Plural(envNames, "variable", "variables"),
		strings.Join(tux.QuotedTerms(envNames), ", "),
		strings.ToLower(envID.EnvName),
	)
}
