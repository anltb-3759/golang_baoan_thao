package handlers

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	excelize "github.com/xuri/excelize/v2"
)

func newMultipartFile(filename, contentType, content string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	h.Set("Content-Type", contentType)
	fw, _ := w.CreatePart(h)
	_, _ = fw.Write([]byte(content))
	w.Close()
	return body, w.FormDataContentType()
}

func makeXLSXFile(rows [][]string) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer f.Close()
	for i, row := range rows {
		for j, val := range row {
			col, _ := excelize.ColumnNumberToName(j + 1)
			_ = f.SetCellValue("Sheet1", fmt.Sprintf("%s%d", col, i+1), val)
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return &buf, nil
}

func TestReadXLSXRows_Success(t *testing.T) {
	// rows: header, hint, data1, data2
	rows := [][]string{
		{"col1", "col2"},
		{"hint1", "hint2"},
		{"val1", "val2"},
		{"val3", "val4"},
	}
	buf, err := makeXLSXFile(rows)
	if err != nil {
		t.Fatalf("create xlsx: %v", err)
	}

	result, err := readXLSXRows(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 data rows, got %d", len(result))
	}
}

func TestReadXLSXRows_NotEnoughRows(t *testing.T) {
	// Only header + hint, no data rows
	rows := [][]string{
		{"col1"},
		{"hint1"},
	}
	buf, err := makeXLSXFile(rows)
	if err != nil {
		t.Fatalf("create xlsx: %v", err)
	}

	_, err = readXLSXRows(buf)
	if err == nil {
		t.Fatal("expected error for no data rows")
	}
}

func TestReadXLSXRows_InvalidFile(t *testing.T) {
	_, err := readXLSXRows(bytes.NewReader([]byte("not an xlsx file")))
	if err == nil {
		t.Fatal("expected error for invalid xlsx")
	}
}

type nopCloserFile struct{ *bytes.Reader }

func (nopCloserFile) Close() error { return nil }

func TestReadCSVRows_TooManyRows(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("header\n")
	for i := 0; i <= maxImportRows; i++ {
		sb.WriteString("row\n")
	}
	_, err := readCSVRows(nopCloserFile{bytes.NewReader([]byte(sb.String()))})
	if err == nil {
		t.Fatal("expected error for too many rows")
	}
}

func TestReadXLSXRows_TooManyRows(t *testing.T) {
	// Build an XLSX with header + hint + 501 data rows
	rows := make([][]string, maxImportRows+3)
	rows[0] = []string{"col1"}
	rows[1] = []string{"hint"}
	for i := 2; i < len(rows); i++ {
		rows[i] = []string{"data"}
	}
	buf, err := makeXLSXFile(rows)
	if err != nil {
		t.Fatalf("make xlsx: %v", err)
	}
	_, err = readXLSXRows(buf)
	if err == nil {
		t.Fatal("expected error for too many rows")
	}
}

func TestParseUploadedCSV_FileTooLarge(t *testing.T) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="big.csv"`)
	h.Set("Content-Type", "text/csv")
	fw, _ := w.CreatePart(h)
	// Write more than 5MB
	chunk := make([]byte, 1024)
	for i := range chunk {
		chunk[i] = 'a'
	}
	for i := 0; i < 6*1024; i++ {
		_, _ = fw.Write(chunk)
	}
	w.Close()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, err := parseUploadedCSV(c)
	if err == nil {
		t.Fatal("expected error for file too large")
	}
}

func TestParseUploadedCSV_XLSXByContentType(t *testing.T) {
	xlsxBuf, err := makeXLSXFile([][]string{
		{"col1"}, {"hint"}, {"data1"},
	})
	if err != nil {
		t.Fatalf("make xlsx: %v", err)
	}

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="data.bin"`)
	h.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	fw, _ := w.CreatePart(h)
	_, _ = fw.Write(xlsxBuf.Bytes())
	w.Close()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	rows, err := parseUploadedCSV(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(rows))
	}
}

func TestParseUploadedCSV_InvalidContentType(t *testing.T) {
	body, ct := newMultipartFile("data.bin", "application/octet-stream", "some binary data")
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, err := parseUploadedCSV(c)
	if err == nil {
		t.Fatal("expected error for invalid content type")
	}
}
