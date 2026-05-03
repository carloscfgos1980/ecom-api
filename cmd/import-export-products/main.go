package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/carloscfgos1980/ecom-api/internal/env"
	"github.com/extrame/xls"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
)

// This tool imports products from an .xls or .xlsx file into the database, replacing existing products.
type productRow struct {
	id          int
	name        string
	price       float64
	quantity    int
	description string
}

// Usage:
// go run main.go -file=path/to/products.xlsx -sheet=Sheet1 -mode=import
// go run main.go -file=path/to/products.xlsx -sheet=Sheet1 -mode=export

func main() {
	// Command-line flags
	var file string
	var sheet string
	var mode string
	flag.StringVar(&file, "file", "", "Path to .xls or .xlsx file")
	flag.StringVar(&sheet, "sheet", "", "Sheet name for .xlsx")
	flag.StringVar(&mode, "mode", "import", "Mode: import or export")
	flag.Parse()

	// Load .env values (if available) to allow env vars to override them
	dotenvValues := loadDotEnvValues()
	mode = strings.ToLower(strings.TrimSpace(mode))

	// Determine DSN for database connection
	dsn := env.GetEnv("DB_URL", "")
	if strings.TrimSpace(dsn) == "" {
		dsn = strings.TrimSpace(dotenvValues["DB_URL"])
	}
	if strings.TrimSpace(dsn) == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/db_ecom?sslmode=disable"
	}
	// Connect to the database
	ctx := context.Background()
	// Use pgx directly for simplicity; in a real app, you might use a connection pool or an ORM
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}
	defer conn.Close(ctx)
	// Run the appropriate mode
	switch mode {
	// For import mode, determine file path and sheet, then import products
	case "import":
		// Determine file path for import
		if strings.TrimSpace(file) == "" {
			file = env.GetEnv("XLS_FILE_PATH_READ", "")
			if strings.TrimSpace(file) == "" {
				file = strings.TrimSpace(dotenvValues["XLS_FILE_PATH_READ"])
			}
		}
		// If file path is still empty, use default
		if strings.TrimSpace(file) == "" {
			log.Fatal("missing file path; set -file or XLS_FILE_PATH_READ")
		}
		// Import products from the specified file and sheet
		if err := importProductsFromFile(ctx, conn, file, sheet); err != nil {
			log.Fatal(err)
		}
	// For export mode, determine file path and handle .xls extension by switching to .xlsx
	case "export":
		// Determine file path for export
		if strings.TrimSpace(file) == "" {
			file = env.GetEnv("XLS_FILE_PATH_WRITE", "")
			if strings.TrimSpace(file) == "" {
				file = strings.TrimSpace(dotenvValues["XLS_FILE_PATH_WRITE"])
			}
		}
		// export to data folder by default, but allow override via env or flag
		if err := exportProductsToXLSX(ctx, conn, file, sheet); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("invalid mode %q; use import or export", mode)
	}
}

