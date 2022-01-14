package main

import (
	"encoding/json"
	"log"

	"github.com/manderson5192/memfs/filesys"
	"github.com/manderson5192/memfs/process"
)

func main() {
	// Create an empty filesystem
	fs := filesys.NewFileSystem()

	// Create a "process".  This encapsulates a working directory (initialized at /) and provides
	// an interface for running actions on the filesystem
	p := process.NewProcessFilesystemContext(fs)

	// Perform a simple example workflow
	err := p.MakeDirectory("school")
	handleError(err)

	err = p.ChangeDirectory("school")
	handleError(err)

	workdir, err := p.WorkingDirectory()
	handleError(err)
	log.Printf("working directory: %s", workdir)

	err = p.MakeDirectory("homework")
	handleError(err)

	err = p.ChangeDirectory("homework")
	handleError(err)

	for _, dirname := range []string{"math", "lunch", "history", "spanish"} {
		err = p.MakeDirectory(dirname)
		handleError(err)
	}

	err = p.RemoveDirectory("lunch")
	handleError(err)

	directoryContents, err := p.ListDirectory(".")
	handleError(err)
	j, err := json.Marshal(directoryContents)
	handleError(err)
	log.Printf("directory contents: %s", string(j))

	workdir, err = p.WorkingDirectory()
	handleError(err)
	log.Printf("working directory: %s", workdir)

	err = p.ChangeDirectory("..")
	handleError(err)

	err = p.MakeDirectory("cheatsheet")

	directoryContents, err = p.ListDirectory(".")
	handleError(err)
	j, err = json.Marshal(directoryContents)
	handleError(err)
	log.Printf("directory contents: %s", string(j))

	err = p.RemoveDirectory("cheatsheet")
	handleError(err)

	err = p.ChangeDirectory("..")
	handleError(err)

	workdir, err = p.WorkingDirectory()
	handleError(err)
	log.Printf("working directory: %s", workdir)
}

// In many cases we just want to print the error and exit when an error is encountered, so we
// provide this (otherwise sloppy) method for convenience
func handleError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
