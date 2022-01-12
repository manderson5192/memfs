package fserrors

import "fmt"

var (
	EExist    = fmt.Errorf("file exists")
	ENoEnt    = fmt.Errorf("file does not exist")
	EIsDir    = fmt.Errorf("target is a directory")
	ENotDir   = fmt.Errorf("target is not a directory")
	EInval    = fmt.Errorf("invalid argument")
	ENoSpace  = fmt.Errorf("no space")
	ENotEmpty = fmt.Errorf("not empty")
)
