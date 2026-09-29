package canvas

import (
	"fmt"
	"io"
	"os"
)

// Postscript generates a PostScript representation of the canvas and either
// returns it as a string or writes it to a file/Writer depending on the
// options.
//
// Mirrors tk/generic/tkCanvPs.c:TkCanvPostscriptObjCmd.
//   - If -file is set: writes to that file and returns "".
//   - If -writer is set: writes to the writer and returns "".
//   - Otherwise: returns the full PostScript string.
func (c *Canvas) Postscript(opts ...PostscriptOption) (string, error) {
	c.compact()
	cfg := psConfig{
		Prolog:    true,
		ColorMode: "color",
	}
	for _, o := range opts {
		o(&cfg)
	}

	// Default region to canvas pixel area.
	x, y, w, h := cfg.X, cfg.Y, cfg.Width, cfg.Height
	if w == 0 || h == 0 {
		x = c.xOrigin
		y = c.yOrigin
		w = c.Win.Width
		h = c.Win.Height
		if w <= 0 || h <= 0 {
			w = 400
			h = 300
		}
	}

	ps := NewPSContext()
	ps.X, ps.Y = x, y
	ps.Width, ps.Height = w, h
	ps.X2, ps.Y2 = x+w, y+h
	ps.ColorLevel = cfg.colorLevel()
	ps.Colormap = cfg.ColorMap
	ps.Fontmap = cfg.FontMap
	ps.Rotate = cfg.Rotate
	ps.Prolog = cfg.Prolog
	if cfg.Anchor != 0 {
		ps.Anchor = cfg.Anchor
	}
	if cfg.PageX != 0 {
		ps.PageX = cfg.PageX
	}
	if cfg.PageY != 0 {
		ps.PageY = cfg.PageY
	}
	if cfg.PageW != 0 {
		ps.PageWidth = cfg.PageW
	}
	if cfg.PageH != 0 {
		ps.PageHeight = cfg.PageH
	}
	if cfg.Title != "" {
		ps.Title = cfg.Title
	} else {
		ps.Title = c.Win.Name
		if ps.Title == "" {
			ps.Title = "takigo canvas"
		}
	}

	// Page scale: default 1.0 (each canvas pixel = 1 PostScript point).
	ps.Scale = 1.0

	// ---- Pre-pass: collect font names ----
	ps.Prepass = true
	for _, entry := range c.items {
		item := entry.item
		if item.State() == ItemStateHidden {
			continue
		}
		bx1, by1, bx2, by2 := item.BBox()
		if bx1 >= ps.X2 || bx2 < ps.X || by1 >= ps.Y2 || by2 < ps.Y {
			continue
		}
		_ = item.Postscript(ps)
	}
	ps.Prepass = false

	// ---- Emit header ----
	ps.EmitHeader()

	// ---- Real pass: emit each item's PostScript ----
	for _, entry := range c.items {
		item := entry.item
		if item.State() == ItemStateHidden {
			continue
		}
		bx1, by1, bx2, by2 := item.BBox()
		if bx1 >= ps.X2 || bx2 < ps.X || by1 >= ps.Y2 || by2 < ps.Y {
			continue
		}
		ps.ResetItemBuf()
		if err := item.Postscript(ps); err != nil {
			return "", err
		}
		body := ps.TakeItemBuf()
		if body == "" {
			continue
		}
		// Wrap each item in gsave/grestore so its graphics state changes
		// don't leak to the next item.
		ps.write("gsave\n")
		ps.write(body)
		ps.write("grestore\n")
	}

	ps.EmitTrailer()

	out := ps.String()

	switch {
	case cfg.File != "":
		if err := os.WriteFile(cfg.File, []byte(out), 0o644); err != nil {
			return "", fmt.Errorf("canvas postscript: write %s: %w", cfg.File, err)
		}
		return "", nil
	case cfg.Channel != nil:
		if _, err := io.WriteString(cfg.Channel, out); err != nil {
			return "", fmt.Errorf("canvas postscript: write to channel: %w", err)
		}
		return "", nil
	default:
		return out, nil
	}
}
