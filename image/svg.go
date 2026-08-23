package image

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// NewPhotoFromSVGFile loads an SVG file by rasterizing it with an external
// converter (rsvg-convert or ImageMagick). The result is decoded as PNG and
// returned as a Photo.
//
// This requires at least one of the following commands on PATH:
//   - rsvg-convert (librsvg)
//   - magick       (ImageMagick 7)
//   - convert      (ImageMagick 6)
//
// Returns an error if no converter is available or the file cannot be parsed.
func NewPhotoFromSVGFile(name, path string) (*Photo, error) {
	data, err := rasterizeSVG(path)
	if err != nil {
		return nil, err
	}
	photo, err := NewPhotoFromReader(name, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("image: decode svg raster %s: %w", path, err)
	}
	return photo, nil
}

// rasterizeSVG converts the SVG at path into a PNG byte stream using the
// first available converter on PATH.
func rasterizeSVG(path string) ([]byte, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("image: open svg %s: %w", path, err)
	}

	if bin, ok := findOnPath("rsvg-convert"); ok {
		// rsvg-convert -f png -o <tmp> <path>; honour the SVG's intrinsic size
		// so vector art renders at its native resolution.
		tmp, err := os.CreateTemp("", "takigo-svg-*.png")
		if err != nil {
			return nil, fmt.Errorf("image: create temp for svg: %w", err)
		}
		tmpPath := tmp.Name()
		tmp.Close()
		defer os.Remove(tmpPath)

		cmd := exec.Command(bin, "-f", "png", "-o", tmpPath, path)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("rsvg-convert failed: %w: %s", err, stderr.String())
		}
		data, err := os.ReadFile(tmpPath)
		if err != nil {
			return nil, fmt.Errorf("read svg raster: %w", err)
		}
		return data, nil
	}

	for _, bin := range []string{"magick", "convert"} {
		if binPath, ok := findOnPath(bin); ok {
			// <bin> <input.svg> png:-   writes PNG to stdout
			args := []string{binPath, path, "png:-"}
			cmd := exec.Command(args[0], args[1:]...)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			data, err := cmd.Output()
			if err != nil {
				return nil, fmt.Errorf("%s failed: %w: %s", filepath.Base(binPath), err, stderr.String())
			}
			return data, nil
		}
	}

	return nil, fmt.Errorf("no SVG rasterizer found (install rsvg-convert or imagemagick)")
}

func findOnPath(name string) (string, bool) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", false
	}
	return path, true
}

// detectImageFormat returns the lowercase file extension (without the dot)
// if path has a recognised image extension, otherwise "".
func detectImageFormat(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if len(ext) > 1 && ext[0] == '.' {
		ext = ext[1:]
	}
	return ext
}
