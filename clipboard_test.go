package takigo_test

import (
	"strings"
	"testing"
	"time"

	takigo "github.com/msorc/takigo"
	"github.com/msorc/takigo/internal/testutil"
	"github.com/msorc/takigo/platform"
)

// A clipboard far over Tk's 4000-byte limit goes through the INCR
// protocol between two X clients (two Apps, two display connections).
func TestClipboardLargeTransferBetweenApps(t *testing.T) {
	testutil.RequireDisplay(t)
	text := strings.Repeat("naïve café — 日本語のテキスト, line of text\n", 2500)

	owner, err := takigo.NewApp(takigo.Title("owner"), takigo.Size(50, 50))
	if err != nil {
		t.Fatal(err)
	}
	reader, err := takigo.NewApp(takigo.Title("reader"), takigo.Size(50, 50))
	if err != nil {
		t.Fatal(err)
	}
	owner.Clipboard().Set(owner.Root().PlatformID, text, platform.CurrentTime)
	owner.Server().Sync(false) // the X server knows the owner before the reader asks

	got := make(chan string, 1)
	reader.DoWhenIdle(func() {
		reader.Clipboard().Get(reader.Root().PlatformID, platform.CurrentTime, func(s string) { got <- s })
	})
	ownerDone, readerDone := make(chan struct{}), make(chan struct{})
	go func() { owner.Run(); close(ownerDone) }()
	go func() { reader.Run(); close(readerDone) }()
	defer func() {
		owner.Quit()
		reader.Quit()
		<-ownerDone
		<-readerDone
	}()

	select {
	case s := <-got:
		if s != text {
			t.Errorf("pasted %d bytes, want the %d-byte clipboard (prefix %q)", len(s), len(text), s[:min(len(s), 40)])
		}
	case <-time.After(15 * time.Second):
		t.Fatal("clipboard transfer did not finish")
	}
}
