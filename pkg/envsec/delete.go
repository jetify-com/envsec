package envsec

import (
	"context"
	"strings"

	"go.jetify.com/envsec/internal/tux"
)

func (e *Envsec) DeleteAll(ctx context.Context, envNames ...string) error {
	if err := e.Store.DeleteAll(ctx, e.EnvID, envNames); err != nil {
		return err
	}
	return tux.WriteHeader(e.Stderr,
		"[DONE] Deleted environment %s %v in environment: %s\n",
		tux.Plural(envNames, "variable", "variables"),
		strings.Join(tux.QuotedTerms(envNames), ", "),
		strings.ToLower(e.EnvID.EnvName),
	)
}
