package validate

import "errors"

type (
	Valid interface{ Valid() error }
)

func All[T Valid](items ...T) error {
	var errs []error
	for _, i := range items {
		if err := i.Valid(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
