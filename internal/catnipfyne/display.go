package catnipfyne

import (
	"image"
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2/canvas"
	"github.com/llgcode/draw2d"
	"github.com/noriah/catnip/input"
	"libdb.so/catnip-fyne/internal/vecd"

	window "github.com/noriah/catnip/util"
)

const ScalingWindow = 1.5 // seconds

const PeakThreshold = 0.01

// Display is a display of audio data.
type Display struct {
	*canvas.Raster
	drawer *vecd.Context // not thread-safe!
	buffer *vecd.DoubleBuffer
	window *window.MovingWindow

	lock sync.Mutex

	binsBuffer [][]float64
	nchannels  int
	peak       float64

	barWidth   float64
	spaceWidth float64
	binWidth   float64

	width  int
	height int

	zeroes int
}

// NewDisplay creates a new display.
func NewDisplay(sampleRate float64, sampleSize int) *Display {
	windowSize := ((int(ScalingWindow * sampleRate)) / sampleSize) * 2

	d := &Display{}
	d.buffer = &vecd.DoubleBuffer{}
	d.window = window.NewMovingWindow(windowSize)
	d.drawer = vecd.NewContext()
	d.Raster = canvas.NewRaster(func(x, y int) image.Image {
		d.lock.Lock()
		defer d.lock.Unlock()

		d.drawer.Resize(x, y)
		d.drawer.Clear()

		d.drawer.SetStrokeColor(color.RGBA{255, 255, 255, 255})
		d.drawer.SetLineWidth(25)
		d.drawer.SetLineCap(draw2d.RoundCap)

		d.width = x
		d.height = y
		d.draw()

		return d.buffer.Swap(d.drawer.Image())
	})
	d.Raster.ScaleMode = canvas.ImageScaleFastest

	d.SetSizes(75, 5)
	return d
}

// SetSizes sets the sizes of the bars and spaces in the display.
func (d *Display) SetSizes(bar, space float64) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.barWidth = bar
	d.spaceWidth = space
	d.binWidth = bar + space
}

// Write implements processor.Output.
func (d *Display) Write(bins [][]float64, nchannels int) error {
	defer d.Refresh()

	d.lock.Lock()
	defer d.lock.Unlock()

	if len(d.binsBuffer) < len(bins) || len(d.binsBuffer[0]) < len(bins[0]) {
		d.binsBuffer = input.MakeBuffers(len(bins), len(bins[0]))
	}
	input.CopyBuffers(d.binsBuffer, bins)

	nbins := d.bins(nchannels)
	var peak float64

	for i := 0; i < nchannels; i++ {
		for _, val := range bins[i][:nbins] {
			if val > peak {
				peak = val
			}
		}
	}

	d.peak = peak
	d.nchannels = nchannels

	return nil
}

// Bins implements processor.Output.
func (d *Display) Bins(nchannels int) int {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.bins(nchannels)
}

func (d *Display) bins(nchannels int) int {
	return d.width / int(d.binWidth)
}

func (d *Display) draw() {
	scale := 1.0 // TODO

	if d.peak >= PeakThreshold {
		d.zeroes = 0

		// do some scaling if we are above the PeakThreshold
		vMean, vSD := d.window.Update(d.peak)
		if t := vMean + (2.0 * vSD); t > 1.0 {
			scale = t
		}
	} else {
		if d.zeroes++; d.zeroes == 5 {
			d.window.Recalculate()
		}
	}

	d.drawHorizontally(d.binsBuffer, d.nchannels, scale)
}

func (d *Display) drawHorizontally(bins [][]float64, nchannels int, absScale float64) {
	wf := float64(d.width)
	hf := float64(d.height)

	delta := 1
	scale := hf / absScale
	nbars := d.bins(d.nchannels)

	// Round up the width so we don't draw a partial bar.
	xColMax := math.Round(wf/d.binWidth) * d.binWidth

	xBin := 0
	xCol := (d.binWidth)/2 + (wf-xColMax)/2

	for _, chBins := range bins {
		for xBin < nbars && xBin >= 0 && xCol < xColMax {
			stop := calculateBar(chBins[xBin]*scale, hf)
			d.drawBar(xCol, hf, stop)

			xCol += d.binWidth
			xBin += delta
		}

		delta = -delta
		xBin += delta // ensure xBin is not out of bounds first.
	}
}

func (d *Display) drawBar(xCol, to, from float64) {
	d.drawer.MoveTo(xCol, from)
	d.drawer.LineTo(xCol, to)
	d.drawer.Stroke()
}

func calculateBar(value, height float64) float64 {
	bar := min(value, height)
	return height - bar
}

func max[T ~int | ~float64](i, j T) T {
	if i > j {
		return i
	}
	return j
}

func min[T ~int | ~float64](i, j T) T {
	if i < j {
		return i
	}
	return j
}
