package inode

import (
	"fmt"
	"strings"

	"github.com/manderson5192/memfs/filepath"
	"github.com/manderson5192/memfs/fserrors"
	"github.com/manderson5192/memfs/utils"
	"github.com/pkg/errors"
)

type DirectoryInode struct {
	basicInode
	deleted  bool
	contents map[string]Inode
}

func NewRootDirectoryInode() *DirectoryInode {
	rootDirInode := &DirectoryInode{
		contents: map[string]Inode{},
	}
	rootDirInode.contents[filepath.SelfDirectoryEntry] = rootDirInode
	rootDirInode.contents[filepath.ParentDirectoryEntry] = rootDirInode
	return rootDirInode
}

func NewDirectoryInode(parent *DirectoryInode) *DirectoryInode {
	newDirInode := &DirectoryInode{
		contents: map[string]Inode{},
	}
	newDirInode.contents[filepath.SelfDirectoryEntry] = newDirInode
	newDirInode.contents[filepath.ParentDirectoryEntry] = parent
	return newDirInode
}

func (i *DirectoryInode) InodeType() InodeType {
	return InodeDirectory
}

func (i *DirectoryInode) Size() int {
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	numEntries := 0
	for name := range i.contents {
		if name == filepath.SelfDirectoryEntry || name == filepath.ParentDirectoryEntry {
			continue
		}
		numEntries++
	}
	return numEntries
}

// Parent obtains the DirectoryInode that is parent to this DirectoryInode
func (i *DirectoryInode) Parent() *DirectoryInode {
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	parentInode, parentExists := i.contents[filepath.ParentDirectoryEntry]
	if !parentExists {
		// This shouldn't happen, so we panic on the condition
		panic("parent entry for directory inode does not exist")
	}
	parentDirectoryInode, castOk := parentInode.(*DirectoryInode)
	if !castOk {
		// This also shouldn't happen
		panic("parent directory cannot cast to directory inode type")
	}
	return parentDirectoryInode
}

// ReverseLookupEntry returns the entry name for the specified child DirectoryInode, or an error
// if it is unable to do so
func (i *DirectoryInode) ReverseLookupEntry(child *DirectoryInode) (string, error) {
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	for entry, inode := range i.contents {
		// Ignore self and parent directory references
		if entry == filepath.SelfDirectoryEntry || entry == filepath.ParentDirectoryEntry {
			continue
		}
		// Ignore non-directory inodes
		if !IsDirectory(inode) {
			continue
		}
		if child == inode {
			return entry, nil
		}
	}
	return "", errors.Wrapf(fserrors.ENoEnt, "entry for directory inode was not found")
}

// IsRootDirectoryInode returns whether this DirectoryInode corresponds to the filesystem's root
func (i *DirectoryInode) IsRootDirectoryInode() bool {
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	parent := i.Parent()
	return i == parent
}

// AddDirectory adds (and returns) a DirectoryInode for a direct child directory named 'name'.  It
// cannot create an entry containing a path separator and it cannot create a subdirectory that
// already exists
func (i *DirectoryInode) AddDirectory(name string) (*DirectoryInode, error) {
	// Check that this directory entry doesn't contain the path separator
	if strings.Contains(name, filepath.PathSeparator) {
		return nil, errors.Wrapf(fserrors.EInval, "cannot add subdirectory inode for a name containing path separator %s: %s", filepath.PathSeparator, name)
	}
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()
	// Disallow adding subdirectories on directories that have already been marked as deleted
	if i.deleted {
		return nil, errors.Wrapf(fserrors.ENoEnt, "cannot add entries to a directory marked for deletion")
	}
	// Make sure that the entry doesn't already exist
	if _, exists := i.contents[name]; exists {
		return nil, errors.Wrapf(fserrors.EExist, "directory entry '%s' already exists", name)
	}
	subdirInode := NewDirectoryInode(i)
	i.contents[name] = subdirInode
	return subdirInode, nil
}

type onExistFunc func(child Inode, name string) (Inode, error)
type onNoExistFunc func(parent *DirectoryInode, name string) (Inode, error)

