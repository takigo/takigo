package radiobutton_test

import (
	"testing"

	"github.com/msorc/takigo/internal/testutil"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/radiobutton"
)

func TestRadiobuttonGroup(t *testing.T) {
	app := testutil.NewTestApp(t)
	choice := widget.NewVariable("b")
	calls := 0
	a := radiobutton.New(app, "a", radiobutton.Text("A"), radiobutton.Value("a"), radiobutton.Var(choice),
		radiobutton.Command(func() { calls++ }))
	b := radiobutton.New(app, "b", radiobutton.Text("B"), radiobutton.Value("b"), radiobutton.Var(choice))

	if a.Selected() || !b.Selected() {
		t.Fatalf("initial selection: a %v, b %v; want b", a.Selected(), b.Selected())
	}
	a.Invoke()
	if !a.Selected() || b.Selected() || choice.Get() != "a" || calls != 1 {
		t.Errorf("after a.Invoke: a %v, b %v, variable %q, calls %d", a.Selected(), b.Selected(), choice.Get(), calls)
	}
	b.Select()
	if a.Selected() || !b.Selected() || calls != 1 {
		t.Errorf("after b.Select: a %v, b %v, calls %d (Select runs no command)", a.Selected(), b.Selected(), calls)
	}
	choice.Set("neither")
	if a.Selected() || b.Selected() {
		t.Error("a button is selected for a value that matches neither")
	}
}

func TestRadiobuttonGeometry(t *testing.T) {
	app := testutil.NewTestApp(t)
	short := radiobutton.New(app, "s", radiobutton.Text("a"))
	long := radiobutton.New(app, "l", radiobutton.Text("a longer label"))
	if long.Win.ReqWidth <= short.Win.ReqWidth {
		t.Errorf("longer text is not wider: %d vs %d", long.Win.ReqWidth, short.Win.ReqWidth)
	}
	w := radiobutton.New(app, "w", radiobutton.Text("a"), radiobutton.Width(20))
	if w.Win.ReqWidth <= short.Win.ReqWidth {
		t.Errorf("Width(20) is not wider: %d vs %d", w.Win.ReqWidth, short.Win.ReqWidth)
	}
}
