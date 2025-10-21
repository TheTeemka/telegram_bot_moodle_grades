package storage

import (
	"encoding/csv"
	"log/slog"
	"os"
	"path/filepath"
)

type CSVwriter struct {
	dir   string
	comma rune
}

func NewCSVWriter(dir string) *CSVwriter {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			slog.Error("failed to create directory", "err", err)
			return nil
		}
	}

	return &CSVwriter{
		dir:   dir,
		comma: ';',
	}
}

func (w *CSVwriter) Write(filename string, records [][]string) error {
	path := filepath.Join(w.dir, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil && err != os.ErrExist {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = w.comma
	for _, record := range records {
		err := writer.Write(record)
		if err != nil {
			slog.Error("error in csvwriting", "err", err)
		}
	}
	writer.Flush()

	return nil
}

func (w *CSVwriter) Read(filename string) ([][]string, error) {
	file, err := os.Open(filepath.Join(w.dir, filename))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = w.comma
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}
