// SPDX-FileCopyrightText: 2024 Jeff Mitchell <jeffrey.mitchell@gmail.com>
// SPDX-License-Identifier: APL-2.0
package twinkly

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type Led interface {
	Red() uint8
	Green() uint8
	Blue() uint8
	White() uint8
	HasWhite() bool
}

type RgbLed struct {
	red   uint8
	green uint8
	blue  uint8
}

func (l RgbLed) Red() uint8 {
	return l.red
}

func (l RgbLed) Green() uint8 {
	return l.green
}

func (l RgbLed) Blue() uint8 {
	return l.blue
}

func (l RgbLed) White() uint8 {
	return 0
}

func (l RgbLed) HasWhite() bool {
	return false
}

type RgbwLed struct {
	RgbLed
	white uint8
}

func (l RgbwLed) White() uint8 {
	return l.white
}

func (l RgbwLed) HasWhite() bool {
	return true
}

func NewRgbLed(red, green, blue uint8) RgbLed {
	return RgbLed{
		red:   red,
		green: green,
		blue:  blue,
	}
}

func NewRgbwLed(red, green, blue, white uint8) RgbwLed {
	return RgbwLed{
		RgbLed: RgbLed{
			red:   red,
			green: green,
			blue:  blue,
		},
		white: white,
	}
}

type Frame struct {
	Leds []Led
}

type LedMovie struct {
	Frames []Frame
}

func (m LedMovie) Marshal() ([]byte, error) {
	if len(m.Frames) == 0 {
		return nil, errors.New("movie must have at least one frame")
	}
	// Verify that the LED profiles are not being mixed
	var prevHasWhite *bool
	var prevLedCount *int
	for i, frame := range m.Frames {
		if len(frame.Leds) == 0 {
			return nil, fmt.Errorf("frame %d has no LEDs", i)
		}
		switch prevLedCount {
		case nil:
			prevLedCount = new(int)
			*prevLedCount = len(frame.Leds)
		default:
			if *prevLedCount != len(frame.Leds) {
				return nil, errors.New("inconsistent LED count in movie frames")
			}
		}
		for _, led := range frame.Leds {
			switch prevHasWhite {
			case nil:
				prevHasWhite = new(bool)
				*prevHasWhite = led.HasWhite()
			default:
				if *prevHasWhite != led.HasWhite() {
					return nil, errors.New("inconsistent white LED availability in movie leds")
				}
			}
		}
	}
	// Calculate size of buffer needed
	ledSize := 3
	if *prevHasWhite {
		ledSize = 4
	}
	buf := make([]byte, len(m.Frames)*len(m.Frames[0].Leds)*ledSize)
	// Populate buffer
	for i, frame := range m.Frames {
		for j, led := range frame.Leds {
			offset := i*len(frame.Leds)*ledSize + j*ledSize
			buf[offset] = led.Red()
			buf[offset+1] = led.Green()
			buf[offset+2] = led.Blue()
			if *prevHasWhite {
				buf[offset+3] = led.White()
			}
		}
	}
	return buf, nil
}

type UploadFullMovieResponse struct {
	Code         Code `json:"code"`
	FramesNumber int  `json:"frames_number"`
}

type SetLedMovieConfigRequest struct {
	FrameDelay   int `json:"frame_delay"`
	LedsNumber   int `json:"leds_number"`
	FramesNumber int `json:"frames_number"`
}

func (c *Client) UploadFullMovie(ctx context.Context, m LedMovie) error {
	buf, err := m.Marshal()
	if err != nil {
		return fmt.Errorf("error marshalling movie: %w", err)
	}
	// Send buffer to device
	var resp UploadFullMovieResponse
	if err := c.doRequest(ctx, http.MethodPost, "/xled/v1/led/movie/full", buf, &resp, WithContentType("application/octet-stream")); err != nil {
		return fmt.Errorf("error uploading full movie: %w", err)
	}
	if resp.FramesNumber != len(m.Frames) {
		return fmt.Errorf("unexpected number of frames in response: %d, input frames: %d", resp.FramesNumber, len(m.Frames))
	}
	// Configure movie
	req := &SetLedMovieConfigRequest{
		FrameDelay:   1000 / len(m.Frames),
		LedsNumber:   len(m.Frames[0].Leds),
		FramesNumber: len(m.Frames),
	}
	var configResp GenericCodeResponse
	if err := c.doRequest(ctx, http.MethodPost, "/xled/v1/led/movie/config", req, &configResp); err != nil {
		return fmt.Errorf("error configuring movie: %w", err)
	}
	if configResp.Code != CodeOk {
		return fmt.Errorf("error code configuring movie: %v", configResp.Code)
	}
	return nil
}

type Movie struct {
	Id             int    `json:"id"`
	Name           string `json:"name"`
	UniqueId       string `json:"unique_id"`
	DescriptorType string `json:"descriptor_type"`
	LedsPerFrame   int    `json:"leds_per_frame"`
	FramesNumber   int    `json:"frames_number"`
	Fps            int    `json:"fps"`
}

type ListMoviesResponse struct {
	Code            Code    `json:"code"`
	Movies          []Movie `json:"movies"`
	AvailableFrames int     `json:"available_frames"`
	MaxCapacity     int     `json:"max_capacity"`
}

func (c *Client) ListMovies(ctx context.Context) (*ListMoviesResponse, error) {
	var resp ListMoviesResponse
	if err := c.doRequest(ctx, http.MethodGet, "/xled/v1/movies", nil, &resp); err != nil {
		return nil, fmt.Errorf("error listing movies: %w", err)
	}
	if resp.Code != CodeOk {
		return nil, fmt.Errorf("error code listing movies: %v", resp.Code)
	}
	return &resp, nil
}

func (c *Client) CreateMovie(ctx context.Context, m Movie, l LedMovie) error {
	// Marshal LED movie
	buf, err := l.Marshal()
	if err != nil {
		return fmt.Errorf("error marshalling LED movie: %w", err)
	}

	var resp GenericCodeResponse
	if err := c.doRequest(ctx, http.MethodPost, "/xled/v1/movies/new", m, &resp); err != nil {
		return fmt.Errorf("error creating movie entry: %w", err)
	}
	switch resp.Code {
	case CodeDuplicateUniqueId:
		return fmt.Errorf("movie with unique ID %q already exists", m.UniqueId)
	case CodeOk:
	default:
		return fmt.Errorf("error code creating movie entry: %v", resp.Code)
	}

	if err := c.doRequest(ctx, http.MethodPost, "/xled/v1/movies/full", buf, &resp, WithContentType("application/octet-stream")); err != nil {
		return fmt.Errorf("error creating movie entry: %w", err)
	}
	if resp.Code != CodeOk {
		return fmt.Errorf("error code creating movie entry: %v", resp.Code)
	}

	return nil
}
