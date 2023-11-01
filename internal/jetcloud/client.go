package jetcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"go.jetpack.io/envsec/internal/envvar"
	"go.jetpack.io/envsec/internal/typeids"
	"go.jetpack.io/pkg/sandbox/auth/session"
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

func (c *client) newProjectID(ctx context.Context, tok *session.Token, repo, subdir string) (typeids.ProjectID, error) {
	p, err := post[struct {
		ID typeids.ProjectID `json:"id"`
	}](ctx, c, tok, "projects", map[string]string{
		"repo_url": repo,
		"subdir":   subdir,
	})
	if err != nil {
		return typeids.NilProjectID, err
	}

	return p.ID, nil
}

func (c *client) getAccessToken(
	ctx context.Context,
	tok *session.Token,
) (string, error) {
	p, err := post[struct {
		Token string `json:"token"`
	}](ctx, c, tok, "oauth/token", map[string]string{})
	if err != nil {
		return "", err
	}
	return p.Token, nil
}

func post[T any](ctx context.Context, client *client, tok *session.Token, path string, data any) (*T, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	src := oauth2.StaticTokenSource(&tok.Token)
	httpClient := oauth2.NewClient(ctx, src)

	req, err := http.NewRequest(
		http.MethodPost,
		client.endpoint(path),
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
