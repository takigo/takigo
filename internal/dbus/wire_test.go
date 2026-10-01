package dbus

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func TestEncodeKnownBytes(t *testing.T) {
	tests := []struct {
		name string
		sig  string
		vals []any
		want []byte
	}{
		{"uint32", "u", []any{uint32(1)}, []byte{1, 0, 0, 0}},
		{"string", "s", []any{"hi"}, []byte{2, 0, 0, 0, 'h', 'i', 0}},
		{"byte then uint32 is padded", "yu", []any{byte(7), uint32(2)}, []byte{7, 0, 0, 0, 2, 0, 0, 0}},
		{"bool", "b", []any{true}, []byte{1, 0, 0, 0}},
		{"signature", "g", []any{"as"}, []byte{2, 'a', 's', 0}},
		{"variant", "v", []any{Variant{"u", uint32(9)}}, []byte{1, 'u', 0, 0, 9, 0, 0, 0}},
		{"array of strings", "as", []any{[]string{"a", "bc"}},
			[]byte{15, 0, 0, 0, 1, 0, 0, 0, 'a', 0, 0, 0, 2, 0, 0, 0, 'b', 'c', 0}},
		{"empty array still aligns its elements", "a(yv)", []any{[]any{}}, []byte{0, 0, 0, 0, 0, 0, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e encoder
			if err := e.values(tt.sig, tt.vals); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(e.buf, tt.want) {
				t.Errorf("encoded %v, want %v", e.buf, tt.want)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	sig := "susssasa{sv}i"
	vals := []any{
		"takigo", uint32(0), "icon", "Summary", "Body text",
		[]string{"default", "Open"},
		map[string]Variant{
			"urgency":  {"y", byte(1)},
			"category": {"s", "im.received"},
			"nested":   {"as", []string{"x", "y"}},
		},
		int32(-1),
	}
	var e encoder
	if err := e.values(sig, vals); err != nil {
		t.Fatal(err)
	}
	d := decoder{buf: e.buf, order: binary.LittleEndian}
	got, err := d.values(sig)
	if err != nil {
		t.Fatal(err)
	}
	if d.pos != len(e.buf) {
		t.Errorf("decoder stopped at %d of %d bytes", d.pos, len(e.buf))
	}
	want := []any{
		"takigo", uint32(0), "icon", "Summary", "Body text",
		[]any{"default", "Open"},
		// A dict decodes as its entries, in the sorted order they were written.
		[]any{
			[]any{"category", Variant{"s", "im.received"}},
			[]any{"nested", Variant{"as", []any{"x", "y"}}},
			[]any{"urgency", Variant{"y", byte(1)}},
		},
		int32(-1),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("decoded\n %#v\nwant\n %#v", got, want)
	}
}

func TestMalformedInput(t *testing.T) {
	var e encoder
	if err := e.values("u", []any{"not a number"}); !errors.Is(err, ErrBadMessage) {
		t.Errorf("type mismatch: %v, want ErrBadMessage", err)
	}
	if err := e.values("us", []any{uint32(1)}); !errors.Is(err, ErrBadMessage) {
		t.Errorf("missing value: %v, want ErrBadMessage", err)
	}
	for _, in := range [][]byte{
		{5, 0, 0, 0, 'a'}, // string shorter than its length
		{200, 0, 0, 0},    // array longer than the message
		{},                // nothing at all
	} {
		for _, sig := range []string{"s", "as", "u"} {
			d := decoder{buf: in, order: binary.LittleEndian}
			if _, err := d.values(sig); err == nil && len(in) < 4 {
				t.Errorf("decoding %v as %q succeeded", in, sig)
			}
		}
	}
	d := decoder{buf: []byte{3, 'u', 'u', 'u', 0, 0, 0, 0}, order: binary.LittleEndian}
	if _, err := d.value("v"); !errors.Is(err, ErrBadMessage) {
		t.Errorf("a variant holding three values: %v, want ErrBadMessage", err)
	}
	if _, _, err := next("(su"); !errors.Is(err, ErrBadMessage) {
		t.Errorf("unbalanced struct: %v", err)
	}
}

func FuzzDecode(f *testing.F) {
	f.Add([]byte{2, 0, 0, 0, 'h', 'i', 0}, "s")
	f.Add([]byte{15, 0, 0, 0, 1, 0, 0, 0, 'a', 0, 0, 0, 2, 0, 0, 0, 'b', 'c', 0}, "as")
	f.Add([]byte{1, 'u', 0, 0, 9, 0, 0, 0}, "v")
	f.Fuzz(func(t *testing.T, data []byte, sig string) {
		if len(sig) > 16 {
			return
		}
		d := decoder{buf: data, order: binary.LittleEndian}
		_, _ = d.values(sig) // must not panic or run away
	})
}

// With a session bus around, talk to the bus driver itself.
func TestSessionBusListNames(t *testing.T) {
	if os.Getenv("DBUS_SESSION_BUS_ADDRESS") == "" {
		t.Skip("no session bus (DBUS_SESSION_BUS_ADDRESS is unset)")
	}
	c, err := SessionBus()
	if err != nil {
		t.Skipf("cannot reach the session bus: %v", err)
	}
	defer c.Close()

	reply, err := c.Call("org.freedesktop.DBus", "/org/freedesktop/DBus", "org.freedesktop.DBus", "ListNames", "")
	if err != nil {
		t.Fatal(err)
	}
	names, ok := reply[0].([]any)
	if !ok || !slices.Contains(names, any("org.freedesktop.DBus")) {
		t.Errorf("ListNames = %v, want it to include org.freedesktop.DBus", reply)
	}

	_, err = c.Call("org.freedesktop.DBus", "/org/freedesktop/DBus", "org.freedesktop.DBus", "NoSuchMethod", "")
	if e, ok := errors.AsType[*Error](err); !ok || e.Name == "" {
		t.Errorf("calling a missing method: %v, want a *dbus.Error", err)
	}
	// The connection is still usable after an error reply.
	if _, err := c.Call("org.freedesktop.DBus", "/org/freedesktop/DBus", "org.freedesktop.DBus", "GetId", ""); err != nil {
		t.Errorf("call after an error reply: %v", err)
	}
}
