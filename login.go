// SPDX-FileCopyrightText: 2024 Jeff Mitchell <jeffrey.mitchell@gmail.com>
// SPDX-License-Identifier: APL-2.0
package twinkly

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	uuid "github.com/hashicorp/go-uuid"
)

type LoginRequest struct {
	Challenge string `json:"challenge"`
}

type LoginResponse struct {
	AuthenticationToken string `json:"authentication_token"`
	ChallengeResponse   string `json:"challenge-response"`
	Code                Code   `json:"code"`
}

type VerifyRequest struct {
	ChallengeResponse string `json:"challenge-response"`
}

type VerifyResponse struct {
	Code Code `json:"code"`
}

// Login returns a client for the Twinkly API; if the call is successful the
// client will be logged in for the lifetime of the (non-renewable) token.
// Although built with an option pattern, currently it is mandatory to specify
// an IP address; this is in anticipation of future support for UDP discovery of
// devices.
func Login(ctx context.Context, opt ...Option) (*Client, error) {
	opts, err := getOpts(opt...)
	if err != nil {
		return nil, fmt.Errorf("error getting options: %w", err)
	}
	if opts.withHost == "" {
		return nil, fmt.Errorf("host is required")
	}

	challenge, err := uuid.GenerateRandomBytes(32)
	if err != nil {
		return nil, fmt.Errorf("error creating challenge: %w", err)
	}

	loginReq := &LoginRequest{
		Challenge: base64.StdEncoding.EncodeToString(challenge),
	}
	buf, err := json.Marshal(loginReq)
	if err != nil {
		return nil, fmt.Errorf("error marshalling login request: %w", err)
	}

	reqUrl := url.URL{
		Scheme: "http",
		Host:   opts.withHost,
		Path:   "/xled/v1/login",
	}
	httpReq, err := http.NewRequest("POST", reqUrl.String(), bytes.NewBuffer(buf))
	if err != nil {
		return nil, fmt.Errorf("error creating login request: %w", err)
	}
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending login request: %w", err)
	}
	defer httpResp.Body.Close()

	loginResp := &LoginResponse{}
	if err := json.NewDecoder(httpResp.Body).Decode(loginResp); err != nil {
		return nil, fmt.Errorf("error decoding login response: %w", err)
	}

	if loginResp.AuthenticationToken == "" {
		return nil, fmt.Errorf("no authentication token received")
	}
	if loginResp.ChallengeResponse == "" {
		return nil, fmt.Errorf("no challenge response received")
	}
	if loginResp.Code != CodeOk {
		return nil, fmt.Errorf("error code received during login: %d", loginResp.Code)
	}

	client := &Client{
		validatedHost:       opts.withHost,
		authenticationToken: loginResp.AuthenticationToken,
	}

	verifyReq := &VerifyRequest{
		ChallengeResponse: loginResp.ChallengeResponse,
	}
	var verifyResp VerifyResponse

	if err = client.doRequest(ctx, http.MethodPost, "/xled/v1/verify", verifyReq, &verifyResp); err != nil {
		return nil, fmt.Errorf("error during verify: %w", err)
	}

	if verifyResp.Code != CodeOk {
		return nil, fmt.Errorf("error code received during verify: %d", verifyResp.Code)
	}

	return client, nil
}
