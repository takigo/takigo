package dialog

import (
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/label"
)

// MessageType identifies the kind of message.
type MessageType int

const (
	MsgInfo     MessageType = iota
	MsgWarning
	MsgError
	MsgQuestion
)

// ButtonSet defines which buttons appear in a message box.
type ButtonSet int

const (
	BtnOK              ButtonSet = iota
	BtnOKCancel
	BtnYesNo
	BtnYesNoCancel
	BtnAbortRetryIgnore
)

// messageConfig holds ShowMessage options.
type messageConfig struct {
	title   string
	message string
	detail  string
	msgType MessageType
	buttons ButtonSet
}

// MessageOption configures ShowMessage.
type MessageOption func(*messageConfig)

func MsgTitle(s string) MessageOption     { return func(c *messageConfig) { c.title = s } }
func MsgMessage(s string) MessageOption   { return func(c *messageConfig) { c.message = s } }
func MsgDetail(s string) MessageOption    { return func(c *messageConfig) { c.detail = s } }
func MsgType(t MessageType) MessageOption { return func(c *messageConfig) { c.msgType = t } }
func MsgButtons(b ButtonSet) MessageOption { return func(c *messageConfig) { c.buttons = b } }

// ShowMessage displays a modal message box and returns the user's choice.
func ShowMessage(parent widget.Caregiver, opts ...MessageOption) DialogResult {
	cfg := messageConfig{
		title:   "Message",
		message: "",
		buttons: BtnOK,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.title == "" {
		switch cfg.msgType {
		case MsgInfo:
			cfg.title = "Information"
		case MsgWarning:
			cfg.title = "Warning"
		case MsgError:
			cfg.title = "Error"
		case MsgQuestion:
			cfg.title = "Question"
		}
	}

	d := New(parent, cfg.title, 350, 150)

	// Layout: icon on left, message+detail on right.
	bodyFrame := d.Content

	// Icon label (unicode).
	iconText := msgIcon(cfg.msgType)
	iconLabel := label.New(bodyFrame, "icon",
		label.Text(iconText),
		label.FontOpt("TkDefaultFont"),
	)
	pack.Pack(iconLabel, pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(10))

	// Message text.
	msgText := cfg.message
	if cfg.detail != "" {
		msgText += "\n" + cfg.detail
	}
	msgLabel := label.New(bodyFrame, "msg",
		label.Text(msgText),
		label.Anchor(0), // AnchorNW
	)
	pack.Pack(msgLabel, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(5), pack.PadY(10))

	// Buttons.
	addButtons(d, msgButtons(cfg.buttons))

	return d.Run()
}

func msgIcon(t MessageType) string {
	switch t {
	case MsgInfo:
		return "\u2139" // ℹ
	case MsgWarning:
		return "\u26A0" // ⚠
	case MsgError:
		return "\u2716" // ✖
	case MsgQuestion:
		return "?"
	default:
		return "\u2139"
	}
}

func msgButtons(bs ButtonSet) []dialogButton {
	switch bs {
	case BtnOK:
		return []dialogButton{{text: "OK", result: ResultOK, isDefault: true}}
	case BtnOKCancel:
		return []dialogButton{
			{text: "OK", result: ResultOK, isDefault: true},
			{text: "Cancel", result: ResultCancel},
		}
	case BtnYesNo:
		return []dialogButton{
			{text: "Yes", result: ResultYes, isDefault: true},
			{text: "No", result: ResultNo},
		}
	case BtnYesNoCancel:
		return []dialogButton{
			{text: "Yes", result: ResultYes, isDefault: true},
			{text: "No", result: ResultNo},
			{text: "Cancel", result: ResultCancel},
		}
	case BtnAbortRetryIgnore:
		return []dialogButton{
			{text: "Abort", result: ResultAbort},
			{text: "Retry", result: ResultRetry, isDefault: true},
			{text: "Ignore", result: ResultIgnore},
		}
	default:
		return []dialogButton{{text: "OK", result: ResultOK, isDefault: true}}
	}
}
