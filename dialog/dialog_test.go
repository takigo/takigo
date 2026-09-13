package dialog

import (
	"testing"
)

func TestDialogResultConstants(t *testing.T) {
	tests := []struct {
		name  string
		value DialogResult
		want  int
	}{
		{"ResultNone", ResultNone, 0},
		{"ResultOK", ResultOK, 1},
		{"ResultCancel", ResultCancel, 2},
		{"ResultYes", ResultYes, 3},
		{"ResultNo", ResultNo, 4},
		{"ResultAbort", ResultAbort, 5},
		{"ResultRetry", ResultRetry, 6},
		{"ResultIgnore", ResultIgnore, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.value) != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.value, tt.want)
			}
		})
	}
}

func TestDialogResultOrdering(t *testing.T) {
	results := []DialogResult{
		ResultNone, ResultOK, ResultCancel, ResultYes,
		ResultNo, ResultAbort, ResultRetry, ResultIgnore,
	}

	for i, r := range results {
		if int(r) != i {
			t.Errorf("Result %d = %d, want %d", i, r, i)
		}
	}
}

func TestDialogButton(t *testing.T) {
	btn := dialogButton{
		text:      "OK",
		result:    ResultOK,
		isDefault: true,
	}

	if btn.text != "OK" {
		t.Errorf("text = %q, want \"OK\"", btn.text)
	}
	if btn.result != ResultOK {
		t.Errorf("result = %d, want %d", btn.result, ResultOK)
	}
	if !btn.isDefault {
		t.Error("isDefault should be true")
	}
}