// getInodeEntry is a convenience method that provides common functionality for getting entry's
// inode from the receiver DirectoryInode `i`.  It also supports running arbitrary logic when entry
// is or is not found in i's entry table.
//
// This function is **not thread safe**.  It should be invoked by a caller holding a Read-level lock
// on i's rwMutex, or a Write-level lock if onExist or onNoExistFunc will mutate i's state.
func (i *DirectoryInode) getInodeEntry(entry string, onExist onExistFunc, onNoExist onNoExistFunc) (Inode, error) {
	// Check that this directory entry doesn't contain the path separator
	if strings.Contains(entry, filepath.PathSeparator) {
		return nil, errors.Wrapf(fserrors.EInval, "entry %s contains illegal character %s", entry, filepath.PathSeparator)
	}
	inode, exists := i.contents[entry]
	if !exists {
		if onNoExist == nil {
			return nil, errors.Wrapf(fserrors.ENoEnt, "entry '%s' does not exist", entry)
		} else {
			return onNoExist(i, entry)
		}
	} else {
		if onExist == nil {
			return inode, nil
		} else {
			return onExist(inode, entry)
		}
	}
}

// InodeEntry holds a Read-level lock on the DirectoryInode and returns the uncasted Inode for the
// provided entry name, or an error.
func (i *DirectoryInode) InodeEntry(entry string) (Inode, error) {
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	return i.getInodeEntry(entry, nil, nil)
}

// DirectoryInodeEntry obtains the Inode corresponding to the named entry, or an error
func (i *DirectoryInode) DirectoryInodeEntry(entry string) (*DirectoryInode, error) {
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	inode, err := i.getInodeEntry(entry, nil, nil)
	if err != nil {
		return nil, err
	}
	dirInode, ok := inode.(*DirectoryInode)
	if !ok {
		return nil, errors.Wrapf(fserrors.ENotDir, "entry '%s' is not a directory", entry)
	}
	// Deny access to DirectoryInodes after they have been marked as deleted.  This case should be
	// rare, but is technically possible
	if dirInode.isDeleted() {
		return nil, errors.Wrapf(fserrors.ENoEnt, "entry '%s' does not exist", entry)
	}
	return dirInode, nil
}

// FileInodeEntry obtains the Inode corresponding to the named entry, or an error
func (i *DirectoryInode) FileInodeEntry(entry string) (*FileInode, error) {
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	inode, err := i.getInodeEntry(entry, nil, nil)
	if err != nil {
		return nil, err
	}
	fileInode, ok := inode.(*FileInode)
	if !ok {
		return nil, errors.Wrapf(fserrors.EIsDir, "entry '%s' is not a file", entry)
	}
	return fileInode, nil
}

// CreateFileInodeEntry will return a FileInode for i.contents[entry], either by looking up and
// casting an existing inode, or by creating a new one altogether.  However, if errOnExist is true,
// then CreateFileInodeEntry will return EEXIST is i.contents[entry] already exists.
func (i *DirectoryInode) CreateFileInodeEntry(entry string, errOnExist bool) (*FileInode, error) {
	// Check that entry doesn't contain the path separator
	if strings.Contains(entry, filepath.PathSeparator) {
		return nil, errors.Wrapf(fserrors.EInval, "name '%s' contains a path separator", entry)
	}
	// Take an exclusive lock in case we end up creating a file
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()
	onExist := func(inode Inode, name string) (Inode, error) {
		if errOnExist {
			return nil, errors.Wrapf(fserrors.EExist, "file '%s' already exists", name)
		} else {
			return inode, nil
		}
	}
	onNoExist := func(dirInode *DirectoryInode, name string) (Inode, error) {
		if dirInode.deleted {
			return nil, errors.Wrapf(fserrors.ENoEnt, "cannot add entries to a directory marked for deletion")
		}
		newFileInode := NewFileInode()
		dirInode.contents[name] = newFileInode
		return newFileInode, nil
	}
	inode, err := i.getInodeEntry(entry, onExist, onNoExist)
	if err != nil {
		return nil, err
	}
	fileInode, ok := inode.(*FileInode)
	if !ok {
		return nil, errors.Wrapf(fserrors.EIsDir, "entry '%s' is not a file", entry)
	}
	return fileInode, nil
}

