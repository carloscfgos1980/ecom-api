package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/extrame/xls"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/tealeg/xlsx"
)

// xlsNumericString converts a value that the xls library has serialized as an
// Excel date string back to its original integer. The xls library converts all
// numeric cells using Excel's date epoch (Dec 30, 1899).
func xlsDateToInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	t, err := time.Parse("2006-01-02T15:04:05Z", s)
	if err != nil {
		return 0, err
	}
	epoch := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	days := t.Sub(epoch).Hours() / 24
	return int(math.Round(days)), nil
}

func firstExistingPath(candidates ...string) (string, error) {
	for _, p := range candidates {
		clean := filepath.Clean(p)
		if _, err := os.Stat(clean); err == nil {
			return clean, nil
		}
	}
	return "", fmt.Errorf("none of the candidate paths exist: %v", candidates)
}

func importProducts(db *sql.DB) {
	xlsPath, err := firstExistingPath(
		"data-exel/products_start.xls",
		"../data-exel/products_start.xls",
		"../../data-exel/products_start.xls",
	)
	if err != nil {
		log.Fatalf("Error resolving xls path: %v", err)
	}

	xlFile, err := xls.Open(xlsPath, "utf-8")
	if err != nil {
		log.Fatalf("Error opening xls file: %s", err)
	}

	sheet := xlFile.GetSheet(0)
	if sheet == nil {
		log.Fatal("No sheets found in xls file")
	}

	fmt.Printf("Sheet: %s, Rows: %d\n", sheet.Name, sheet.MaxRow)
	if sheet.MaxRow < 1 {
		log.Fatal("Sheet is empty")
	}

	inserted := 0
	skipped := 0
	_, err = db.Exec(`TRUNCATE TABLE products RESTART IDENTITY`)
	if err != nil {
		log.Fatalf("Failed to clear products table before import: %v", err)
	}

	for i := 1; i <= int(sheet.MaxRow); i++ {
		row := sheet.Row(i)
		if row == nil {
			continue
		}

		name := strings.TrimSpace(row.Col(1))
		if name == "" {
			skipped++
			continue
		}

		price, err := xlsDateToInt(row.Col(2))
		if err != nil {
			log.Printf("Row %d: cannot parse price %q: %v", i, row.Col(2), err)
			skipped++
			continue
		}

		quantity, err := xlsDateToInt(row.Col(3))
		if err != nil {
			log.Printf("Row %d: cannot parse quantity %q: %v", i, row.Col(3), err)
			skipped++
			continue
		}

		description := strings.TrimSpace(row.Col(4))

		now := time.Now()
		_, err = db.Exec(
			`INSERT INTO products (name, price, quantity, description, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			name, fmt.Sprintf("%d", price), quantity, description, now, now,
		)
		if err != nil {
			log.Printf("Row %d: failed to insert %q: %v", i, name, err)
			skipped++
			continue
		}
		inserted++
		fmt.Printf("Inserted: %s (price=%d, qty=%d)\n", name, price, quantity)
	}

	fmt.Printf("\nDone. Inserted: %d, Skipped: %d\n", inserted, skipped)
}

func exportProducts(db *sql.DB) {
	rows, err := db.Query(`SELECT id, name, price::text, quantity, description FROM products ORDER BY id ASC`)
	if err != nil {
		log.Fatalf("Error querying products: %v", err)
	}
	defer rows.Close()

	outPath, err := firstExistingPath("data-exel", "../data-exel", "../../data-exel")
	if err != nil {
		log.Fatalf("Error resolving output directory: %v", err)
	}

	file := xlsx.NewFile()
	sheet, err := file.AddSheet("products")
	if err != nil {
		log.Fatalf("Error creating sheet: %v", err)
	}

	header := sheet.AddRow()
	header.AddCell().SetString("Id")
	header.AddCell().SetString("name")
	header.AddCell().SetString("price")
	header.AddCell().SetString("quantity")
	header.AddCell().SetString("description")

	count := 0
	for rows.Next() {
		var id int64
		var name, price, description string
		var quantity int

		if err := rows.Scan(&id, &name, &price, &quantity, &description); err != nil {
			log.Fatalf("Error scanning row: %v", err)
		}

		r := sheet.AddRow()
		r.AddCell().SetInt64(id)
		r.AddCell().SetString(name)
		r.AddCell().SetString(price)
		r.AddCell().SetInt(quantity)
		r.AddCell().SetString(description)
		count++
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Rows iteration error: %v", err)
	}

	exportFile := filepath.Join(outPath, "products_export.xlsx")
	if err := file.Save(exportFile); err != nil {
		log.Fatalf("Error saving export file: %v", err)
	}

	fmt.Printf("Exported %d products to %s\n", count, exportFile)
}

func main() {
	mode := flag.String("mode", "import", "Operation mode: import or export")
	flag.Parse()

	_ = godotenv.Load(".env", "../.env", "../../.env")

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	defer db.Close()

	switch strings.ToLower(strings.TrimSpace(*mode)) {
	case "import":
		importProducts(db)
	case "export":
		exportProducts(db)
	default:
		log.Fatalf("invalid mode %q. Use -mode import or -mode export", strconv.Quote(*mode))
	}
}
