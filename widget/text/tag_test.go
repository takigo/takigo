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
	doc.putTag(&Tag{Name: "sel", Priority: selPriority})
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

func TestTagAddMerges(t *testing.T) {
	tests := []struct {
		name string
		adds [][2]Index
		want [][2]Index
	}{
		{"overlapping", [][2]Index{{{1, 0}, {1, 5}}, {{1, 3}, {1, 8}}}, [][2]Index{{{1, 0}, {1, 8}}}},
		{"adjacent", [][2]Index{{{1, 0}, {1, 5}}, {{1, 5}, {1, 8}}}, [][2]Index{{{1, 0}, {1, 8}}}},
		{"contained", [][2]Index{{{1, 0}, {1, 8}}, {{1, 3}, {1, 5}}}, [][2]Index{{{1, 0}, {1, 8}}}},
		{"same twice", [][2]Index{{{1, 2}, {1, 5}}, {{1, 2}, {1, 5}}}, [][2]Index{{{1, 2}, {1, 5}}}},
		{"bridging", [][2]Index{{{1, 0}, {1, 2}}, {{1, 6}, {1, 8}}, {{1, 2}, {1, 6}}}, [][2]Index{{{1, 0}, {1, 8}}}},
		{"disjoint sorted", [][2]Index{{{1, 6}, {1, 8}}, {{1, 0}, {1, 2}}}, [][2]Index{{{1, 0}, {1, 2}}, {{1, 6}, {1, 8}}}},
		{"across lines", [][2]Index{{{1, 8}, {2, 2}}, {{2, 1}, {2, 5}}}, [][2]Index{{{1, 8}, {2, 5}}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := docWithText("0123456789\nabcdefgh")
			for _, r := range tc.adds {
				doc.TagAdd("b", r[0], r[1])
			}
			if got := rangesOf(doc, "b"); !slices.Equal(got, tc.want) {
				t.Fatalf("ranges = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTagRemoveTrimsAndSplits(t *testing.T) {
	tests := []struct {
		name       string
		start, end Index
		want       [][2]Index
	}{
		{"trim left", Index{1, 0}, Index{1, 4}, [][2]Index{{{1, 4}, {1, 8}}}},
		{"trim right", Index{1, 6}, Index{1, 10}, [][2]Index{{{1, 2}, {1, 6}}}},
		{"split middle", Index{1, 4}, Index{1, 5}, [][2]Index{{{1, 2}, {1, 4}}, {{1, 5}, {1, 8}}}},
		{"outside before", Index{1, 0}, Index{1, 2}, [][2]Index{{{1, 2}, {1, 8}}}},
		{"outside after", Index{1, 8}, Index{1, 10}, [][2]Index{{{1, 2}, {1, 8}}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc := docWithText("0123456789")
			doc.TagAdd("b", Index{1, 2}, Index{1, 8})
			doc.TagRemove("b", tc.start, tc.end)
			if got := rangesOf(doc, "b"); !slices.Equal(got, tc.want) {
				t.Fatalf("ranges = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTagRemoveSpanningSeveral(t *testing.T) {
	doc := docWithText("0123456789")
	doc.TagAdd("b", Index{1, 0}, Index{1, 2})
	doc.TagAdd("b", Index{1, 4}, Index{1, 6})
	doc.TagAdd("b", Index{1, 8}, Index{1, 10})
	doc.TagRemove("b", Index{1, 1}, Index{1, 9})
	got := rangesOf(doc, "b")
	want := [][2]Index{{{1, 0}, {1, 1}}, {{1, 9}, {1, 10}}}
	if !slices.Equal(got, want) {
		t.Fatalf("ranges = %v, want %v", got, want)
	}
}

func TestTagToEndSurvivesRightRemainder(t *testing.T) {
	doc := docWithText("0123456789")
	tw := &TextWidget{doc: doc}
	tw.TagAdd("b", "1.0", "end")
	doc.TagRemove("b", Index{1, 2}, Index{1, 4})
	doc.Insert(doc.EndIndex(), "XY")
	got := rangesOf(doc, "b")
	want := [][2]Index{{{1, 0}, {1, 2}}, {{1, 4}, {1, 12}}}
	if !slices.Equal(got, want) {
		t.Fatalf("ranges = %v, want %v", got, want)
	}
	doc.Insert(Index{1, 2}, "Z")
	if got := rangesOf(doc, "b"); got[0][1] != (Index{1, 2}) {
		t.Fatalf("left piece grew: %v", got)
	}
}

func TestTagDeleteMergesAdjacent(t *testing.T) {
	doc := docWithText("0123456789")
	doc.TagAdd("b", Index{1, 0}, Index{1, 3})
	doc.TagAdd("b", Index{1, 6}, Index{1, 10})
	doc.Delete(Index{1, 3}, Index{1, 6})
	got := rangesOf(doc, "b")
	want := [][2]Index{{{1, 0}, {1, 7}}}
	if !slices.Equal(got, want) {
		t.Fatalf("ranges = %v, want %v", got, want)
	}
}

func TestTagDeleteNewlineShiftsLaterRanges(t *testing.T) {
	doc := docWithText("0123\nabcd\nwxyz")
	doc.TagAdd("b", Index{3, 1}, Index{3, 3})
	doc.Delete(Index{1, 4}, Index{2, 0})
	got := rangesOf(doc, "b")
	want := [][2]Index{{{2, 1}, {2, 3}}}
	if !slices.Equal(got, want) {
		t.Fatalf("ranges = %v, want %v", got, want)
	}
}

func TestTagPriorityIsCreationOrder(t *testing.T) {
	doc := docWithText("hello world")
	doc.putTag(&Tag{Name: "sel", Priority: selPriority})
	doc.TagConfigure("late", nil, nil)
	doc.TagAdd("early", Index{1, 0}, Index{1, 11})
	doc.TagAdd("late", Index{1, 0}, Index{1, 11})
	doc.TagAdd("sel", Index{1, 0}, Index{1, 11})
	doc.TagAdd("latest", Index{1, 0}, Index{1, 11})
	var names []string
	for _, tg := range doc.TagsAt(Index{1, 3}) {
		names = append(names, tg.Name)
	}
	want := []string{"late", "early", "latest", "sel"}
	if !slices.Equal(names, want) {
		t.Fatalf("TagsAt order = %v, want %v", names, want)
	}
}

func TestTagsOnLine(t *testing.T) {
	doc := docWithText("0123\nabcd\nwxyz\nend")
	doc.TagAdd("a", Index{1, 1}, Index{1, 3})
	doc.TagAdd("a", Index{2, 3}, Index{3, 1})
	doc.TagAdd("a", Index{4, 0}, Index{4, 2})
	doc.TagAdd("b", Index{1, 0}, Index{4, 3})
	for line, want := range map[int]int{1: 2, 2: 2, 3: 2, 4: 2} {
		n := 0
		doc.tagsOnLine(line, func(*Tag, TagRange) { n++ })
		if n != want {
			t.Errorf("line %d: %d ranges, want %d", line, n, want)
		}
	}
	doc.TagRemove("b", Index{1, 0}, Index{4, 3})
	n := 0
	doc.tagsOnLine(2, func(_ *Tag, tr TagRange) {
		n++
		if tr.Start != (Index{2, 3}) {
			t.Errorf("line 2 got range %v", tr)
		}
	})
	if n != 1 {
		t.Errorf("line 2: %d ranges, want 1", n)
	}
}

func TestTagDeleteMergesTrimmedPair(t *testing.T) {
	doc := docWithText("0123456789")
	doc.TagAdd("b", Index{1, 2}, Index{1, 5})
	doc.TagAdd("b", Index{1, 7}, Index{1, 9})
	doc.Delete(Index{1, 4}, Index{1, 8})
	got := rangesOf(doc, "b")
	want := [][2]Index{{{1, 2}, {1, 5}}}
	if !slices.Equal(got, want) {
		t.Fatalf("ranges = %v, want %v", got, want)
	}
}

func TestPutTagReplaces(t *testing.T) {
	doc := docWithText("0123456789")
	doc.putTag(&Tag{Name: "sel", Priority: selPriority})
	doc.putTag(&Tag{Name: "sel", Priority: selPriority, Underline: true})
	doc.TagAdd("sel", Index{1, 0}, Index{1, 3})
	tags := doc.TagsAt(Index{1, 1})
	if len(tags) != 1 || !tags[0].Underline {
		t.Fatalf("TagsAt = %v", tags)
	}
}

func TestSetSelectionUnchangedIsNoop(t *testing.T) {
	tw := newBenchWidget(docWithText("hello world"), 200, 100)
	tw.setSelection(Index{1, 2}, Index{1, 6})
	before := tw.doc.TagRangesFor("sel")
	tw.setSelection(Index{1, 6}, Index{1, 2})
	after := tw.doc.TagRangesFor("sel")
	if len(after) != 1 || after[0].Start != (Index{1, 2}) || after[0].End != (Index{1, 6}) {
		t.Fatalf("sel = %v", after)
	}
	if &before[0] != &after[0] {
		t.Fatal("unchanged selection was rewritten")
	}
	tw.setSelection(Index{1, 2}, Index{1, 2})
	if tw.HasSelection() {
		t.Fatal("empty selection left sel ranges")
	}
}
