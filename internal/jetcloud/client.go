package jetcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"go.jetpack.io/envsec/internal/auth"
	"go.jetpack.io/envsec/internal/envvar"
	"go.jetpack.io/envsec/internal/typeids"
	"golang.org/x/oauth2"
)

type client struct {
	apiHost string
}

func newClient() *client {
	return &client{
		apiHost: envvar.Get(
			"ENVSEC_API_HOST",
			"https://envsec-service-prod.cloud.jetpack.dev",
		),
	}
}

func (c *client) endpoint(path string) string {
	endpointURL, err := url.JoinPath(c.apiHost, path)
	if err != nil {
		panic(err)
	}
	return endpointURL
}

func (c *client) newProjectID(ctx context.Context, user *auth.User, repo, subdir string) (typeids.ProjectID, error) {
	fmt.Fprintf(os.Stderr, "Creating new project for repo=%s subdir=%s\n", repo, subdir)

	p, err := post[struct {
		ID typeids.ProjectID `json:"id"`
	}](ctx, c, user, map[string]string{
		"repo_url": repo,
		"subdir":   subdir,
	})
	if err != nil {
		return typeids.NilProjectID, err
	}

	return p.ID, nil
}

func post[T any](ctx context.Context, c *client, user *auth.User, data any) (*T, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: user.AccessToken.Raw},
	)
	httpClient := oauth2.NewClient(ctx, src)

	req, err := http.NewRequest(
		http.MethodPost,
		c.endpoint("projects"),
		bytes.NewBuffer(dataBytes),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("request failed %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
