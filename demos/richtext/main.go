// Demo: an HTML help page shown in a text widget with the richtext
// package. There is no Tk counterpart: it shows the text widget's block
// boxes, list markers, rules, tab stops and image options.
package main

import (
	"fmt"
	"os"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/cursor"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/ttk"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/text"
	"github.com/takigo/takigo/widget/text/richtext"
)

const page = `
<style>
  .note { background: #ddf4ff; border-left: 4px solid #0969da; padding: 6px 12px; border-radius: 6px; }
  .warn { background: #fff8c5; border: 1px solid #d4a72c; padding: 6px 12px; border-radius: 6px; }
  .note, .warn { margin-bottom: 1em; }
  .muted { color: #59636e; }
  .lead { font-size: 1.15em; }
</style>

<h1 id="top">Rich text in takigo</h1>
<p class="lead">This page is plain <b>HTML</b> and <b>CSS</b>, turned into text widget
tags by the <code>richtext</code> package.</p>

<p><img src="earth.gif" alt="Earth" width="48" style="vertical-align: middle">
Contents: <a href="#inline">Inline styles</a> · <a href="#blocks">Blocks</a> ·
<a href="#floats">Floats</a> · <a href="#lists">Lists</a> · <a href="#tables">Tables</a> ·
<a href="https://www.tcl-lang.org/">an outside link</a></p>

<h2 id="inline">Inline styles</h2>
<p>Text can be <b>bold</b>, <i>italic</i>, <u>underlined</u>, <s>struck out</s>,
<mark>highlighted</mark>, <span style="color: #cf222e">coloured</span>,
<small>small</small> or <big>big</big>. Inline <code>code</code> and keys such as
<kbd>Ctrl</kbd>+<kbd>C</kbd> use the fixed font. Formulas work too:
E = mc<sup>2</sup>, H<sub>2</sub>O.</p>

<h2 id="blocks">Blocks</h2>
<blockquote>A quotation has a bar in its left margin. Margins between blocks
collapse as they do in a browser.</blockquote>

<div class="note"><b>Note.</b> A block with a background, a border or padding
becomes a <i>block box</i>, drawn behind its lines with rounded corners.</div>

<div class="warn"><b>Careful.</b> Boxes nest: this one holds a code block.
<pre>richtext.Render(tw, page)</pre>
and goes on after it.</div>

<pre>func main() {
    app, _ := takigo.NewApp()
    tw := text.New(app, "t")
    richtext.Render(tw, "&lt;h1&gt;Hello&lt;/h1&gt;")
    app.Run()
}</pre>

<p style="text-align: center" class="muted">This paragraph is centred.</p>
<hr>

<h2 id="floats">Floating images</h2>
<p><img src="earth.gif" alt="Earth" style="float: left; width: 128px; margin: 0 12px 6px 0">
An image with <code>float: left</code> or <code>float: right</code> leaves the line and
sits against the margin.</p>
<p>The lines beside it are narrowed until its bottom, even when they belong to the next
paragraph, as in a browser. Once the image ends, the text goes on below it across the full
width of the page again.</p>

<h2 id="lists">Lists</h2>
<ul>
  <li>Bullets hang in the margin and are not part of the text.</li>
  <li>Lists nest:
    <ul><li>second level</li><li>with its own marker
      <ul><li>and a third</li></ul></li></ul>
  </li>
</ul>
<ol start="7">
  <li>Ordered lists count from <code>start</code>,</li>
  <li>in any <code>list-style-type</code>:
    <ol type="i"><li>roman,</li><li>alphabetic or decimal.</li></ol></li>
</ol>
<dl><dt>Definition lists</dt><dd>indent their descriptions.</dd></dl>

<h2 id="tables">Tables</h2>
<table>
  <tr><th>Widget</th><th>Tk source</th><th>Notes</th></tr>
  <tr><td>text</td><td><code>tkText.c</code></td><td>Tags, marks, embedded windows and images,
    undo, peers, and now block boxes and floats.</td></tr>
  <tr><td>canvas</td><td><code>tkCanvas.c</code></td><td>Items with tags, anti-aliased outside a
    Classic App.</td></tr>
  <tr><td rowspan="2">entry</td><td><code>tkEntry.c</code></td><td>One line.</td></tr>
  <tr><td colspan="2"><img src="earth.gif" style="width: 16px; vertical-align: middle"> A cell can
    span columns or rows, and hold images.</td></tr>
</table>
<p class="muted">Columns are as wide as their content allows; long cells wrap.</p>
<p><img src="Tk_feather.png" alt="Tk feather" style="max-width: 120px"></p>
<p><a href="#top">Back to the top</a></p>
`

func main() {
	app, err := takigo.NewApp(takigo.Title("Rich Text Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("richtext"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	status := ttk.NewLabel(f, "status", ttk.LabelText("Click a link."))
	pack.Pack(status, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	yscroll := ttk.NewScrollbar(f, "scroll",
		ttk.ScrollbarOrient(ttk.Vertical),
		ttk.ScrollbarCommand(func(widget.ScrollRequest) {}),
	)
	tw := text.New(f, "text",
		text.Width(72),
		text.Height(36),
		text.PadXOpt(12),
		text.PadYOpt(8),
		text.ReadOnly(true),
		text.Cursor(cursor.LeftPtr),
		text.YScrollCommand(func(first, last float64) { yscroll.Set(first, last) }),
	)
	yscroll.Command = widget.ScrollY(tw)
	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(tw, pack.Expand(true), pack.FillOpt(pack.FillBoth))

	err = richtext.Render(tw, page,
		richtext.BaseDir(demohelper.DemoDir("images")),
		richtext.OnLink(func(href string) {
			_ = status.Configure(ttk.LabelText("Link: " + href))
		}),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	app.Run()
}