// InodeEntry represents basic information about an entry in a DirectoryInode's entry table
type InodeEntry struct {
	Name string
	Type InodeType
}

func (i *DirectoryInode) InodeEntries() []InodeEntry {
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	toReturn := make([]InodeEntry, 0, len(i.contents))
	for entryName, inode := range i.contents {
		if entryName == filepath.SelfDirectoryEntry || entryName == filepath.ParentDirectoryEntry {
			continue
		}
		toReturn = append(toReturn, InodeEntry{
			Name: entryName,
			Type: inode.InodeType(),
		})
	}
	return toReturn
}

// LookupSubdirectory will return a DirectoryInode for the specified subdirectory relative to this
// DirectoryInode.  It assumes that subdirectory is a relative path, even if it begins with a path
// separator character.  If the specified subdirectory can't be found, or if any named directory
// entry along its path is not a directory (e.g. if it is a file), then it will return an error.  If
// subdirectory is the empty string, then the receiver DirectoryInode will be returned.
func (i *DirectoryInode) LookupSubdirectory(subdirectory string) (*DirectoryInode, error) {
	if subdirectory == "" {
		return i, nil
	}
	if !filepath.IsRelativePath(subdirectory) {
		return nil, errors.Wrapf(fserrors.EInval, "'%s' is not a relative path", subdirectory)
	}
	currentDirInode := i
	currentSubdirectory := subdirectory
	for len(currentSubdirectory) > 0 {
		// Parse a directory entry from the beginning of currentSubdirectory
		currentSubdirectory = strings.TrimLeft(currentSubdirectory, filepath.PathSeparator)
		entryName, remainder, _ := utils.Cut(currentSubdirectory, filepath.PathSeparator)
		// Get the directory inode for this entry
		dirInode, getEntryErr := currentDirInode.DirectoryInodeEntry(entryName)
		if getEntryErr != nil {
			return nil, errors.Wrapf(getEntryErr, "cannot find subdirectory '%s'", subdirectory)
		}
		// iterate
		currentDirInode = dirInode
		currentSubdirectory = remainder
	}
	return currentDirInode, nil
}

// delete marks this DirectoryInode as deleted.  It will only succeed if this directory is empty.
func (i *DirectoryInode) delete() error {
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()
	// Check: is the directory already deleted?
	if i.deleted {
		return nil
	}
	// Check: is the directory empty?
	for entry := range i.contents {
		if entry == filepath.SelfDirectoryEntry || entry == filepath.ParentDirectoryEntry {
			continue
		}
		return errors.Wrapf(fserrors.ENotEmpty, "directory is not empty")
	}
	// mark as deleted
	i.deleted = true
	return nil
}

func (i *DirectoryInode) isDeleted() bool {
	i.rwMutex.RLock()
	defer i.rwMutex.RUnlock()
	return i.deleted
}

// doDeleteDirectory is a convenience method that provides common functionality for deleting a child
// DirectoryInode from `i` that is currently under the entry name `entry`.
//
// This function is **not thread safe**.  It should only be invoked when a Write-level lock is held
// on the DirectoryInode.
func (i *DirectoryInode) doDeleteDirectory(entry string) error {
	// Check: disallow removing the special "." and ".." directories
	if entry == "." || entry == ".." {
		return errors.Wrapf(fserrors.EInval, "refusing to remove '.' or '..' directory: skipping '%s", entry)
	}
	// Get the DirectoryInode for entry
	inode, exists := i.contents[entry]
	if !exists {
		return errors.Wrapf(fserrors.ENoEnt, "entry '%s' does not exist", entry)
	}
	dirInode, ok := inode.(*DirectoryInode)
	if !ok {
		return errors.Wrapf(fserrors.ENotDir, "entry '%s' is not a directory", entry)
	}
	// Make sure we can successfully delete entry's directory
	if err := dirInode.delete(); err != nil {
		return errors.Wrapf(err, "failed to delete directory entry '%s'", entry)
	}
	// Finally, remove the entry
	delete(i.contents, entry)
	return nil
}

func (i *DirectoryInode) DeleteDirectory(entry string) error {
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()
	return i.doDeleteDirectory(entry)
}

