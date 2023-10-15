package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

func NewClient(ctx context.Context, token string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("github token must be provided")
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, src)
	return &Client{httpClient}, nil
}

type errors []struct {
	Message string
}

// Error implements error interface.
func (e errors) Error() string {
	return e[0].Message
}

type Client struct {
	c *http.Client
}

func (c Client) Query(query string) ([]byte, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(struct {
		Query string `json:"query"`
	}{
		Query: query,
	})
	if err != nil {
		return nil, err
	}

	const url = "https://api.github.com/graphql"
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return nil, err
	}

	resp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("non-200 OK status code: %v body: %q", resp.Status, body)
	}
	var out struct {
		Data   *json.RawMessage
		Errors errors
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Errors) > 0 {
		return nil, out.Errors
	}

	return *out.Data, nil
}
