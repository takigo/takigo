package takigo_test

import (
	"context"
	"errors"
	"testing"
	"time"

	takigo "github.com/takigo/takigo"

	"github.com/takigo/takigo/dialog"
	"github.com/takigo/takigo/internal/testutil"
)

func TestRunContextQuitsWhenDone(t *testing.T) {
	testutil.RequireDisplay(t)
	app, err := takigo.NewApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := app.RunContext(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("RunContext = %v, want context.DeadlineExceeded", err)
	}
}

func TestRunContextReturnsNilOnQuit(t *testing.T) {
	testutil.RequireDisplay(t)
	app, err := takigo.NewApp()
	if err != nil {
		t.Fatal(err)
	}
	app.RunOnMain(app.Quit)
	if err := app.RunContext(context.Background()); err != nil {
		t.Errorf("RunContext = %v, want nil", err)
	}
}

func TestDialogClosesWhenContextDone(t *testing.T) {
	app := testutil.NewTestApp(t)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result := dialog.ResultOK
	app.RunOnMain(func() {
		result = dialog.ShowMessage(dialog.WithContext(ctx, app), dialog.MsgMessage("wait"))
		app.Quit()
	})
	watchdog := time.AfterFunc(5*time.Second, app.Quit)
	defer watchdog.Stop()
	app.MainLoop()

	if result != dialog.ResultNone {
		t.Errorf("result = %v, want ResultNone", result)
	}
	if ctx.Err() == nil {
		t.Error("the dialog returned before the context was done")
	}
	for _, c := range app.Root().Children {
		if c.Name == "dialog" {
			t.Error("the dialog window is still there")
		}
	}
}
