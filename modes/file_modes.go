package modes

import "os"

const (
	// Copying these values from Go's os module for convenience.  This is not, strictly speaking,
	// the most cross-platform-friendly choice.
	O_RDONLY = os.O_RDONLY
	O_WRONLY = os.O_WRONLY
	O_RDWR   = os.O_RDWR
	O_CREATE = os.O_CREATE
	O_APPEND = os.O_APPEND
	O_TRUNC  = os.O_TRUNC
	O_EXCL   = os.O_EXCL
)

func CombineModes(modes ...int) int {
	toReturn := 0
	for _, mode := range modes {
		toReturn = toReturn | mode
	}
	return toReturn
}

func IsWriteAllowed(mode int) bool {
	return checkMode(mode, O_WRONLY) || checkMode(mode, O_RDWR)
}

func IsReadOnly(mode int) bool {
	// Per POSIX, O_RDONLY is zero, so we can't just use (mode & O_RDONLY) == O_RDONLY
	return !IsWriteAllowed(mode)
}

func checkMode(combinedModes int, singleMode int) bool {
	return combinedModes&singleMode == singleMode
}

func IsWriteOnly(mode int) bool {
	return checkMode(mode, O_WRONLY)
}

func IsCreateMode(mode int) bool {
	return checkMode(mode, O_CREATE)
}

func IsAppendMode(mode int) bool {
	return checkMode(mode, O_APPEND)
}

func IsTruncateMode(mode int) bool {
	return IsWriteAllowed(mode) && checkMode(mode, O_TRUNC)
}

func IsExclusiveMode(mode int) bool {
	// O_EXCL is only applicable when O_CREATE is set
	return IsCreateMode(mode) && checkMode(mode, O_EXCL)
}
