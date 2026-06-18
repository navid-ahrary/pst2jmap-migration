package csv

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/navid/pst2jmap-migration/internal/model"
)

func ReadJobs(path string) ([]model.Job, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)

	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("csv contains no jobs")
	}

	header := rows[0]

	expected := []string{
		"email",
		"password",
		"pstfile",
	}

	for i, v := range expected {
		if header[i] != v {
			return nil, fmt.Errorf(
				"invalid header, expected: email,password,pstfile",
			)
		}
	}

	jobs := make([]model.Job, 0, len(rows)-1)

	for i, row := range rows[1:] {

		if len(row) < 3 {
			return nil, fmt.Errorf(
				"invalid row %d: expected 3 columns",
				i+2,
			)
		}

		jobs = append(jobs, model.Job{
			Email:    row[0],
			Password: row[1],
			PSTFile:  row[2],
			Row:      i + 2,
		})
	}

	return jobs, nil
}
