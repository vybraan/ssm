package tui

import (
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/textarea"
	tea "github.com/charmbracelet/bubbletea/v2"
	lg "github.com/charmbracelet/lipgloss/v2"
)

func newTextarea() textarea.Model {
	t := textarea.New()
	t.Prompt = ""
	t.Placeholder = `Every line is a command and runs in its own ssh session`
	t.ShowLineNumbers = true
	t.Styles.Cursor = textarea.CursorStyle{
		Color:      lg.Color("212"),
		Shape:      tea.CursorBlock,
		Blink:      false,
		BlinkSpeed: 0,
	}
	t.Styles.Focused.Placeholder = focusedPlaceholderStyle
	t.Styles.Focused.CursorLine = cursorLineStyle
	t.Styles.Focused.Base = focusedBorderStyle
	t.Styles.Focused.EndOfBuffer = endOfBufferStyle
	t.Styles.Blurred.Placeholder = placeholderStyle
	t.Styles.Blurred.Base = blurredBorderStyle
	t.Styles.Blurred.EndOfBuffer = endOfBufferStyle
	t.KeyMap.DeleteWordBackward.SetEnabled(false)
	t.KeyMap.LineNext = key.NewBinding(key.WithKeys("down"))
	t.KeyMap.LinePrevious = key.NewBinding(key.WithKeys("up"))
	t.Blur()
	return t
}
