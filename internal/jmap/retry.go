package jmap

import (
	"fmt"
	"time"
)

func Retry(
	attempts int,
	sleep time.Duration,
	fn func() error,
) error {

	var err error

	for i := 1; i <= attempts; i++ {

		err = fn()

		if err == nil {
			return nil
		}

		if i == attempts {
			break
		}

		fmt.Printf(
			"Retry %d/%d after error: %v\n",
			i,
			attempts,
			err,
		)

		time.Sleep(sleep)
	}

	return err
}
