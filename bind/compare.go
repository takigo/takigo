package bind

// The comparison of two matching bindings ports the ordering of
// tk/generic/tkBind.c: Compare and CompareModMasks decide between two
// single-event bindings of a tag, IsBetterMatch decides between a
// completed multi-event sequence and a single-event binding, and the
// registration number breaks a full tie in favour of the binding defined
// last.

// repetitions returns the repetition count of the pattern: 1 normally, 2
// for Double and 3 for Triple, as tkBind.c's ParseEventDescription counts
// the MULT_CLICKS modifiers.
func (p *Pattern) repetitions() int {
	switch {
	case p.Modifiers&ModTriple != 0:
		return 3
	case p.Modifiers&ModDouble != 0:
		return 2
	}
	return 1
}

// hasDetail reports whether the pattern names a specific key or button,
// which is Tk's TkPattern.info.
func (p *Pattern) hasDetail() bool {
	return p.KeySym != 0 || p.Button != 0
}

// countSpecialized returns the specialized count of the pattern: its
// repetition count when it names a detail, else 0.
func (p *Pattern) countSpecialized() int {
	if p.hasDetail() {
		return p.repetitions()
	}
	return 0
}

// modMask returns the modifier mask of the pattern. The double and triple
// clicks are not part of it: Tk keeps them in the pattern's count.
func (p *Pattern) modMask() Modifier {
	return p.Modifiers &^ (ModDouble | ModTriple)
}

// countSpecialized returns the sum of the specialized counts of the
// patterns, as tkBind.c's CountSpecialized does.
func (s Sequence) countSpecialized() int {
	n := 0
	for i := range s.Patterns {
		n += s.Patterns[i].countSpecialized()
	}
	return n
}

// totalCount returns the sum of the repetition counts of the patterns, as
// tkBind.c's PatSeq.count.
func (s Sequence) totalCount() int {
	n := 0
	for i := range s.Patterns {
		n += s.Patterns[i].repetitions()
	}
	return n
}

// weightedCount returns the weighted count of tkBind.c's IsBetterMatch:
// the repetition count of the last pattern counts more than that of the
// earlier ones, so a longer sequence wins.
func (s Sequence) weightedCount() int {
	const million = 1_000_000
	weight := [...]int{0, 1, million, million * million, million * million * million}
	n := 0
	for i := range s.Patterns {
		n += weight[min(s.Patterns[i].repetitions(), len(weight)-1)]
	}
	return n
}

// compareSeq ports tkBind.c's Compare: a positive result means that snd
// is the better match, a negative one that fst is.
func compareSeq(fst, snd Sequence) int {
	if d := snd.countSpecialized() - fst.countSpecialized(); d != 0 {
		return d
	}
	return snd.totalCount() - fst.totalCount()
}

// compareModMasks ports tkBind.c's CompareModMasks: a positive result
// means that snd requires strictly more modifiers than fst and is
// therefore the more specific match.
func compareModMasks(fst, snd Sequence) int {
	fstCount, sndCount := 0, 0
	count := func(f, s Modifier) {
		if f == s {
			return
		}
		if f&s == f { // fst is a subset of snd: snd is more specific
			sndCount++
		}
		if s&f == s { // snd is a subset of fst: fst is more specific
			fstCount++
		}
	}
	n := min(len(fst.Patterns), len(snd.Patterns))
	// The masks of all patterns but the last, as Tk's lastModMaskArr.
	for i := 0; i+1 < n; i++ {
		count(fst.Patterns[i].modMask(), snd.Patterns[i].modMask())
	}
	// Finally the masks of the last patterns.
	if n > 0 {
		count(fst.Patterns[n-1].modMask(), snd.Patterns[n-1].modMask())
	}
	return sndCount - fstCount
}
