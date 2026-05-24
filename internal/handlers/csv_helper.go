package handlers

import (
	"bytes"
	"encoding/csv"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	excelize "github.com/xuri/excelize/v2"
)

const maxImportSize = 5 * 1024 * 1024 // 5 MB
const maxImportRows = 500

// csvBOM is the UTF-8 byte order mark so Excel reads Vietnamese correctly.
var csvBOM = []byte{0xEF, 0xBB, 0xBF}

func setCSVHeaders(c *echo.Context, filename string) {
	c.Response().Header().Set("Content-Type", "text/csv; charset=utf-8")
	c.Response().Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
}

func writeCSVBOM(c *echo.Context) {
	_, _ = c.Response().Write(csvBOM)
}

// parseUploadedCSV opens the multipart file named "file", skips the header row,
// and returns all data rows. Accepts both .csv and .xlsx files.
func parseUploadedCSV(c *echo.Context) ([][]string, error) {
	fh, err := c.FormFile("file")
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Vui lòng chọn file CSV hoặc Excel")
	}
	if fh.Size > maxImportSize {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "File không được vượt quá 5MB")
	}

	name := strings.ToLower(fh.Filename)
	isXLSX := strings.HasSuffix(name, ".xlsx") || strings.HasSuffix(name, ".xls")

	ct := fh.Header.Get("Content-Type")
	if !isXLSX {
		xlsxCTs := []string{
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			"application/vnd.ms-excel",
		}
		for _, x := range xlsxCTs {
			if ct == x {
				isXLSX = true
				break
			}
		}
	}

	if !isXLSX && ct != "text/csv" && ct != "text/plain" &&
		!strings.HasSuffix(name, ".csv") {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Chỉ chấp nhận file CSV hoặc Excel (.xlsx)")
	}

	f, err := fh.Open()
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Không thể mở file")
	}
	defer f.Close()

	if isXLSX {
		return readXLSXRows(f)
	}
	return readCSVRows(f)
}

func readCSVRows(r multipart.File) ([][]string, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	all, err := reader.ReadAll()
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "File CSV không đúng định dạng")
	}
	if len(all) < 2 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "File không có dữ liệu (thiếu header hoặc dữ liệu)")
	}
	data := all[1:] // skip header
	if len(data) > maxImportRows {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "File vượt quá 500 dòng dữ liệu cho phép")
	}
	return data, nil
}

func readXLSXRows(r io.Reader) ([][]string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Không thể đọc file Excel")
	}

	xf, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "File Excel không đúng định dạng")
	}
	defer xf.Close()

	sheetName := xf.GetSheetName(0)
	rows, err := xf.GetRows(sheetName)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Không thể đọc sheet Excel")
	}

	// Template format: row 1 = column headers, row 2 = hint/example row.
	// Data starts at row 3, so always skip the first 2 rows.
	const skipRows = 2
	if len(rows) <= skipRows {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "File không có dữ liệu (cần ít nhất 1 dòng dữ liệu sau dòng tiêu đề và dòng ví dụ)")
	}
	dataRows := rows[skipRows:]
	if len(dataRows) > maxImportRows {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "File vượt quá 500 dòng dữ liệu cho phép")
	}
	return dataRows, nil
}

// safeCol returns cols[i] or empty string if i is out of range.
func safeCol(cols []string, i int) string {
	if i < len(cols) {
		return cols[i]
	}
	return ""
}
