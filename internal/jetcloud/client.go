package jetcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

type errorResponse struct {
	Error struct {
		Message string `json:"message,omitempty"`
	} `json:"error"`
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

	body, bodyReadErr := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// Non-200 responses can still have a JSON body with an error message.
		// Try to parse it and return that error.
		if bodyReadErr == nil {
			errResponse := errorResponse{}
			_ = json.Unmarshal(body, &errResponse)
			if errResponse.Error.Message != "" {
				return nil, errors.New(errResponse.Error.Message)
			}
		}
		return nil, fmt.Errorf("request failed %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	if bodyReadErr != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
