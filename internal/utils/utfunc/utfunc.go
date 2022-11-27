package utfunc

import (
	"fmt"

	"repo-scanner/internal/utils/serror"
)

// Try function
func Try(fn func() serror.SError) (errx serror.SError) {
	if fn == nil {
		return errx
	}

	func() {
		defer func() {
			if errx != nil {
				return
			}

			if errRcv := recover(); errRcv != nil {
				errx = serror.Newsc(1, fmt.Sprintf("%+v", errRcv), "Unexpected exception has occurred")
			}
		}()

		errx = fn()
	}()

	return errx
}
