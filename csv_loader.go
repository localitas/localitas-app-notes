package notes

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func LoadCSVVariable(dataDir, varName, filePath, blockID string) (*Variable, error) {
	fullPath := filepath.Join(dataDir, filePath)

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV must have header + at least one data row")
	}

	headers := records[0]
	data := records[1:]

	schema := &DataSchema{
		Columns:   make([]*ColumnSchema, len(headers)),
		Rows:      len(data),
		SizeBytes: 0,
	}

	for colIdx, header := range headers {
		col := &ColumnSchema{
			Name:         header,
			SampleValues: make([]interface{}, 0, min(3, len(data))),
		}

		isNumeric := true
		isFloat := false
		hasNull := false
		var numericValues []float64

		for rowIdx, row := range data {
			if colIdx >= len(row) {
				hasNull = true
				continue
			}
			val := row[colIdx]

			if rowIdx < 3 {
				col.SampleValues = append(col.SampleValues, val)
			}

			if val == "" {
				hasNull = true
				continue
			}

			if f, err := strconv.ParseFloat(val, 64); err == nil {
				numericValues = append(numericValues, f)
				if _, intErr := strconv.ParseInt(val, 10, 64); intErr != nil {
					isFloat = true
				}
			} else {
				isNumeric = false
			}
		}

		col.Nullable = hasNull

		if isNumeric && len(numericValues) > 0 {
			if isFloat {
				col.DType = "float64"
			} else {
				col.DType = "int64"
			}

			minVal := numericValues[0]
			maxVal := numericValues[0]
			sum := 0.0
			for _, v := range numericValues {
				if v < minVal {
					minVal = v
				}
				if v > maxVal {
					maxVal = v
				}
				sum += v
			}
			col.Min = minVal
			col.Max = maxVal
			col.Mean = math.Round(sum/float64(len(numericValues))*100) / 100

			if !isFloat {
				col.Sum = int64(sum)
			}
		} else {
			col.DType = "string"
		}

		schema.Columns[colIdx] = col
	}

	fi, _ := file.Stat()
	if fi != nil {
		schema.SizeBytes = fi.Size()
	}

	return &Variable{
		Name:        varName,
		Type:        "dataframe",
		Source:      fmt.Sprintf("csv:%s", filePath),
		Schema:      schema,
		DefinedIn:   blockID,
		LastUpdated: time.Now().UTC(),
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
