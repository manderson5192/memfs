# MemFS: In-Memory Fileystem

## Prerequisites

This code is written in Go and was developed with Go 1.17.6.  If you don't have Go installed
already, then please install it from here: https://go.dev/doc/install

## Quick Start

Take a look at [main.go](main.go) and run with `make run`.  Take a look at `ProcessFilesystemContext` in [process.go](process/process.go) to see the avaialble function calls and craft your own program :).

## Developing, Building, Testing, Running, etc.

For your convenience, this repo includes a Makefile.  It enables you to:
* Test with `make test`
* Build `main.go` to `./build/main` with `make build`
* Run `main.go` (with building `./build/main`) with `make run`

Development is best done using VSCode with the Go extension (search for `golang.go` and install).

## Design

MemFS is an in-memory filesystem built around a tree of `inode.DirectoryInode`
structs that have `inode.DirectoryInode` and `inode.FileInode` children.  The [inode/](inode/) package has the
code that implements low-level operations on this tree.

Since MemFS is an in-memory filesystem, and since it
is implemented in Go, it does not have its own allocator for carving inodes out of anonymous blocks of memory
(all on-disk filesystems need to be able to allocate/deallocate raw on-disk block storage).  Instead, MemFS leans on Go's heap
allocation and garbage collection.  This is convenient from an implementation standpoint, but also has some neat properties.
For example, if a file is logically deleted by one goroutine (a simulated Linux "process") while another process has it open,
that other process is able to continue reading from/writing to the file as long as it wants to, just like in Linux.
The file's underlying memory won't be cleaned up until the second process deallocates its reference to the file (i.e. "closes" it).  At that point, Go's garbage collector will clean up the now-unused memory backing that file.

Notably, every MemFS inode has a `sync.RWMutex` that is held at Read
or Write levels, depending on the low-level inode operation.  In general, a Read level lock is held while looking up subdirectories
or files, while reading file data, or while examining directory entries.  A Write level lock is held while file or directory contents are mutated.

Two packages are implemented on top of the [inode/](inode/) package: [file/](file/) and [directory/](directory/).  Their interfaces
(`file.File` and `directory.Directory`) are implemented by package-private structs `file.file` and `directory.directory`.  These structs each encapsulate a reference to an inode (`file` also contains an offset, a mode, and a `sync.Mutex` (to synchronize access to the file offset)).  `file.File` and `directory.Directory` are semantically equivalent to a Linux file descriptors: they represent handles used by a single process, and are a layer of indirection on top of the inodes that actually implement underlying storage.

Finally, on top of [file/](file/) and [directory/](directory/) is the [process/](process/) package that exports and implements the
`process.ProcessFilesystemContext` interface.  `process.ProcessFilesystemContext` is implemented by `process.processContext`, a package-private struct that encapsulates a `filesys.FileSystem` (this interface just provides a reference to the root directory's `directory.Directory`) and a `directory.Directory` for the current working directory.

## Features
### Basic Capabilities
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

### Extended Capabilities

* Move and copy
    - [x] You can move ~~or copy~~ files and directories
        * NOTE: a `cp` command can be implemented with the provided syscalls
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
