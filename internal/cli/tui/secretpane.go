package tui

import (
	"image"
	"strings"

	ui "github.com/metaspartan/gotui/v5"
	"github.com/metaspartan/gotui/v5/widgets"
)

type secretField struct {
	name         string
	payload      string
	visible      bool
	defaultField bool
}

type SecretPane struct {
	ui.Block

	fields      []secretField
	selected    int
	scroll      int
	placeholder string
	focused     bool
	needsScroll bool

	normalBorder ui.Style
	focusBorder  ui.Style
	scrollbar    *widgets.Scrollbar
}

func NewSecretPane() *SecretPane {
	p := &SecretPane{
		Block:        *ui.NewBlock(),
		selected:     -1,
		placeholder:  "Select a secret",
		normalBorder: ui.Theme.Block.Border,
		focusBorder:  ui.NewStyle(ui.ColorGreen),
		scrollbar:    widgets.NewScrollbar(),
	}
	p.BorderRounded = true
	p.Title = "Secret"
	p.scrollbar.Border = false
	p.scrollbar.BeginRune = 0
	p.scrollbar.EndRune = 0
	return p
}

func (p *SecretPane) SetFocused(focused bool) {
	p.focused = focused
	if focused {
		p.BorderStyle = p.focusBorder
	} else {
		p.BorderStyle = p.normalBorder
	}
}

func (p *SecretPane) SetFields(fields []secretField, placeholder string) {
	p.fields = fields
	p.placeholder = placeholder
	if len(fields) == 0 {
		p.selected = -1
	} else {
		p.selected = 0
	}
	p.scroll = 0
	p.needsScroll = true
}

func (p *SecretPane) selectedField() (secretField, bool) {
	if p.selected < 0 || p.selected >= len(p.fields) {
		return secretField{}, false
	}
	return p.fields[p.selected], true
}

func (p *SecretPane) MoveSelection(delta int) {
	if len(p.fields) == 0 {
		return
	}
	next := p.selected + delta
	if next < 0 {
		next = 0
	}
	if next >= len(p.fields) {
		next = len(p.fields) - 1
	}
	if next != p.selected {
		p.selected = next
		p.needsScroll = true
	}
}

func (p *SecretPane) ToggleVisible() {
	if p.selected < 0 || p.selected >= len(p.fields) {
		return
	}
	p.fields[p.selected].visible = !p.fields[p.selected].visible
	p.needsScroll = true
}

func (p *SecretPane) Draw(buf *ui.Buffer) {
	p.Block.Draw(buf)

	if len(p.fields) == 0 {
		p.drawPlaceholder(buf)
		return
	}

	content := p.Inner
	if content.Dx() <= 0 || content.Dy() <= 0 {
		return
	}

	if content.Dx() > 2 {
		content.Max.X--
	}

	heights := p.fieldHeights(content.Dx())
	visibleCount := p.adjustScroll(heights, content.Dy())
	if visibleCount == 0 {
		return
	}

	y := content.Min.Y
	for i := p.scroll; i < len(p.fields); i++ {
		height := heights[i]
		if height < 3 {
			height = 3
		}
		if y+height > content.Max.Y {
			if y == content.Min.Y {
				height = content.Max.Y - y
			} else {
				break
			}
		}

		rect := image.Rect(content.Min.X, y, content.Max.X, y+height)
		p.drawField(buf, rect, p.fields[i], i == p.selected)

		y += height
		if i < len(p.fields)-1 && y+fieldGap <= content.Max.Y {
			y += fieldGap
		}

		if y >= content.Max.Y {
			break
		}
	}

	if len(p.fields) > visibleCount && content.Dx() < p.Inner.Dx() {
		p.drawScrollbar(buf, visibleCount)
	}
}

func (p *SecretPane) drawPlaceholder(buf *ui.Buffer) {
	if p.placeholder == "" {
		return
	}
	style := ui.NewStyle(ui.ColorGrey)
	buf.SetString(p.placeholder, style, image.Pt(p.Inner.Min.X, p.Inner.Min.Y))
}

const fieldGap = 1

func (p *SecretPane) fieldHeights(width int) []int {
	heights := make([]int, len(p.fields))
	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}
	for i, field := range p.fields {
		text := p.fieldText(field)
		lines := countWrappedLines(text, innerWidth)
		if lines < 1 {
			lines = 1
		}
		heights[i] = lines + 2
	}
	return heights
}

func (p *SecretPane) adjustScroll(heights []int, viewHeight int) int {
	if len(p.fields) == 0 {
		return 0
	}
	if p.scroll < 0 {
		p.scroll = 0
	}
	if p.scroll >= len(p.fields) {
		p.scroll = len(p.fields) - 1
	}

	if p.needsScroll {
		p.needsScroll = false
		p.scroll = p.selected
	}

	_, last := p.visibleRange(heights, viewHeight, p.scroll)
	if p.selected < p.scroll {
		p.scroll = p.selected
	} else if p.selected > last {
		p.scroll = p.selected
	}

	count, _ := p.visibleRange(heights, viewHeight, p.scroll)
	return count
}

func (p *SecretPane) visibleRange(heights []int, viewHeight, start int) (count, last int) {
	if start < 0 {
		start = 0
	}
	y := 0
	count = 0
	last = start - 1

	for i := start; i < len(heights); i++ {
		height := heights[i]
		if height < 3 {
			height = 3
		}
		if count > 0 {
			y += fieldGap
		}
		if y+height > viewHeight {
			if count == 0 {
				last = i
				count = 1
			}
			break
		}
		y += height
		count++
		last = i
		if y >= viewHeight {
			break
		}
	}

	if count == 0 {
		return 0, start
	}

	return count, last
}

func (p *SecretPane) drawField(buf *ui.Buffer, rect image.Rectangle, field secretField, selected bool) {
	paragraph := widgets.NewParagraph()
	paragraph.SetRect(rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y)
	paragraph.BorderRounded = true
	paragraph.Title = fieldTitle(field)
	paragraph.WrapText = true
	paragraph.Text = p.fieldText(field)

	if selected {
		paragraph.BorderStyle = p.focusBorder
	} else {
		paragraph.BorderStyle = p.normalBorder
	}

	paragraph.Draw(buf)
}

func fieldTitle(field secretField) string {
	if field.defaultField {
		return field.name + " (default)"
	}
	return field.name
}

func (p *SecretPane) fieldText(field secretField) string {
	if field.visible {
		return field.payload
	}
	return "***"
}

func (p *SecretPane) drawScrollbar(buf *ui.Buffer, visibleCount int) {
	p.scrollbar.Max = len(p.fields)
	p.scrollbar.Current = p.scroll
	p.scrollbar.PageSize = visibleCount
	p.scrollbar.SetRect(p.Inner.Max.X-1, p.Inner.Min.Y, p.Inner.Max.X, p.Inner.Max.Y)
	p.scrollbar.Draw(buf)
}

func countWrappedLines(text string, width int) int {
	if width <= 0 {
		return 1
	}
	lines := strings.Split(text, "\n")
	count := 0
	for _, line := range lines {
		cells := ui.ParseStyles(line, ui.NewStyle(ui.ColorWhite))
		cells = ui.WrapCells(cells, uint(width))
		rows := ui.SplitCells(cells, '\n')
		if len(rows) == 0 {
			count++
		} else {
			count += len(rows)
		}
	}
	return count
}
