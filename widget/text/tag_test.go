package text

import (
	"slices"
	"testing"
)

func docWithText(s string) *Document {
	doc := NewDocument()
	doc.Insert(Index{1, 0}, s)
	return doc
}

func rangesOf(doc *Document, tag string) [][2]Index {
	var out [][2]Index
	for _, tr := range doc.TagRangesFor(tag) {
		out = append(out, [2]Index{tr.Start, tr.End})
	}
	return out
}

func TestTagAddAndRangesFor(t *testing.T) {
	doc := docWithText("hello world\nsecond line")
	doc.TagAdd("b", Index{1, 0}, Index{1, 5})
	got := rangesOf(doc, "b")
	want := [][2]Index{{{1, 0}, {1, 5}}}
	if !slices.Equal(got, want) {
		t.Fatalf("ranges = %v, want %v", got, want)
	}
	if _, ok := doc.Tags["b"]; !ok {
		t.Fatal("TagAdd did not create the tag")
	}
	if len(doc.TagRangesFor("missing")) != 0 {
		t.Fatal("unknown tag has ranges")
	}
}

func TestTagAddEmptyRangeIgnored(t *testing.T) {
	doc := docWithText("abc")
	doc.TagAdd("b", Index{1, 2}, Index{1, 2})
	doc.TagAdd("b", Index{1, 3}, Index{1, 1})
	if n := len(doc.TagRangesFor("b")); n != 0 {
		t.Fatalf("got %d ranges for empty/reversed add", n)
	}
}

func TestTagsAt(t *testing.T) {
	doc := docWithText("hello world")
	doc.TagAdd("b", Index{1, 2}, Index{1, 6})
	tests := []struct {
		idx  Index
		want []string
	}{
		{Index{1, 1}, nil},
		{Index{1, 2}, []string{"b"}},
		{Index{1, 5}, []string{"b"}},
		{Index{1, 6}, nil},
	}
	for _, tc := range tests {
		var names []string
		for _, tg := range doc.TagsAt(tc.idx) {
			names = append(names, tg.Name)
		}
		if !slices.Equal(names, tc.want) {
			t.Errorf("TagsAt(%v) = %v, want %v", tc.idx, names, tc.want)
		}
	}
}

func TestTagsAtPriorityOrder(t *testing.T) {
	doc := docWithText("hello world")
	doc.Tags["sel"] = &Tag{Name: "sel", Priority: 1000}
	doc.TagAdd("sel", Index{1, 0}, Index{1, 11})
	doc.TagAdd("a", Index{1, 0}, Index{1, 11})
	var names []string
	for _, tg := range doc.TagsAt(Index{1, 3}) {
		names = append(names, tg.Name)
	}
	if want := []string{"a", "sel"}; !slices.Equal(names, want) {
		t.Fatalf("TagsAt order = %v, want %v", names, want)
	}
}

func TestTagRemoveWhole(t *testing.T) {
	doc := docWithText("hello world")
	doc.TagAdd("b", Index{1, 2}, Index{1, 6})
	doc.TagRemove("b", Index{1, 0}, Index{1, 11})
	if n := len(doc.TagRangesFor("b")); n != 0 {
		t.Fatalf("got %d ranges after removing whole range", n)
	}
}

func TestTagRemoveUntouchedNeighbour(t *testing.T) {
	doc := docWithText("hello world")
	doc.TagAdd("b", Index{1, 0}, Index{1, 3})
	doc.TagAdd("b", Index{1, 7}, Index{1, 11})
	doc.TagRemove("b", Index{1, 3}, Index{1, 7})
	got := rangesOf(doc, "b")
	want := [][2]Index{{{1, 0}, {1, 3}}, {{1, 7}, {1, 11}}}
	if !slices.Equal(got, want) {
		t.Fatalf("ranges = %v, want %v", got, want)
	}
}

