package main

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/jefferai/go-twinkly"
)

func main() {
	ctx := context.Background()
	host := "192.168.10.74"
	client, err := twinkly.Login(ctx, twinkly.WithHost(host))
	if err != nil {
		fmt.Printf("error logging in: %v\n", err)
		return
	}

	if err := client.SetLedOperationMode(ctx, twinkly.LedOperationModeMovie, 0); err != nil {
		fmt.Printf("error setting LED operation mode: %v\n", err)
		return
	}
	ledMovie := twinkly.LedMovie{
		Frames: []twinkly.Frame{
			{
				Leds: make([]twinkly.Led, 600),
			},
		},
	}

	ledColors := []twinkly.Led{
		twinkly.NewRgbLed(255, 0, 0),
		// twinkly.NewRgbLed(0, 255, 0),
		twinkly.NewRgbLed(0, 0, 255),
		twinkly.NewRgbLed(153, 50, 204),
		twinkly.NewRgbLed(255, 165, 0),
		twinkly.NewRgbLed(0, 255, 255),
		twinkly.NewRgbLed(255, 255, 0),
	}

	for i := 0; i < 600; i++ {
		ledMovie.Frames[0].Leds[i] = ledColors[rand.Int()%len(ledColors)]
	}

	movies, err := client.ListMovies(ctx)
	if err != nil {
		fmt.Printf("error listing movies: %v\n", err)
		return
	}
	fmt.Printf("movies: %#v\n", movies)

	movie := twinkly.Movie{
		Name:           "Twankles",
		UniqueId:       "F3712DC0-7FDE-4C6C-B9C8-813A5CBBC837",
		DescriptorType: "rgb_raw",
		LedsPerFrame:   600,
		FramesNumber:   1,
		Fps:            1,
	}

	if err := client.CreateMovie(ctx, movie, ledMovie); err != nil {
		fmt.Printf("error creating movie: %v\n", err)
		return
	}

	fmt.Println("done")
}
