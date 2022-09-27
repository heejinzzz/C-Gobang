package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

func newCheckerboardCell(img *canvas.Image) *checkerboardCell {
	cell := &checkerboardCell{Img: img}
	cell.ExtendBaseWidget(cell)
	return cell
}

type checkerboardCell struct {
	widget.BaseWidget
	Img   *canvas.Image
	OnTap func()
}

func (cell *checkerboardCell) SetOnTap(onTap func()) {
	cell.OnTap = onTap
}

func (cell *checkerboardCell) Tapped(_ *fyne.PointEvent) {
	if cell.OnTap == nil {
		return
	}
	cell.OnTap()
}

func (cell *checkerboardCell) CreateRenderer() fyne.WidgetRenderer {
	renderer := &checkerboardCellRenderer{Img: cell.Img}
	return renderer
}

type checkerboardCellRenderer struct {
	Img *canvas.Image
}

func (renderer *checkerboardCellRenderer) MinSize() fyne.Size {
	return fyne.NewSize(1, 1)
}

func (renderer *checkerboardCellRenderer) Layout(size fyne.Size) {
	renderer.Img.Resize(size)
}

func (renderer *checkerboardCellRenderer) Destroy() {
}

func (renderer *checkerboardCellRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{renderer.Img}
}

func (renderer *checkerboardCellRenderer) Refresh() {
	renderer.Img.Refresh()
}
