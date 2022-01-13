package fserrors

import "fmt"

// These error constants are used throughout MemFS so that users can examine arbitrarily-wrapped
// errors to determine _why_ their call failed and not just _whether_ it did.  Users can employ Go's
// errors.Is() function to determine whether an error is a descendant of one of these errors
var (
	EExist    = fmt.Errorf("file exists")
	ENoEnt    = fmt.Errorf("file does not exist")
	EIsDir    = fmt.Errorf("target is a directory")
	ENotDir   = fmt.Errorf("target is not a directory")
	EInval    = fmt.Errorf("invalid argument")
	ENoSpace  = fmt.Errorf("no space")
	ENotEmpty = fmt.Errorf("not empty")
)
