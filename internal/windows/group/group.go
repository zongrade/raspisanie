package group

import "fyne.io/fyne/v2"

func Group(a *fyne.App) *fyne.Window {
	w := (*a).NewWindow("Выбор группы")
	return &w
}

func GroupWithMenu(a *fyne.App, m fyne.Menu) *fyne.Window {
	w := (*a).NewWindow("Выбор группы")
	return &w
}
