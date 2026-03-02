// Demo: Unicode text display.
// Ported from Tk's unicodeout.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	d := demohelper.Setup("Unicode Text", 450, 400,
		"Unicode text samples from various scripts.\nAll rendered via Xft/fontconfig.")
	app := d.App

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
		l := label.New(app, "lang_"+s.lang,
			label.Text(fmt.Sprintf("%s: %s", s.lang, s.text)),
			label.Anchor(option.AnchorW),
			label.PadX(10), label.PadY(2),
		)
		pack.Pack(l, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
		_ = l
	}

	d.Run()
}
