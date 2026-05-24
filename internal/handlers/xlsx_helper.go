package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"
	excelize "github.com/xuri/excelize/v2"
)

// XLSXColumn defines a column header and its hint text.
type XLSXColumn struct {
	Header string
	Hint   string
}

// writeXLSXTemplate generates an Excel template with a styled header row,
// a hint row, and writes it to the response.
func writeXLSXTemplate(c *echo.Context, filename string, cols []XLSXColumn) error {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"

	// --- header style: indigo background, white bold text ---
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"4F46E5"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Color: "FFFFFF", Size: 11},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "FFFFFF", Style: 1},
			{Type: "right", Color: "FFFFFF", Style: 1},
		},
	})

	// --- hint style: light yellow background, italic gray text ---
	hintStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFBEB"}, Pattern: 1},
		Font: &excelize.Font{Italic: true, Color: "92400E", Size: 10},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
			WrapText:   true,
		},
	})

	// Text number format so data cells preserve leading zeros (e.g. CCCD starting with 0)
	textStyle, _ := f.NewStyle(&excelize.Style{NumFmt: 49}) // 49 = "@" (text)

	// Write header row (row 1) and hint row (row 2)
	for i, col := range cols {
		colLetter, _ := excelize.ColumnNumberToName(i + 1)
		headerCell := colLetter + "1"
		hintCell := colLetter + "2"

		_ = f.SetCellValue(sheet, headerCell, col.Header)
		_ = f.SetCellStyle(sheet, headerCell, headerCell, headerStyle)

		_ = f.SetCellValue(sheet, hintCell, col.Hint)
		_ = f.SetCellStyle(sheet, hintCell, hintCell, hintStyle)

		// Set column width based on hint length (min 18, max 35)
		width := len([]rune(col.Hint)) / 2
		if width < 18 {
			width = 18
		}
		if width > 35 {
			width = 35
		}
		_ = f.SetColWidth(sheet, colLetter, colLetter, float64(width))

		// Apply text format to data rows so Excel treats all input as text
		_ = f.SetColStyle(sheet, colLetter, textStyle)
		// Re-apply per-cell styles for header/hint since SetColStyle can reset them
		_ = f.SetCellStyle(sheet, headerCell, headerCell, headerStyle)
		_ = f.SetCellStyle(sheet, hintCell, hintCell, hintStyle)
	}

	// Freeze top row so header stays visible when scrolling
	_ = f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	// Row heights
	_ = f.SetRowHeight(sheet, 1, 28)
	_ = f.SetRowHeight(sheet, 2, 40)

	c.Response().Header().Set("Content-Type",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Response().Header().Set("Content-Disposition",
		`attachment; filename="`+filename+`"`)
	c.Response().WriteHeader(http.StatusOK)

	return f.Write(c.Response())
}