func TestTagAdjustOnInsert(t *testing.T) {
	tests := []struct {
		name   string
		at     Index
		text   string
		toEnd  bool
		want   [][2]Index
		before [2]Index
	}{
		{"before start", Index{1, 1}, "XY", false, [][2]Index{{{1, 6}, {1, 10}}}, [2]Index{{1, 4}, {1, 8}}},
		{"at start", Index{1, 4}, "XY", false, [][2]Index{{{1, 6}, {1, 10}}}, [2]Index{{1, 4}, {1, 8}}},
		{"inside", Index{1, 5}, "XY", false, [][2]Index{{{1, 4}, {1, 10}}}, [2]Index{{1, 4}, {1, 8}}},
		{"at end", Index{1, 8}, "XY", false, [][2]Index{{{1, 4}, {1, 8}}}, [2]Index{{1, 4}, {1, 8}}},
		{"after end", Index{1, 9}, "XY", false, [][2]Index{{{1, 4}, {1, 8}}}, [2]Index{{1, 4}, {1, 8}}},
		{"newline inside", Index{1, 6}, "\n", false, [][2]Index{{{1, 4}, {2, 2}}}, [2]Index{{1, 4}, {1, 8}}},
		{"newline before", Index{1, 0}, "\n", false, [][2]Index{{{2, 4}, {2, 8}}}, [2]Index{{1, 4}, {1, 8}}},
		{"later line", Index{2, 0}, "XY\n", false, [][2]Index{{{1, 4}, {1, 8}}}, [2]Index{{1, 4}, {1, 8}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := docWithText("0123456789\nabc")
			doc.TagAdd("b", tc.before[0], tc.before[1])
			doc.Insert(tc.at, tc.text)
			if got := rangesOf(doc, "b"); !slices.Equal(got, tc.want) {
				t.Fatalf("ranges = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTagAdjustOnInsertToEnd(t *testing.T) {
	doc := docWithText("0123")
	tw := &TextWidget{doc: doc}
	tw.TagAdd("b", "1.0", "end")
	doc.Insert(Index{1, 4}, "XY")
	got := rangesOf(doc, "b")
	want := [][2]Index{{{1, 0}, {1, 6}}}
	if !slices.Equal(got, want) {
		t.Fatalf("ranges = %v, want %v", got, want)
	}
}

func TestTagAdjustOnDelete(t *testing.T) {
	tests := []struct {
		name       string
		start, end Index
		want       [][2]Index
	}{
		{"before", Index{1, 0}, Index{1, 2}, [][2]Index{{{1, 2}, {1, 6}}}},
		{"inside", Index{1, 5}, Index{1, 6}, [][2]Index{{{1, 4}, {1, 7}}}},
		{"spanning", Index{1, 3}, Index{1, 9}, nil},
		{"overlap start", Index{1, 2}, Index{1, 6}, [][2]Index{{{1, 2}, {1, 4}}}},
		{"overlap end", Index{1, 6}, Index{1, 9}, [][2]Index{{{1, 4}, {1, 6}}}},
		{"after", Index{1, 9}, Index{1, 10}, [][2]Index{{{1, 4}, {1, 8}}}},
		{"cross line before", Index{1, 1}, Index{2, 1}, [][2]Index{{{1, 2}, {1, 4}}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := docWithText("0123456789\nabcdefgh")
			if tc.name == "cross line before" {
				doc.TagAdd("b", Index{2, 2}, Index{2, 4})
			} else {
				doc.TagAdd("b", Index{1, 4}, Index{1, 8})
			}
			doc.Delete(tc.start, tc.end)
			if got := rangesOf(doc, "b"); !slices.Equal(got, tc.want) {
				t.Fatalf("ranges = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTagConfigureCreatesTag(t *testing.T) {
	doc := NewDocument()
	doc.TagConfigure("x", nil, nil, TagUnderline(true))
	tg, ok := doc.Tags["x"]
	if !ok || !tg.Underline {
		t.Fatal("TagConfigure did not create/configure the tag")
	}
}
