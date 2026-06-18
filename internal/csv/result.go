// internal/csv/result.go

package csv

import (
	"encoding/csv"
	"os"
	"strconv"

	"github.com/navid/pst2jmap-migration/internal/model"
)

func WriteResults(
	path string,
	results []model.MigrationResult,
) error {

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	err = w.Write([]string{
		"row",
		"email",
		"pstfile",
		"status",
		"error",
	})

	if err != nil {
		return err
	}

	for _, r := range results {

		err := w.Write([]string{
			strconv.Itoa(r.Row),
			r.Email,
			r.PSTFile,
			r.Status,
			r.Error,
		})

		if err != nil {
			return err
		}
	}

	return nil
}
