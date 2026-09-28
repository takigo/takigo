package font

import (
	"strings"
	"testing"
)

func TestParseDescriptorXLFD(t *testing.T) {
	tests := []struct {
		input   string
		want    Attributes
		wantErr bool
	}{
		{
			input: "-*-helvetica-bold-r-normal--*-120-*-*-*-*-*-*",
			want:  Attributes{Family: "helvetica", Weight: WeightBold, Slant: SlantRoman, Size: 12},
		},
		{
			input: "-*-times-medium-i-normal--*-140-*-*-*-*-*-*",
			want:  Attributes{Family: "times", Weight: WeightNormal, Slant: SlantItalic, Size: 14},
		},
		{
			// XLFD with pixel size 0 = use default (12pt)
			input: "-*-courier-bold-o-normal--*-0-*-*-*-*-*-*",
			want:  Attributes{Family: "courier", Weight: WeightBold, Slant: SlantOblique, Size: 12},
		},
		{
			// Not enough dashes for XLFD, falls to simple format
			input: "invalid",
			want:  Attributes{Family: "invalid", Size: 12},
		},
		{
			// 3 dashes < 13, falls to simple format
			input: "-*-incomplete",
			want:  Attributes{Family: "-*-incomplete", Size: 12},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDescriptor(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDescriptor(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseDescriptor(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got.Family != tt.want.Family {
				t.Errorf("Family = %q, want %q", got.Family, tt.want.Family)
			}
			if got.Weight != tt.want.Weight {
				t.Errorf("Weight = %v, want %v", got.Weight, tt.want.Weight)
			}
			if got.Slant != tt.want.Slant {
				t.Errorf("Slant = %v, want %v", got.Slant, tt.want.Slant)
			}
			if got.Size != tt.want.Size {
				t.Errorf("Size = %v, want %v", got.Size, tt.want.Size)
			}
		})
	}
}

func TestParseDescriptorOptionValue(t *testing.T) {
	tests := []struct {
		input   string
		want    Attributes
		wantErr bool
	}{
		{
			input: "-family Times -size 12 -weight bold",
			want:  Attributes{Family: "Times", Size: 12, Weight: WeightBold},
		},
		{
			input: "-family Helvetica -size 10 -slant italic",
			want:  Attributes{Family: "Helvetica", Size: 10, Slant: SlantItalic},
		},
		{
			input: "-family Courier -size 14 -weight bold -slant oblique -underline 1 -overstrike true",
			want:  Attributes{Family: "Courier", Size: 14, Weight: WeightBold, Slant: SlantOblique, Underline: true, Overstrike: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDescriptor(tt.input)
			if err != nil {
				t.Errorf("ParseDescriptor(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got.Family != tt.want.Family {
				t.Errorf("Family = %q, want %q", got.Family, tt.want.Family)
			}
			if got.Size != tt.want.Size {
				t.Errorf("Size = %v, want %v", got.Size, tt.want.Size)
			}
			if got.Weight != tt.want.Weight {
				t.Errorf("Weight = %v, want %v", got.Weight, tt.want.Weight)
			}
			if got.Slant != tt.want.Slant {
				t.Errorf("Slant = %v, want %v", got.Slant, tt.want.Slant)
			}
			if got.Underline != tt.want.Underline {
				t.Errorf("Underline = %v, want %v", got.Underline, tt.want.Underline)
			}
			if got.Overstrike != tt.want.Overstrike {
				t.Errorf("Overstrike = %v, want %v", got.Overstrike, tt.want.Overstrike)
			}
		})
	}
}

func TestParseDescriptorSimple(t *testing.T) {
	tests := []struct {
		input string
		want  Attributes
	}{
		{"Helvetica 12", Attributes{Family: "Helvetica", Size: 12}},
		{"Times 14 bold", Attributes{Family: "Times", Size: 14, Weight: WeightBold}},
		{"Courier 10 italic", Attributes{Family: "Courier", Size: 10, Slant: SlantItalic}},
		{"Helvetica 12 bold italic", Attributes{Family: "Helvetica", Size: 12, Weight: WeightBold, Slant: SlantItalic}},
		{"sans-serif", DefaultAttributes()},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDescriptor(tt.input)
			if err != nil {
				t.Errorf("ParseDescriptor(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got.Family != tt.want.Family {
				t.Errorf("Family = %q, want %q", got.Family, tt.want.Family)
			}
			if got.Size != tt.want.Size {
				t.Errorf("Size = %v, want %v", got.Size, tt.want.Size)
			}
			if got.Weight != tt.want.Weight {
				t.Errorf("Weight = %v, want %v", got.Weight, tt.want.Weight)
			}
			if got.Slant != tt.want.Slant {
				t.Errorf("Slant = %v, want %v", got.Slant, tt.want.Slant)
			}
		})
	}
}

func TestAttributesDescriptor(t *testing.T) {
	tests := []struct {
		attrs Attributes
		want  string
	}{
		{Attributes{Family: "Helvetica", Size: 12, Weight: WeightBold, Slant: SlantItalic}, "Helvetica Bold Italic 12"},
		{Attributes{Family: "Times", Size: 14}, "Times 14"},
		{Attributes{Family: "", Size: 0}, "sans-serif 10"},
		{Attributes{Family: "Courier", Size: 10, Weight: WeightBold, Slant: SlantOblique}, "Courier Bold Oblique 10"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.attrs.Descriptor()
			if got != tt.want {
				t.Errorf("Descriptor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultAttributes(t *testing.T) {
	def := DefaultAttributes()
	if def.Family != "sans-serif" {
		t.Errorf("Family = %q, want sans-serif", def.Family)
	}
	if def.Size != 12 {
		t.Errorf("Size = %v, want 12", def.Size)
	}
	if def.Weight != WeightNormal {
		t.Errorf("Weight = %v, want WeightNormal", def.Weight)
	}
	if def.Slant != SlantRoman {
		t.Errorf("Slant = %v, want SlantRoman", def.Slant)
	}
}

func TestMetricsLinespace(t *testing.T) {
	m := Metrics{Ascent: 10, Descent: 4}
	if m.Linespace() != 14 {
		t.Errorf("Linespace() = %d, want 14", m.Linespace())
	}
}

func TestRegistryDefineGet(t *testing.T) {
	// Use a mock opener that always fails (we test the registry logic, not font opening)
	opener := &mockOpener{}
	reg := NewRegistry(opener)

	// Test Define with new attributes
	attrs := Attributes{Family: "TestFont", Size: 12, Weight: WeightBold}
	reg.Define("TestFont", attrs)

	// Verify GetAttrs returns the defined attributes
	gotAttrs, ok := reg.GetAttrs("TestFont")
	if !ok {
		t.Error("GetAttrs should return true for defined font")
	}
	if gotAttrs.Family != "TestFont" || gotAttrs.Size != 12 || gotAttrs.Weight != WeightBold {
		t.Errorf("GetAttrs = %+v, want %+v", gotAttrs, attrs)
	}

	// Test GetAttrs for non-existent font
	_, ok = reg.GetAttrs("NonExistent")
	if ok {
		t.Error("GetAttrs should return false for non-existent font")
	}
}

func TestRegistryDerive(t *testing.T) {
	opener := &mockOpener{}
	reg := NewRegistry(opener)

	reg.Define("TestFont", Attributes{Family: "Base", Size: 10, Weight: WeightNormal})

	// Derive with size override
	derived := reg.Derive("TestFont", 14, 0)
	if !strings.Contains(derived, "Base") || !strings.Contains(derived, "14") {
		t.Errorf("Derive with size = %q, want to contain Base and 14", derived)
	}

	// Derive with weight override
	derived = reg.Derive("TestFont", 0, WeightBold)
	if !strings.Contains(derived, "Base") || !strings.Contains(derived, "Bold") {
		t.Errorf("Derive with weight = %q, want to contain Base and Bold", derived)
	}

	// Derive non-existent font (uses fallback)
	derived = reg.Derive("NonExistent", 12, WeightBold)
	if !strings.Contains(derived, "sans-serif") || !strings.Contains(derived, "12") || !strings.Contains(derived, "Bold") {
		t.Errorf("Derive fallback = %q, want to contain sans-serif, 12, Bold", derived)
	}
}

func TestRegistryClose(t *testing.T) {
	opener := &mockOpener{}
	reg := NewRegistry(opener)

	// Should not panic
	reg.Close()
}

type mockOpener struct{}

func (m *mockOpener) OpenFont(attrs Attributes) (Font, error) {
	return nil, &mockFontError{}
}

func (m *mockOpener) Families() []string { return nil }

type mockFontError struct{}

func (e *mockFontError) Error() string { return "mock font error" }
