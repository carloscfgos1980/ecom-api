package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	xls "github.com/extrame/xls"
	_ "github.com/lib/pq"
	"github.com/tealeg/xlsx"

	"github.com/carloscfgos1980/ecom-api/internal/config"
)

// productRow holds a parsed row from the Excel file
type productRow struct {
	Id          int64
	Name        string
	Price       string
	Quantity    int32
	Description string
}

// xlsEpoch is used to convert Excel date serials back to numbers.
// The extrame/xls library parses numeric cells as time.Time strings
// (RFC3339), so we reverse-engineer the original serial number.
var xlsEpoch = time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)

// parseXLSNumeric converts the string value from extrame/xls (which may be an
// RFC3339 date-serial string or a plain number) back to the original integer.
func parseXLSNumeric(s string) (int64, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// not a date-serial string — try parsing as a plain number
		f, err2 := strconv.ParseFloat(s, 64)
		if err2 != nil {
			return 0, fmt.Errorf("cannot parse %q as number: %v", s, err)
		}
		return int64(math.Round(f)), nil
	}
	days := t.Sub(xlsEpoch).Hours() / 24
	return int64(math.Round(days)), nil
}

// readXLS parses products_start.xls and returns product rows (header row skipped).
func readXLS(path string) ([]productRow, error) {
	f, err := xls.Open(path, "utf-8")
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}

	sheet := f.GetSheet(0)
	if sheet == nil {
		return nil, fmt.Errorf("no sheets found in %s", path)
	}

	var rows []productRow
	for i := 1; i <= int(sheet.MaxRow); i++ { // row 0 is the header
		row := sheet.Row(i)
		if row.LastCol() < 5 {
			continue
		}
		id := row.Col(0)
		if id == "" {
			continue
		}
		idInt, err := parseXLSNumeric(id)
		if err != nil {
			log.Printf("row %d: skipping — invalid id %q: %v", i+1, id, err)
			continue
		}

		name := strings.TrimSpace(row.Col(1))
		if name == "" {
			continue
		}

		// price — stored as a numeric Excel serial by the library
		priceSerial, err := parseXLSNumeric(row.Col(2))
		if err != nil {
			log.Printf("row %d: skipping — invalid price %q: %v", i+1, row.Col(2), err)
			continue
		}

		// quantity — same issue
		qtySerial, err := parseXLSNumeric(row.Col(3))
		if err != nil {
			log.Printf("row %d: skipping — invalid quantity %q: %v", i+1, row.Col(3), err)
			continue
		}

		desc := strings.TrimSpace(row.Col(4))

		rows = append(rows, productRow{
			Id:          idInt,
			Name:        name,
			Price:       strconv.FormatInt(priceSerial, 10),
			Quantity:    int32(qtySerial),
			Description: desc,
		})
	}
	return rows, nil
}

// insertProducts bulk-inserts rows into the products table using the XLS id.
// Existing rows with the same id are silently skipped.
func insertProducts(db *sql.DB, rows []productRow) {
	const query = `
		INSERT INTO products (id, name, price, quantity, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`

	inserted := 0
	for _, r := range rows {
		_, err := db.Exec(query, r.Id, r.Name, r.Price, r.Quantity, r.Description)
		if err != nil {
			log.Printf("failed to insert %q (id=%d): %v", r.Name, r.Id, err)
			continue
		}
		log.Printf("inserted → id: %-5d name: %-20s price: %-8s qty: %d  desc: %s",
			r.Id, r.Name, r.Price, r.Quantity, r.Description)
		inserted++
	}
	log.Printf("done — %d/%d rows inserted", inserted, len(rows))
}

// exportProducts queries all rows from the products table and writes them to an xlsx file.
func exportProducts(db *sql.DB, path string) error {
	rows, err := db.Query(`SELECT id, name, price, quantity, description FROM products ORDER BY id`)
	if err != nil {
		return fmt.Errorf("querying products: %w", err)
	}
	defer rows.Close()

	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Products")
	if err != nil {
		return fmt.Errorf("creating sheet: %w", err)
	}

	// header row
	header := sheet.AddRow()
	for _, h := range []string{"id", "name", "price", "quantity", "description"} {
		header.AddCell().SetString(h)
	}

	count := 0
	for rows.Next() {
		var id int64
		var name, price, description string
		var quantity int32
		if err := rows.Scan(&id, &name, &price, &quantity, &description); err != nil {
			return fmt.Errorf("scanning row: %w", err)
		}
		r := sheet.AddRow()
		r.AddCell().SetInt64(id)
		r.AddCell().SetString(name)
		r.AddCell().SetString(price)
		r.AddCell().SetInt(int(quantity))
		r.AddCell().SetString(description)
		count++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating rows: %w", err)
	}

	if err := file.Save(path); err != nil {
		return fmt.Errorf("saving xlsx to %s: %w", path, err)
	}
	log.Printf("exported %d products → %s", count, path)
	return nil
}

func main() {
	// Command-line flags
	var file string
	var sheet string
	var mode string
	flag.StringVar(&file, "file", "", "Path to .xls or .xlsx file")
	flag.StringVar(&sheet, "sheet", "", "Sheet name for .xlsx")
	flag.StringVar(&mode, "mode", "import", "Mode: import or export")
	flag.Parse()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DB_URL)
	if err != nil {
		log.Fatalf("db open error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("db ping error: %v", err)
	}
	log.Println("Connected to database")
	switch mode {
	case "import":
		// Delete all existing products before importing
		_, err = db.Exec("DELETE FROM products")
		if err != nil {
			log.Fatalf("failed to delete existing products: %v", err)
		}
		log.Println("Deleted all existing products")

		log.Printf("Reading products from %s", cfg.XLS_FILE_PATH_READ)

		rows, err := readXLS(cfg.XLS_FILE_PATH_READ)
		if err != nil {
			log.Fatalf("parse error: %v", err)
		}
		log.Printf("Parsed %d product rows from %s", len(rows), cfg.XLS_FILE_PATH_READ)
		insertProducts(db, rows)
	case "export":
		if err := exportProducts(db, cfg.XLS_FILE_PATH_WRITE); err != nil {
			log.Fatalf("export error: %v", err)
		}
	default:
		log.Fatalf("invalid mode: %s (must be 'import' or 'export')", mode)
	}

}
