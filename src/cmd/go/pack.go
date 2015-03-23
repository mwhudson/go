// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// TODO(mwhudson): this is just copied from cmd/pack/pack.go.  Maybe
// there should be cmd/internal/pack.go?

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	arHeader = "!<arch>\n"
	entryLen = 16 + 12 + 6 + 6 + 8 + 10 + 1 + 1
)

// An Archive represents an open archive file. It is always scanned sequentially
// from start to end, without backing up.
type Archive struct {
	fd    *os.File // Open file descriptor.
	files []string // Explicit list of files to be processed.
}

// archive opens (and if necessary creates) the named archive.
func archive(name string, mode int, files []string) (*Archive, error) {
	fd, err := os.OpenFile(name, mode, 0)
	if err != nil {
		return nil, err
	}
	err = checkHeader(fd)
	if err != nil {
		return nil, err
	}
	return &Archive{
		fd:    fd,
		files: files,
	}, nil
}

// checkHeader verifies the header of the file. It assumes the file
// is positioned at 0 and leaves it positioned at the end of the header.
func checkHeader(fd *os.File) error {
	buf := make([]byte, len(arHeader))
	_, err := io.ReadFull(fd, buf)
	if err != nil {
		return err
	}
	if string(buf) != arHeader {
		return fmt.Errorf("%s is not an archive: bad header", fd.Name())
	}
	return nil
}

// An Entry is the internal representation of the per-file header information of one entry in the archive.
type Entry struct {
	name  string
	mtime int64
	uid   int
	gid   int
	mode  os.FileMode
	size  int64
}

// readMetadata reads and parses the metadata for the next entry in the archive.
func (ar *Archive) readMetadata() *Entry {
	buf := make([]byte, entryLen)
	_, err := io.ReadFull(ar.fd, buf)
	if err == io.EOF {
		// No entries left.
		return nil
	}
	if err != nil || buf[entryLen-2] != '`' || buf[entryLen-1] != '\n' {
		log.Fatal("file is not an archive: bad entry")
	}
	entry := new(Entry)
	entry.name = strings.TrimRight(string(buf[:16]), " ")
	if len(entry.name) == 0 {
		log.Fatal("file is not an archive: bad name")
	}
	buf = buf[16:]
	str := string(buf)
	get := func(width, base, bitsize int) int64 {
		v, err := strconv.ParseInt(strings.TrimRight(str[:width], " "), base, bitsize)
		if err != nil {
			log.Fatal("file is not an archive: bad number in entry: ", err)
		}
		str = str[width:]
		return v
	}
	// %-16s%-12d%-6d%-6d%-8o%-10d`
	entry.mtime = get(12, 10, 64)
	entry.uid = int(get(6, 10, 32))
	entry.gid = int(get(6, 10, 32))
	entry.mode = os.FileMode(get(8, 8, 32))
	entry.size = get(10, 10, 64)
	return entry
}

// scan scans the archive and executes the specified action on each entry.
// When action returns, the file offset is at the start of the next entry.
func (ar *Archive) scan(action func(*Entry)) {
	for {
		entry := ar.readMetadata()
		if entry == nil {
			break
		}
		action(entry)
	}
}

// output copies the entry to the specified writer.
func (ar *Archive) output(entry *Entry, w io.Writer) error {
	n, err := io.Copy(w, io.LimitReader(ar.fd, entry.size))
	if err != nil {
		return err
	}
	if n != entry.size {
		return fmt.Errorf("short file")
	}
	if entry.size&1 == 1 {
		_, err := ar.fd.Seek(1, 1)
		if err != nil {
			return err
		}
	}
	return nil
}

// skip skips the entry without reading it.
func (ar *Archive) skip(entry *Entry) {
	size := entry.size
	if size&1 == 1 {
		size++
	}
	_, err := ar.fd.Seek(size, 1)
	if err != nil {
		log.Fatal(err)
	}
}

// match reports whether the entry matches the argument list.
// If it does, it also drops the file from the to-be-processed list.
func (ar *Archive) match(entry *Entry) bool {
	for i, name := range ar.files {
		if entry.name == name {
			copy(ar.files[i:], ar.files[i+1:])
			ar.files = ar.files[:len(ar.files)-1]
			return true
		}
	}
	return false
}