func importProductsFromFile(ctx context.Context, conn *pgx.Conn, file, sheet string) error {
	rows, err := readRows(file, sheet)
	if err != nil {
		return fmt.Errorf("read rows failed: %w", err)
	}
	if len(rows) == 0 {
		log.Println("no rows to import")
		return nil
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx failed: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM products`); err != nil {
		return fmt.Errorf("failed to delete existing products: %w", err)
	}

	const insertQ = `INSERT INTO products (id, name, price, quantity, description) VALUES ($1, $2, $3, $4, $5)`
	for _, r := range rows {
		_, err := tx.Exec(ctx, insertQ, r.id, r.name, r.price, r.quantity, r.description)
		if err != nil {
			return fmt.Errorf("insert failed for %q: %w", r.name, err)
		}
	}

	if _, err := tx.Exec(ctx, `
		SELECT setval(
			pg_get_serial_sequence('products', 'id'),
			COALESCE((SELECT MAX(id) FROM products), 1),
			true
		)
	`); err != nil {
		return fmt.Errorf("failed to sync products id sequence: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	log.Printf("imported %d products", len(rows))
	return nil
}

func exportProductsToXLSX(ctx context.Context, conn *pgx.Conn, filePath, sheetName string) error {

	if strings.TrimSpace(sheetName) == "" {
		sheetName = "products"
	}

	rows, err := conn.Query(ctx, `
		SELECT id, name, price, quantity, COALESCE(description, '')
		FROM products
		ORDER BY id
	`)
	if err != nil {
		return fmt.Errorf("query products failed: %w", err)
	}
	defer rows.Close()

	f := excelize.NewFile()
	defer f.Close()

	defaultSheet := f.GetSheetName(0)
	if defaultSheet == "" {
		defaultSheet = "Sheet1"
	}
	f.SetSheetName(defaultSheet, sheetName)

	headers := []string{"id", "name", "price", "quantity", "description"}
	for i, h := range headers {
		cellName, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cellName, h)
	}

	rowNum := 2
	for rows.Next() {
		var p productRow
		if err := rows.Scan(&p.id, &p.name, &p.price, &p.quantity, &p.description); err != nil {
			return fmt.Errorf("scan product row failed: %w", err)
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), p.id)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), p.name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), p.price)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), p.quantity)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), p.description)
		rowNum++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("read query rows failed: %w", err)
	}

	if dir := filepath.Dir(filePath); strings.TrimSpace(dir) != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create output directory failed: %w", err)
		}
	}

	if err := f.SaveAs(filePath); err != nil {
		return fmt.Errorf("save xls failed: %w", err)
	}

	log.Printf("exported %d products to %s", rowNum-2, filePath)
	return nil
}

func readRows(path, sheetName string) ([]productRow, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".xls":
		return readXLS(path)
	case ".xlsx":
		return readXLSX(path, sheetName)
	default:
		return nil, fmt.Errorf("unsupported extension: %s", filepath.Ext(path))
	}
}

func readXLSX(path, sheetName string) ([]productRow, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if strings.TrimSpace(sheetName) == "" {
		sheetName = f.GetSheetName(0)
	}
	if sheetName == "" {
		return nil, errors.New("sheet not found")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	return parseRows(rows)
}

func readXLS(path string) ([]productRow, error) {
	wb, err := xls.Open(path, "utf-8")
	if err != nil {
		return nil, err
	}
	if wb.NumSheets() == 0 {
		return nil, errors.New("workbook has no sheets")
	}
	sh := wb.GetSheet(0)
	if sh == nil {
		return nil, errors.New("first sheet not available")
	}

	rows := make([][]string, 0, int(sh.MaxRow)+1)
	for i := 0; i <= int(sh.MaxRow); i++ {
		r := sh.Row(i)
		if r == nil {
			rows = append(rows, nil)
			continue
		}
		rows = append(rows, []string{r.Col(0), r.Col(1), r.Col(2), r.Col(3), r.Col(4)})
	}

	return parseRows(rows)
}

func parseRows(rows [][]string) ([]productRow, error) {
	if len(rows) < 2 {
		return nil, errors.New("file must have header and at least one row")
	}

	headers := mapHeaders(rows[0])
	idIdx, ok := findHeader(headers, "id", "product_id")
	if !ok {
		return nil, errors.New("missing id header")
	}
	nameIdx, ok := findHeader(headers, "name", "product", "product_name")
	if !ok {
		return nil, errors.New("missing name header")
	}
	priceIdx, ok := findHeader(headers, "price", "unit_price")
	if !ok {
		return nil, errors.New("missing price header")
	}
	qtyIdx, ok := findHeader(headers, "quantity", "qty", "stock")
	if !ok {
		return nil, errors.New("missing quantity header")
	}
	descIdx, hasDesc := findHeader(headers, "description", "desc")

	out := make([]productRow, 0, len(rows)-1)
	seenIDs := make(map[int]struct{}, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		r := rows[i]
		idText := cell(r, idIdx)
		name := strings.TrimSpace(cell(r, nameIdx))
		priceText := cell(r, priceIdx)
		qtyText := cell(r, qtyIdx)

		if name == "" && strings.TrimSpace(priceText) == "" && strings.TrimSpace(qtyText) == "" {
			continue
		}
		if name == "" {
			return nil, fmt.Errorf("row %d: empty name", i+1)
		}

		id, err := parseID(idText)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i+1, err)
		}
		if _, exists := seenIDs[id]; exists {
			return nil, fmt.Errorf("row %d: duplicate id %d in file", i+1, id)
		}
		seenIDs[id] = struct{}{}

		price, err := parsePrice(priceText)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i+1, err)
		}
		qty, err := parseQuantity(qtyText)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i+1, err)
		}

		var desc string
		if hasDesc {
			d := strings.TrimSpace(cell(r, descIdx))
			desc = d
		}

		out = append(out, productRow{id: id, name: name, price: price, quantity: qty, description: desc})
	}

	return out, nil
}

func mapHeaders(headerRow []string) map[string]int {
	m := map[string]int{}
	for i, h := range headerRow {
		n := strings.ToLower(strings.TrimSpace(h))
		n = strings.ReplaceAll(n, " ", "_")
		if n != "" {
			m[n] = i
		}
	}
	return m
}

func findHeader(m map[string]int, options ...string) (int, bool) {
	for _, option := range options {
		if i, ok := m[option]; ok {
			return i, true
		}
	}
	return -1, false
}

func cell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return row[idx]
}

func parsePrice(v string) (float64, error) {
	t := strings.TrimSpace(v)
	t = strings.ReplaceAll(t, "$", "")
	t = strings.ReplaceAll(t, ",", "")
	if t == "" {
		return 0, errors.New("price is empty")
	}
	p, err := strconv.ParseFloat(t, 64)
	if err != nil {
		if serial, ok := parseExcelDateSerial(t); ok {
			p = float64(serial)
		} else {
			return 0, fmt.Errorf("invalid price %q", v)
		}
	}
	if p < 0 {
		return 0, errors.New("price cannot be negative")
	}
	return p, nil
}

func parseQuantity(v string) (int, error) {
	t := strings.TrimSpace(v)
	t = strings.ReplaceAll(t, ",", "")
	if t == "" {
		return 0, errors.New("quantity is empty")
	}
	q, err := strconv.Atoi(t)
	if err != nil {
		f, err2 := strconv.ParseFloat(t, 64)
		if err2 != nil || f != float64(int(f)) {
			if serial, ok := parseExcelDateSerial(t); ok {
				q = serial
			} else {
				return 0, fmt.Errorf("invalid quantity %q", v)
			}
		} else {
			q = int(f)
		}
	}
	if q < 0 {
		return 0, errors.New("quantity cannot be negative")
	}
	return q, nil
}

func parseID(v string) (int, error) {
	t := strings.TrimSpace(v)
	t = strings.ReplaceAll(t, ",", "")
	if t == "" {
		return 0, errors.New("id is empty")
	}

	id, err := strconv.Atoi(t)
	if err != nil {
		f, err2 := strconv.ParseFloat(t, 64)
		if err2 != nil || f != float64(int(f)) {
			if serial, ok := parseExcelDateSerial(t); ok {
				id = serial
			} else {
				return 0, fmt.Errorf("invalid id %q", v)
			}
		} else {
			id = int(f)
		}
	}

	if id <= 0 {
		return 0, errors.New("id must be greater than zero")
	}

	return id, nil
}

func parseExcelDateSerial(value string) (int, bool) {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return 0, false
	}
	excelEpoch := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	serial := int(t.Sub(excelEpoch).Hours() / 24)
	if serial < 0 {
		return 0, false
	}
	return serial, true
}

func loadDotEnvValues() map[string]string {
	values := map[string]string{}
	paths := []string{".env", "../.env", "../../.env"}

	for _, p := range paths {
		m, err := godotenv.Read(p)
		if err != nil {
			continue
		}
		for k, v := range m {
			values[k] = v
		}
		_ = godotenv.Overload(p)
	}

	return values
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
