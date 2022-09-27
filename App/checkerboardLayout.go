package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"math"
)

// 由 layout.gridLayout 修改而来，保证了grid中每个元素的长、宽相等
type checkerboardLayout struct {
	Cols            int
	vertical, adapt bool
}

func (c *checkerboardLayout) horizontal() bool {
	if c.adapt {
		return fyne.IsHorizontal(fyne.CurrentDevice().Orientation())
	}
	return !c.vertical
}

func (c *checkerboardLayout) countRows(objects []fyne.CanvasObject) int {
	count := 0
	for _, child := range objects {
		if child.Visible() {
			count++
		}
	}
	return int(math.Ceil(float64(count) / float64(c.Cols)))
}

func getLeading(size float64, offset int) float32 {
	ret := (size + float64(theme.Padding())) * float64(offset)

	return float32(math.Round(ret))
}

func getTrailing(size float64, offset int) float32 {
	return getLeading(size, offset+1) - theme.Padding()
}

func (c *checkerboardLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	rows := c.countRows(objects)

	padWidth := float32(c.Cols-1) * theme.Padding()
	padHeight := float32(rows-1) * theme.Padding()
	cellWidth := float64(size.Width-padWidth) / float64(c.Cols)
	cellHeight := float64(size.Height-padHeight) / float64(rows)

	if !c.horizontal() {
		padWidth, padHeight = padHeight, padWidth
		cellWidth = float64(size.Width-padWidth) / float64(rows)
		cellHeight = float64(size.Height-padHeight) / float64(c.Cols)
	}

	// 仅在官方库的代码上添加了这一小段，这是唯一修改的地方
	if cellWidth < cellHeight {
		cellHeight = cellWidth
	} else {
		cellWidth = cellHeight
	}

	row, col := 0, 0
	i := 0
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		x1 := getLeading(cellWidth, col)
		y1 := getLeading(cellHeight, row)
		x2 := getTrailing(cellWidth, col)
		y2 := getTrailing(cellHeight, row)

		child.Move(fyne.NewPos(x1, y1))
		child.Resize(fyne.NewSize(x2-x1, y2-y1))

		if c.horizontal() {
			if (i+1)%c.Cols == 0 {
				row++
				col = 0
			} else {
				col++
			}
		} else {
			if (i+1)%c.Cols == 0 {
				col++
				row = 0
			} else {
				row++
			}
		}
		i++
	}
}

func (c *checkerboardLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	rows := c.countRows(objects)
	minSize := fyne.NewSize(0, 0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		minSize = minSize.Max(child.MinSize())
	}

	if c.horizontal() {
		minContentSize := fyne.NewSize(minSize.Width*float32(c.Cols), minSize.Height*float32(rows))
		return minContentSize.Add(fyne.NewSize(theme.Padding()*fyne.Max(float32(c.Cols-1), 0), theme.Padding()*fyne.Max(float32(rows-1), 0)))
	}

	minContentSize := fyne.NewSize(minSize.Width*float32(rows), minSize.Height*float32(c.Cols))
	return minContentSize.Add(fyne.NewSize(theme.Padding()*fyne.Max(float32(rows-1), 0), theme.Padding()*fyne.Max(float32(c.Cols-1), 0)))
}

func newCheckerboardLayout(cols int) fyne.Layout {
	return &checkerboardLayout{Cols: cols}
}
