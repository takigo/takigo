package square_test

import (
	"testing"

	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/widget/square"
)

func TestSquareConfigure(t *testing.T) {
	app := testutil.NewTestApp(t)
	s := square.New(app, "s", square.SizeOpt(20), square.PosXOpt(5), square.PosYOpt(7))
	if err := s.Configure(square.SizeOpt(40), square.Background("no-such-colour")); err == nil {
		t.Error("Configure with a bad colour returned no error")
	}
	s.SetPosition(10, 12)
	s.Destroy()
	s.Destroy()
}
