//go:build !windows

package filesystem

import "errors"

func openNativeDialogWindow(_ NativeDialogRequest, _ func(NativeDialogResult)) error {
	return errors.New("native dialog request API is only implemented on windows")
}

func processDialogCallbacks() {}

func shutdownNativeDialogs() {}
