// SPDX-FileCopyrightText: 2024 Jeff Mitchell <jeffrey.mitchell@gmail.com>
// SPDX-License-Identifier: APL-2.0

package twinkly

import (
	"context"
	"fmt"
	"net/http"
)

type GetLedOperationModeResponse struct {
	Mode LedOperationMode `json:"mode"`
	Code Code             `json:"code"`
}

type SetLedOperationModeRequest struct {
	Mode     LedOperationMode `json:"mode"`
	EffectId int              `json:"effect_id"`
}

func (c *Client) GetLedOperationMode(ctx context.Context) (LedOperationMode, error) {
	var resp GetLedOperationModeResponse
	if err := c.doRequest(ctx, http.MethodGet, "/xled/v1/led/mode", nil, &resp); err != nil {
		return "", fmt.Errorf("error getting LED operation mode: %w", err)
	}
	if resp.Code != CodeOk {
		return "", fmt.Errorf("error code getting LED operation mode: %v", resp.Code)
	}
	return resp.Mode, nil
}

func (c *Client) SetLedOperationMode(ctx context.Context, mode LedOperationMode, effectId int) error {
	req := &SetLedOperationModeRequest{
		Mode:     mode,
		EffectId: effectId,
	}
	var resp GenericCodeResponse
	if err := c.doRequest(ctx, http.MethodPost, "/xled/v1/led/mode", req, &resp); err != nil {
		return fmt.Errorf("error getting LED operation mode: %w", err)
	}
	if resp.Code != CodeOk {
		return fmt.Errorf("error code setting LED operation mode: %v", resp.Code)
	}
	return nil
}
