package main

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/font/opentype"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
)

// TermGrid представляет собой структуру для отображения сетки символов.
type TermGrid struct {
	cells       [][]rune    // Двумерный массив для хранения символов в сетке
	cellSize    image.Point // Размер одной ячейки сетки (ширина, высота)
	textColor   color.NRGBA // Цвет текста
	bgColor     color.NRGBA // Цвет фона
	cursor      image.Point // Позиция курсора в сетке (строка, столбец)
	font        font.Font   // Шрифт для отрисовки текста
	fontSize    unit.Sp     // Размер шрифта
	needsRedraw bool        // Флаг необходимости перерисовки
}

// NewTermGrid создает новый экземпляр TermGrid с заданными азмерами.
func NewTermGrid(rows, cols int) *TermGrid {
	grid := &TermGrid{
		cells:       make([][]rune, rows),
		cellSize:    image.Point{10, 20},                         // Начальный размер ячейки
		textColor:   color.NRGBA{R: 255, G: 255, B: 255, A: 255}, // Белый цвет по умолчанию
		bgColor:     color.NRGBA{R: 0, G: 0, B: 0, A: 255},       // Черный цвет по умолчанию
		cursor:      image.Point{0, 0},
		font:        font.Font{Typeface: "Go"}, // Используем шрифт Go по умолчанию
		fontSize:    unit.Sp(14),
		needsRedraw: true,
	}
	for i := range grid.cells {
		grid.cells[i] = make([]rune, cols)
	}
	return grid
}

func (g *TermGrid) LoadSystemFonts() []string {
	var fontPaths []string

	switch runtime.GOOS {
	case "windows":
		fontPaths = append(fontPaths, filepath.Join(os.Getenv("WINDIR"), "Fonts"))
	case "darwin":
		fontPaths = append(fontPaths, "/Library/Fonts", "/System/Library/Fonts")
	default: // Linux и другие Unix-подобные системы
		fontPaths = append(fontPaths, "/usr/share/fonts", "/usr/local/share/fonts")
		if home, err := os.UserHomeDir(); err == nil {
			fontPaths = append(fontPaths, filepath.Join(home, ".fonts"))
		}
	}

	var fonts []string
	for _, path := range fontPaths {
		filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && (filepath.Ext(path) == ".ttf" || filepath.Ext(path) == ".otf") {
				fonts = append(fonts, path)
			}
			return nil
		})
	}

	return fonts
}

func (g *TermGrid) SetFontFromPath(path string) error {
	// Читаем файл шрифта
	fontData, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл шрифта: %v", err)
	}

	// Парсим шрифт
	_, err = opentype.Parse(fontData)
	if err != nil {
		return fmt.Errorf("не удалось распарсить шрифт: %v", err)
	}

	// Устанавливаем новый шрифт
	g.font = font.Font{Typeface: font.Typeface(filepath.Base(path))}
	g.needsRedraw = true

	return nil
}

func (g *TermGrid) SetTextColor(color color.NRGBA) {
	g.textColor = color
	g.needsRedraw = true
}

func (g *TermGrid) SetBackgroundColor(color color.NRGBA) {
	g.bgColor = color
	g.needsRedraw = true
}

func (g *TermGrid) SetFontSize(size unit.Sp) {
	g.fontSize = size
	g.needsRedraw = true
}

// Layout реализует интерфейс layout.Widget для TermGrid
func (g *TermGrid) Layout(gtx layout.Context) layout.Dimensions {
	// Вычисляем размер сетки
	size := image.Point{
		X: g.cellSize.X * len(g.cells[0]),
		Y: g.cellSize.Y * len(g.cells),
	}

	// Рисуем фон
	paint.FillShape(gtx.Ops, g.bgColor, clip.Rect{Max: size}.Op())

	// Рисуем символы
	for row, line := range g.cells {
		for col, char := range line {
			if char == 0 {
				continue // Пропускаем пустые ячейки
			}
			g.renderCell(gtx, row, col, char)
		}
	}

	// Рисуем курсор
	cursorRect := image.Rectangle{
		Min: image.Point{X: g.cursor.Y * g.cellSize.X, Y: g.cursor.X * g.cellSize.Y},
		Max: image.Point{X: (g.cursor.Y + 1) * g.cellSize.X, Y: (g.cursor.X + 1) * g.cellSize.Y},
	}
	paint.FillShape(gtx.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: 128}, clip.Rect(cursorRect).Op())

	return layout.Dimensions{Size: size}
}

func (g *TermGrid) renderCell(gtx layout.Context, row, col int, char rune) {
	pos := f32.Point{
		X: float32(col * g.cellSize.X),
		Y: float32(row * g.cellSize.Y),
	}
	defer op.Offset(image.Point{
		X: int(pos.X),
		Y: int(pos.Y),
	}).Push(gtx.Ops).Pop()
	paint.ColorOp{Color: g.textColor}.Add(gtx.Ops)
	widget.Label{}.Layout(gtx, text.NewShaper(), g.font, g.fontSize, string(char), op.CallOp{})
}

func (g *TermGrid) SetCell(row, col int, char rune) {
	if row >= 0 && row < len(g.cells) && col >= 0 && col < len(g.cells[0]) {
		g.cells[row][col] = char
		g.needsRedraw = true
	}
}

func (g *TermGrid) SetText(text string) {
	g.cursor = image.Point{0, 0}
	for _, char := range text {
		if char == '\n' {
			g.NewLine()
		} else {
			g.AppendChar(char)
		}
	}
}

func (g *TermGrid) AppendChar(char rune) {
	if g.cursor.Y < len(g.cells[0]) {
		g.cells[g.cursor.X][g.cursor.Y] = char
		g.cursor.Y++
		if g.cursor.Y >= len(g.cells[0]) {
			g.NewLine()
		}
		g.needsRedraw = true
	}
}

func (g *TermGrid) NewLine() {
	if g.cursor.X < len(g.cells)-1 {
		g.cursor.X++
		g.cursor.Y = 0
		g.needsRedraw = true
	}
}

func (g *TermGrid) Backspace() {
	if g.cursor.Y > 0 {
		g.cursor.Y--
		g.cells[g.cursor.X][g.cursor.Y] = 0
	} else if g.cursor.X > 0 {
		g.cursor.X--
		g.cursor.Y = len(g.cells[g.cursor.X]) - 1
		for g.cursor.Y > 0 && g.cells[g.cursor.X][g.cursor.Y-1] == 0 {
			g.cursor.Y--
		}
	}
	g.needsRedraw = true
}

// Resize обновляет размеры ячеек при изменении размера окна
func (g *TermGrid) Resize(width, height int) {
	g.cellSize = image.Point{
		X: width / len(g.cells[0]),
		Y: height / len(g.cells),
	}
	g.needsRedraw = true
}

// Добавим метод для установки шрифта
func (g *TermGrid) SetFont(f font.Font) {
	g.font = f
	g.needsRedraw = true
}
