// Demo: Unicode text display.
// Ported from Tk's unicodeout.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Unicode Text"), takigo.Size(450, 400))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("Unicode text samples from various scripts.\nAll rendered via Xft/fontconfig."),
		label.Anchor(option.AnchorW),
		label.PadX(10), label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))
	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"), button.Command(func() { app.Quit() }),
		button.PadX(10), button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	// Unicode samples.
	samples := []struct {
		lang string
		text string
	}{
		{"English", "The quick brown fox jumps over the lazy dog."},
		{"Russian", "Быстрая коричневая лиса перепрыгнула через ленивую собаку."},
		{"Greek", "Η γρήγορη καφετιά αλεπού πήδηξε πάνω από το τεμπέλικο σκυλί."},
		{"Japanese", "色は匂へど散りぬるを我が世誰ぞ常ならむ"},
		{"Chinese", "天地玄黃 宇宙洪荒 日月盈昃 辰宿列張"},
		{"Korean", "키스의 고유 조건은 입술끼리 만, 아니 , , 합니다."},
		{"Arabic", "صِف خَلقَ خَودِ كَمِثلِ الشَمسِ إِذ بَزَغَت"},
		{"Hebrew", "דג סקרן שט בים מאוכזב ולפתע מצא חברה"},
		{"Math", "∀x∈ℝ: ∑(i=1..n) xᵢ² ≥ 0, ∫₀∞ f(x)dx = π"},
		{"Emoji", "🎉 🎨 🖌️ 🎭 🎪 🎬 🎧 🎹 🎷 🎸"},
	}

	for _, s := range samples {
		l := label.New(root, "lang_"+s.lang, app,
			label.Text(fmt.Sprintf("%s: %s", s.lang, s.text)),
			label.Anchor(option.AnchorW),
			label.PadX(10), label.PadY(2),
		)
		pack.Pack(l.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
		_ = l
	}

	// Root events.
	app.Dispatcher().Bind(root.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			pack.ArrangeContainer(root)
		}
	})
	app.Dispatcher().Bind(root.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		d := root.Display.XDisplay
		d.SetForeground(root.GC, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), root.GC, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})
	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_Escape {
			app.Quit()
		}
	})

	_ = msg
	_ = dismissBtn
	app.MainLoop()
}
