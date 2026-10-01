package checkbutton_test

import (
	"testing"

	"github.com/msorc/takigo/internal/testutil"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/checkbutton"
)

func TestCheckbuttonVariable(t *testing.T) {
	app := testutil.NewTestApp(t)
	v := widget.NewVariable("no")
	calls := 0
	c := checkbutton.New(app, "c", checkbutton.Text("opt"),
		checkbutton.OnValueOpt("yes"), checkbutton.OffValueOpt("no"),
		checkbutton.Var(v), checkbutton.Command(func() { calls++ }))

	if c.Selected() {
		t.Fatal("selected with the off value")
	}
	c.Invoke()
	if !c.Selected() || v.Get() != "yes" || calls != 1 {
		t.Errorf("after Invoke: selected %v, variable %q, command calls %d", c.Selected(), v.Get(), calls)
	}
	c.Toggle()
	if c.Selected() || v.Get() != "no" {
		t.Errorf("after Toggle: selected %v, variable %q", c.Selected(), v.Get())
	}
	v.Set("yes")
	if !c.Selected() {
		t.Error("setting the variable to the on value did not select")
	}
	v.Set("something else")
	if c.Selected() {
		t.Error("selected for a value that is neither on nor off")
	}
}

func TestCheckbuttonDefaultVariableAndState(t *testing.T) {
	app := testutil.NewTestApp(t)
	c := checkbutton.New(app, "c", checkbutton.Text("opt"))
	if c.Selected() {
		t.Error("a new checkbutton is selected")
	}
	c.Invoke()
	if !c.Selected() {
		t.Error("Invoke did not select a checkbutton with its own variable")
	}

	calls := 0
	d := checkbutton.New(app, "d", checkbutton.Text("off"),
		checkbutton.State(widget.StateDisabled), checkbutton.Command(func() { calls++ }))
	d.Invoke()
	if d.Selected() || calls != 0 {
		t.Errorf("a disabled checkbutton was invoked: selected %v, calls %d", d.Selected(), calls)
	}
}

func TestCheckbuttonGeometry(t *testing.T) {
	app := testutil.NewTestApp(t)
	short := checkbutton.New(app, "s", checkbutton.Text("a"))
	long := checkbutton.New(app, "l", checkbutton.Text("a longer label"))
	if long.Win.ReqWidth <= short.Win.ReqWidth {
		t.Errorf("longer text is not wider: %d vs %d", long.Win.ReqWidth, short.Win.ReqWidth)
	}
	// The indicator takes room: without it the widget is narrower.
	plain := checkbutton.New(app, "p", checkbutton.Text("a"), checkbutton.IndicatorOnOpt(false))
	if plain.Win.ReqWidth >= short.Win.ReqWidth {
		t.Errorf("no indicator is not narrower: %d vs %d", plain.Win.ReqWidth, short.Win.ReqWidth)
	}
}