// doDeleteFile is a convenience method that provides common functionality for deleting a child
// FileInode from `i` that is currently under the entry name `entry`
//
// This function is **not thread safe**.  It should only be invoked when a Write-level lock is held
// on the DirectoryInode
func (i *DirectoryInode) doDeleteFile(entry string) error {
	// Get the FileInode for entry
	inode, exists := i.contents[entry]
	if !exists {
		return errors.Wrapf(fserrors.ENoEnt, "entry '%s' does not exist", entry)
	}
	_, ok := inode.(*FileInode)
	if !ok {
		return errors.Wrapf(fserrors.EIsDir, "entry '%s' is not a file", entry)
	}
	// Remove the entry
	delete(i.contents, entry)
	return nil
}

func (i *DirectoryInode) DeleteFile(entry string) error {
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()
	return i.doDeleteFile(entry)
}

func (i *DirectoryInode) SetParent(parent *DirectoryInode) {
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()
	i.contents[filepath.ParentDirectoryEntry] = parent
}

// MoveEntry will relocate the inode specified by src that is currently a child of srcParentInode
// to the entry specified by dst that will be a child of dstParentInode
func MoveEntry(srcParentInode, dstParentInode *DirectoryInode, src, dst *filepath.PathInfo) error {
	// Check that srcEntry is not the special self or parent directory entries
	if src.Entry == filepath.SelfDirectoryEntry || src.Entry == filepath.ParentDirectoryEntry {
		return errors.Wrapf(fserrors.EInval, "cannot move '.' or '..' entries")
	}
	// Check the same for dstEntry
	if dst.Entry == filepath.SelfDirectoryEntry || dst.Entry == filepath.ParentDirectoryEntry {
		return errors.Wrapf(fserrors.EInval, "cannot overwrite '.' or '..' entries")
	}
	// Check that the dst entry name doesn't contain the path separator
	if strings.Contains(dst.Entry, filepath.PathSeparator) {
		return errors.Wrapf(fserrors.EInval, "entry name '%s' contains the path separator", dst.Entry)
	}
	// Edge case: srcParentInode and dstParentInode are the same.  That requires a different locking
	// discipline, so we special-case it
	if srcParentInode == dstParentInode {
		return srcParentInode.renameEntry(src, dst)
	}
	srcParentInode.rwMutex.Lock()
	defer srcParentInode.rwMutex.Unlock()
	dstParentInode.rwMutex.Lock()
	defer dstParentInode.rwMutex.Unlock()
	// Disallow adding files to directories that have already been marked as deleted
	if dstParentInode.deleted {
		return errors.Wrapf(fserrors.ENoEnt, "cannot add entries to a directory marked for deletion")
	}
	// Get the inode for the srcEntry
	srcInode, exists := srcParentInode.contents[src.Entry]
	if !exists {
		return errors.Wrapf(fserrors.ENoEnt, "source entry '%s' does not exist", src.Entry)
	}
	if srcInode.InodeType() == InodeFile && src.MustBeDir {
		// src ended with a separator, so it ought to be a directory, but we found a file.
		return errors.Wrapf(fserrors.ENotDir, "src entry is a file but name references a directory")
	}
	if srcInode.InodeType() == InodeFile && dst.MustBeDir {
		// dst ended with a separator, so it ought to be a directory, but src is a file
		return errors.Wrapf(fserrors.ENotDir, "dst's name references a directory but src is a file")
	}
	// Insert the inode into its new location
	switch srcInodeTyped := srcInode.(type) {
	case *FileInode:
		if err := dstParentInode.doInsertFileInode(dst.Entry, srcInodeTyped); err != nil {
			return err
		}
	case *DirectoryInode:
		if err := dstParentInode.doInsertDirectoryInode(dst.Entry, srcInodeTyped); err != nil {
			return err
		}
	default:
		return fmt.Errorf("source entry '%s' has malformed inode of type '%s'", src.Entry, srcInode.InodeType().String())
	}
	// Remove the inode from its old location
	delete(srcParentInode.contents, src.Entry)
	return nil
}

