package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"net/url"
)

func newFooter() *fyne.Container {
	label1 := widget.NewLabel("Written by:")
	logo := widget.NewIcon(myLogoBlue)
	label2 := widget.NewLabel("heejinzzz")
	ctn1 := container.NewHBox(label1, logo, label2)

	label3 := widget.NewLabel("Follow me on github:")
	myUrl, _ := url.Parse("https://github.com/heejinzzz")
	link := widget.NewHyperlink("github.com/heejinzzz", myUrl)
	ctn2 := container.NewHBox(label3, link)

	label4 := widget.NewLabel("Contact me by email:  1273860443@qq.com")

	return container.NewVBox(ctn1, ctn2, label4)
}
