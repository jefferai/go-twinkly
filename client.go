// SPDX-FileCopyrightText: 2024 Jeff Mitchell <jeffrey.mitchell@gmail.com>
// SPDX-License-Identifier: APL-2.0

package twinkly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type GenericCodeResponse struct {
	Code Code `json:"code"`
}

type Client struct {
	validatedHost       string
	authenticationToken string
}

func (c *Client) AuthenticationToken() string {
	return c.authenticationToken
}

func (c *Client) ValidatedHost() string {
	return c.validatedHost
}

func (c *Client) doRequest(ctx context.Context, method, path string, reqBody, respBody any, opt ...Option) error {
	opts, err := getOpts(opt...)
	if err != nil {
		return fmt.Errorf("error getting options: %w", err)
	}

	var reader io.Reader
	if reqBody != nil {
		switch opts.withContentType {
		case "application/octet-stream":
			bodyBytes, ok := reqBody.([]byte)
			if !ok {
				return fmt.Errorf("error casting request body to []byte, type is %T", reqBody)
			}
			reader = bytes.NewReader(bodyBytes)
		case "application/json":
			fallthrough
		default:
			buf, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("error marshalling request: %w", err)
			}
			reader = bytes.NewBuffer(buf)
		}
	}

	reqUrl := url.URL{
		Scheme: "http",
		Host:   c.validatedHost,
		Path:   path,
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, reqUrl.String(), reader)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	if c.authenticationToken != "" {
		httpReq.Header.Set("X-Auth-Token", c.authenticationToken)
	}
	if opts.withContentType != "" {
		httpReq.Header.Set("Content-Type", opts.withContentType)
	}
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer httpResp.Body.Close()

	if err := json.NewDecoder(httpResp.Body).Decode(respBody); err != nil {
		return fmt.Errorf("error decoding response: %w", err)
	}

	return nil
}
