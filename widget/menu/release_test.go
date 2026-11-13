package menu_test

import (
	"testing"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/widget/menu"
)

// pointer sends button and motion events to a posted menu at a point
// inside it, in menu coordinates.
type pointer struct {
	t      *testing.T
	m      *menu.Menu
	ox, oy int // root position of the menu
	x, y   int
}

func (p *pointer) send(typ event.Type, button uint) {
	p.t.Helper()
	p.m.App.Dispatcher().Dispatch(&event.Event{
		Type: typ, Window: p.m.Window().PlatformID, Button: button,
		X: p.x, Y: p.y, RootX: p.ox + p.x, RootY: p.oy + p.y,
	})
}

func (p *pointer) click(button uint) {
	p.send(event.ButtonPressType, button)
	p.send(event.ButtonReleaseType, button)
}

// postAt posts m at (ox, oy) with the pointer over the middle of its
// entry i, the way a context menu appears under a stationary pointer.
func postAt(t *testing.T, m *menu.Menu, ox, oy, i int, post func(x, y int)) *pointer {
	t.Helper()
	post(ox, oy)
	t.Cleanup(m.Unpost)
	n := len(m.Entries())
	h := m.Win.Height / n
	p := &pointer{t: t, m: m, ox: ox, oy: oy, x: m.Win.Width / 2, y: i*h + h/2}
	p.send(event.EnterType, 0)
	return p
}

// TestMenuReleaseOfPostingPress checks that the release of the press that
// posted a menu does not invoke the entry it lands on, even after Windows'
// motion event at the unchanged position, while motion makes it count.
func TestMenuReleaseOfPostingPress(t *testing.T) {
	app := testutil.NewTestApp(t)
	for _, tc := range []struct {
		name string
		post func(m *menu.Menu) func(x, y int)
	}{
		{"Post", func(m *menu.Menu) func(x, y int) { return m.Post }},
		{"PostFromButton", func(m *menu.Menu) func(x, y int) { return m.PostFromButton }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := menu.New(app, "post"+tc.name, menu.TearOffOpt(false))
			invoked := 0
			m.AddCommand("Open", func() { invoked++ })
			p := postAt(t, m, 40, 40, 0, tc.post(m))

			p.send(event.MotionType, 0) // Windows: motion where the popup appeared
			p.send(event.ButtonReleaseType, 1)
			if invoked != 0 || !m.IsPosted() {
				t.Fatalf("release of the posting press: invoked=%d posted=%v, want 0 and posted", invoked, m.IsPosted())
			}
			p.y++
			p.send(event.MotionType, 0)
			p.send(event.ButtonReleaseType, 1)
			if invoked != 1 || m.IsPosted() {
				t.Errorf("release after a drag: invoked=%d posted=%v, want 1 and unposted", invoked, m.IsPosted())
			}
		})
	}
}

// TestMenuClickWithoutMotion checks that a click on a menu posted under a
// stationary pointer invokes the entry under it, as in Tk, where a press
// on the menu (tk::MenuButtonDown) starts the press whose release
// tk::MenuInvoke handles.
func TestMenuClickWithoutMotion(t *testing.T) {
	app := testutil.NewTestApp(t)

	t.Run("popup", func(t *testing.T) {
		m := menu.New(app, "popup", menu.TearOffOpt(false))
		invoked := 0
		m.AddCommand("Open", func() { invoked++ })
		m.AddCommand("Save", func() {})
		p := postAt(t, m, 40, 40, 0, m.Post)

		p.send(event.ButtonReleaseType, 3) // the right-click that posted it
		p.send(event.MotionType, 0)        // Windows: motion where the popup appeared
		p.click(1)
		if invoked != 1 || m.IsPosted() {
			t.Errorf("click without motion: invoked=%d posted=%v, want 1 and unposted", invoked, m.IsPosted())
		}
	})

	t.Run("menubutton", func(t *testing.T) {
		m := menu.New(app, "mb", menu.TearOffOpt(false))
		invoked := 0
		m.AddCommand("Open", func() { invoked++ })
		p := postAt(t, m, 40, 40, 0, m.PostFromButton)

		p.send(event.ButtonReleaseType, 1) // the press on the menubutton
		p.click(1)
		if invoked != 1 || m.IsPosted() {
			t.Errorf("click without motion: invoked=%d posted=%v, want 1 and unposted", invoked, m.IsPosted())
		}
	})

	t.Run("cascade", func(t *testing.T) {
		m := menu.New(app, "top", menu.TearOffOpt(false))
		sub := menu.New(m, "sub", menu.TearOffOpt(false))
		invoked := 0
		sub.AddCommand("Inner", func() { invoked++ })
		m.AddCascade("More", sub)
		p := postAt(t, m, 40, 40, 0, m.Post)

		p.click(1)
		if !sub.IsPosted() {
			t.Fatal("clicking the cascade entry did not post its menu")
		}
		sp := &pointer{t: t, m: sub, ox: 40 + m.Win.Width, oy: 40, x: sub.Win.Width / 2, y: sub.Win.Height / 2}
		sp.send(event.EnterType, 0)
		sp.click(1)
		if invoked != 1 || sub.IsPosted() || m.IsPosted() {
			t.Errorf("click in the cascade: invoked=%d posted=%v/%v, want 1 and both unposted",
				invoked, m.IsPosted(), sub.IsPosted())
		}
	})
}
