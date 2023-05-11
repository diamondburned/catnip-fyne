package vecd

import (
	"image"

	"github.com/llgcode/draw2d/draw2dimg"
)

// Context is a vector drawing context.
type Context struct {
	*draw2dimg.GraphicContext
	buffer *image.RGBA
}

// NewContext creates a new vector drawing context.
func NewContext() *Context {
	buffer := image.NewRGBA(image.Rect(0, 0, 0, 0))
	gc := draw2dimg.NewGraphicContext(buffer)
	return &Context{gc, buffer}
}

// Resize resizes the drawing context.
func (c *Context) Resize(w, h int) {
	curr := c.buffer.Rect.Size()
	if curr.X == w && curr.Y == h {
		return
	}

	// This is actually not that slow! Maybe up to 4Kx4K it's fine.
	// It also gives Fyne the opportunity to blit the buffer directly.
	c.buffer = image.NewRGBA(image.Rect(0, 0, w, h))
	c.GraphicContext = draw2dimg.NewGraphicContext(c.buffer)
}

// Size returns the size of the drawing context.
func (c *Context) Size() image.Point {
	return c.buffer.Rect.Size()
}

// Image returns the image of the drawing context.
func (c *Context) Image() *image.RGBA {
	return c.buffer
}

// Clear clears the drawing context. The canvas after this call will be
// completely transparent.
func (c *Context) Clear() {
	// go optimizes this to a memclr
	for i := range c.buffer.Pix {
		c.buffer.Pix[i] = 0
	}
}
