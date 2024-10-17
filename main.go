package main

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func main() {
	go func() {

		w := new(app.Window)
		w.Option(app.Title("Term"), app.Size(unit.Dp(1200), unit.Dp(800)))

		err := run(w)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(w *app.Window) error {
	var ops op.Ops
	grid := NewTermGrid(20, 40)

	if err := grid.SetFontFromPath(grid.LoadSystemFonts()[0]); err != nil {
		log.Printf("Error setting font: %v", err)
	}
	// Создаем кнопку
	var button widget.Clickable

	// Создаем тему для материального дизайна
	th := material.NewTheme()

	grid.SetText("Привет, мир!\nЭто тестовый текст для TermGrid.\nОн занимает несколько строк.")

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// Обновляем размер TermGrid при изменении размера окна
			grid.Resize(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)
			// Отрисовываем кнопку
			button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return material.Button(th, &button, "Нажми меня").Layout(gtx)
			})

			// Обрабатываем нажатие кнопки

			// Отрисовываем TermGrid
			grid.Layout(gtx)

			e.Frame(gtx.Ops)
		case key.Event:
			if e.State == key.Press {
				handleKeyPress(grid, e)
			}
		}
	}
}

func handleKeyPress(grid *TermGrid, e key.Event) {
	switch e.Name {
	// case key.NameUpArrow:
	// 	grid.MoveCursor(0, -1)
	// case key.NameDownArrow:
	// 	grid.MoveCursor(0, 1)
	// case key.NameLeftArrow:
	// 	grid.MoveCursor(-1, 0)
	// case key.NameRightArrow:
	// 	grid.MoveCursor(1, 0)
	case key.NameEnter:
		grid.NewLine()
	default:
		// if e.Modifiers == 0 {
		// 	grid.InsertRune(e.Rune)
		// }
	}
}
