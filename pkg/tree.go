package pkg

import (
	"context"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var (
	global_inode uint64 = 1
)

type Tree struct {
	kInfo
}

func (tree *Tree) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir | 0555
	return nil
}

func (tree *Tree) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	//dump:=tree.db.Dump()
	return []fuse.Dirent{
		fuse.Dirent{
			Inode: 2,
			Type:  fuse.DT_Dir,
			Name:  "nodes",
		},
		fuse.Dirent{
			Inode: 3,
			Type:  fuse.DT_Dir,
			Name:  "pods",
		},
		fuse.Dirent{
			Inode: 4,
			Type:  fuse.DT_File,
			Name:  "limits.rc",
		},
	}, nil
}

type Dir struct {
	inode    uint64
	name     string
	children []Node
}

func newDir(name string) *Dir {
	global_inode++
	return &Dir{
		inode:    global_inode,
		name:     name,
		children: nil,
	}
}

func (dir *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = dir.inode
	a.Mode = os.ModeDir | 0555
	return nil
}

func (dir *Dir) Dirent() fuse.Dirent {
	return fuse.Dirent{
		Inode: dir.inode,
		Type:  fuse.DT_Dir,
		Name:  dir.name,
	}
}

func (dir *Dir) Add(node Node) {
	if dir.children == nil {
		dir.children = []Node{}
	}
	dir.children = append(dir.children, node)
}

func (dir *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	for _, child := range dir.children {
		if child.Dirent().Name == name {
			return child, nil
		}
	}
	return nil, syscall.ENOENT
}

func (dir *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	dirDirs := []fuse.Dirent{}
	for _, node := range dir.children {
		dirDirs = append(dirDirs, node.Dirent())
	}
	return dirDirs, nil
}

type File struct {
	inode uint64
	name  string
	read  func() ([]byte, error)
}

func newFile(name string, read func() ([]byte, error)) *File {
	global_inode++

	return &File{
		inode: global_inode,
		name:  name,
		read:  read,
	}
}

func (file *File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = file.inode
	a.Mode = 0555
	bt, _ := file.read()
	a.Size = uint64(len(bt))
	return nil
}

func (file *File) Dirent() fuse.Dirent {
	return fuse.Dirent{
		Inode: file.inode,
		Type:  fuse.DT_File,
		Name:  file.name,
	}
}

func (file *File) ReadAll(ctx context.Context) ([]byte, error) {
	return file.read()
}

type Link struct {
	inode    uint64
	name     string
	redirect string
}

func newLink(name, redirect string) *Link {
	global_inode++
	return &Link{
		inode:    global_inode,
		name:     name,
		redirect: redirect,
	}
}

func (link *Link) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = link.inode
	a.Mode = os.ModeSymlink | 0555
	return nil
}

func (link *Link) Dirent() fuse.Dirent {
	return fuse.Dirent{
		Inode: link.inode,
		Type:  fuse.DT_Link,
		Name:  link.name,
	}
}

func (link *Link) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
	return link.redirect, nil
}

type Node interface {
	fs.Node
	Dirent() fuse.Dirent
}
