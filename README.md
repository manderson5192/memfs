# MemFS: In-Memory Fileystem

## Prerequisites

This code is written in Go and was developed with Go 1.17.6.  If you don't have Go installed
already, then please install it from here: https://go.dev/doc/install

## Developing, Building, Testing, Running, etc.

For your convenience, this repo includes a Makefile.  It enables you to:
* Test with `make test`
* Build `main.go` to `./build/main` with `make build`
* Run `main.go` (with building `./build/main`) with `make run`

Development is best done using VSCode with the Go extension (search for `golang.go` and install).

## Design

TODO

## Specifications Met
### Basic requirements
- [x] Change the current working directory: `process.WorkingDirectory()`
- [x] Get the current working directory: `process.ChangeDirectory()`
- [x] Create a new directory: `process.MakeDirectory()`
- [x] Get the directory contents: `process.ListDirectory()`
- [x] Remove a directory: `process.RemoveDirectory()`
- [x] Create a new file: `process.CreateFile()`
- [x] Write file contents: `file.TruncateAndWriteAll()`, `file.Write()`, `file.WriteAt()`
- [x] Get file contents: `file.ReadAll()`, `file.Read()`, `file.ReadAt()`
- [x] Move a file: `process.Rename()`
- [x] Find a file/directory: `process.FindAll()`

### Extensions

* Move and copy
    - [x] You can move ~~or copy~~ files and directories (note: a `cp` command can be implemented with the provided syscalls)
    - [ ] Support merging the contents of two directories when moving or copying one into
the other.
    - [ ] Handle name collisions in some way (e.g. auto renaming files, merging
directories.)
* Operations on paths:
    - [x] When doing basic operations (changing the current working directory, creating or moving files or folders, etc), you can use absolute paths instead of only operating on objects in the current working directory.
    - [x] You can use relative paths (relative to the current working directory) as well, including the special “..” path that refers to the parent directory.
    - [x] When creating directories, you can choose to automatically create any intermediate directories on the path that don’t exist yet.
        * NOTE: I opted not to implement creating intermediate directories during file creation or rename() operations because these are not operations that are commonly performed (or supported) on *Nix systems
* Walk a subtree
    - [x] You can walk through all the recursive contents of a directory, invoking a passed-in function on each child directory/file: `process.Walk()`
    - [x] While walking, the passed-in function can arbitrarily choose not to recurse into certain subdirectories: `process.Walk()`
    - [x] Use this to implement some recursive operation. For example, finding the first file in a subtree whose name matches a regex: `process.FindFirstMatchingFile()`
* Stream for file contents
    - [x] Reading a file can be done in chunks or as a stream, not just all contents at once: `file.Read()` and `file.Write()`
    - [x] You can also write to a file in chunks or a stream (Go doesn't have the concept of streams, but it does have the generic io.Reader and io.Writer interfaces, which `file.File` does implement such that `ioutil` methods can be employed)
    - [x] You can have one writer and multiple readers of a file at the same time (implemented with a r/w lock on each file inode)
    - [x] You can continue reading/writing from a file even if it gets moved to a different path before you’re done: (there's a test for this)
    - [x] You can read and write starting from any part of a file and also jump to a different part (random access): `file.Seek()`, `file.ReadAt()`, `file.WriteAt()`