// renameEntry is a special case implementation of MoveEntry where src and dst are both children
// of a single DirectoryInode `i`
func (i *DirectoryInode) renameEntry(src, dst *filepath.PathInfo) error {
	// Special case: do nothing
	if src.Entry == dst.Entry {
		return nil
	}
	i.rwMutex.Lock()
	defer i.rwMutex.Unlock()
	// Disallow moving files in directories that have already been marked as deleted
	if i.deleted {
		return errors.Wrapf(fserrors.ENoEnt, "cannot move entries in a directory marked for deletion")
	}
	// Get the inode to be moved
	inode, exists := i.contents[src.Entry]
	if !exists {
		return fmt.Errorf("source entry '%s' does not exist", src.Entry)
	}
	if inode.InodeType() == InodeFile && src.MustBeDir {
		// src ended with a separator, so it ought to be a directory, but we found a file.
		return errors.Wrapf(fserrors.ENotDir, "src entry is a file but name references a directory")
	}
	if inode.InodeType() == InodeFile && dst.MustBeDir {
		// dst ended with a separator, so it ought to be a directory, but src is a file
		return errors.Wrapf(fserrors.ENotDir, "dst's name references a directory but src is a file")
	}
	switch inodeTyped := inode.(type) {
	case *FileInode:
		if err := i.doInsertFileInode(dst.Entry, inodeTyped); err != nil {
			return err
		}
	case *DirectoryInode:
		if err := i.doInsertDirectoryInode(dst.Entry, inodeTyped); err != nil {
			return err
		}
	default:
		return fmt.Errorf("source entry '%s' has malformed inode of type '%s'", src.Entry, inodeTyped.InodeType().String())
	}
	delete(i.contents, src.Entry)
	return nil
}

// doInsertFileInode is a convenience method that provides common functionality for inserting
// FileInode `newEntry` into i's entry table under the entry name `entry`.  If an entry by this name
// already exists, then this method will delete that inode.
//
// This function is **not thread safe**.  It should only be invoked when a Write-level lock is held
// on the DirectoryInode
func (i *DirectoryInode) doInsertFileInode(entry string, newEntry *FileInode) error {
	// if an entry by this name already exists, then we are meant to delete it
	if oldEntry, exists := i.contents[entry]; exists {
		switch oldEntry.(type) {
		case *FileInode:
			if err := i.doDeleteFile(entry); err != nil {
				return errors.Wrapf(err, "failed to delete existing file")
			}
		case *DirectoryInode:
			if err := i.doDeleteDirectory(entry); err != nil {
				return errors.Wrapf(err, "failed to delete existing directory")
			}
		default:
			return fmt.Errorf("existing entry '%s' has malformed inode of type '%s'", entry, oldEntry.InodeType().String())
		}
	}
	i.contents[entry] = newEntry
	return nil
}

// doInsertDirectoryInode is a convenience method that provides common functionality for inserting
// DirectoryInode `newEntry` into i's entry table under the entry name `entry`.  If an entry by this
// name already exists, then this method will delete that inode.
//
// This function is **not thread safe**.  It should only be invoked when a Write-level lock is held
// on the DirectoryInode
func (i *DirectoryInode) doInsertDirectoryInode(entry string, newEntry *DirectoryInode) error {
	// if an entry by this name already exists, then we are meant to delete it
	if oldEntry, exists := i.contents[entry]; exists {
		switch oldEntry.(type) {
		case *FileInode:
			// Interestingly, the POSIX spec says that rename(2) should return an error (EISDIR)
			// if the source ("old") path specifies a directory but the destination ("new") path
			// coincides with a file.  We could do that here, but it doesn't seem strictly
			// necessary, so we will allow it.
			if err := i.doDeleteFile(entry); err != nil {
				return errors.Wrapf(err, "failed to delete existing file")
			}
		case *DirectoryInode:
			if err := i.doDeleteDirectory(entry); err != nil {
				return errors.Wrapf(err, "failed to delete existing directory")
			}
		default:
			return fmt.Errorf("existing entry '%s' has malformed inode of type '%s'", entry, oldEntry.InodeType().String())
		}
	}
	// insert the entry into this directory
	i.contents[entry] = newEntry
	// update the newEntry inode's parent pointer to point to this inode
	newEntry.SetParent(i)
	return nil
}
