package main

import (
	"context"
	"log"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/noriah/catnip"
	"github.com/noriah/catnip/dsp"
	"github.com/noriah/catnip/dsp/window"
	"libdb.so/catnip-fyne/internal/catnipfyne"

	_ "github.com/noriah/catnip/input/all"
)

func main() {
	a := app.New()

	config := catnip.Config{
		Backend:      "pipewire",
		Device:       "spotify",
		SampleRate:   44100,
		SampleSize:   1024,
		ChannelCount: 2,
		SetupFunc: func() error {
			// TODO: output.Init with the right sampling sizes and windowing
			return nil
		},
		StartFunc: func(ctx context.Context) (context.Context, error) {
			return ctx, nil
		},
		CleanupFunc: func() error {
			return nil
		},
		Windower: window.Lanczos(),
	}

	output := catnipfyne.NewDisplay(config.SampleRate, config.SampleSize)
	config.Output = output

	config.Analyzer = dsp.NewAnalyzer(dsp.AnalyzerConfig{
		SampleRate: config.SampleRate,
		SampleSize: config.SampleSize,
		SquashLow:  true,
		BinMethod:  dsp.MaxSampleValue(),
	})

	config.Smoother = dsp.NewSmoother(dsp.SmootherConfig{
		SampleRate:      config.SampleRate,
		SampleSize:      config.SampleSize,
		ChannelCount:    config.ChannelCount,
		SmoothingFactor: 0.6415,
		SmoothingMethod: dsp.SmoothSimpleAverage,
	})

	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()

		// TODO: PR to swap the order of these
		if err := catnip.Run(&config, ctx); err != nil {
			panic(err)
		}
	}()

	defer log.Println("closing")

	w := a.NewWindow("Hello")
	w.SetContent(output.Raster) // this is required for some reason
	w.SetPadded(false)
	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}
