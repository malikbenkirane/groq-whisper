package dcheck

import "fmt"

func Wrap(errDefer, err error, msg string, args ...any) error {
	if errDefer != nil && err != nil {
		return fmt.Errorf("%w then %s: %w",
			err, fmt.Sprintf(msg, args...), errDefer)
	}
	if errDefer != nil {
		return fmt.Errorf("%s: %w", msg, errDefer)
	}
	return err
}
