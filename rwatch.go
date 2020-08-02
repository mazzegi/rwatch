package rwatch

import (
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

type RecursiveWatcher struct {
	root     string
	watcher  *fsnotify.Watcher
	Messages chan Message
	doneC    chan struct{}
	subDirs  map[string]struct{}
}

func NewRecursiveWatcher(root string) (*RecursiveWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(root) {
		root, err = filepath.Abs(root)
		if err != nil {
			return nil, err
		}
	}
	rw := &RecursiveWatcher{
		root:     root,
		watcher:  w,
		Messages: make(chan Message),
		doneC:    make(chan struct{}),
		subDirs:  map[string]struct{}{},
	}
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			err = w.Add(path)
			if err != nil {
				return err
			}
			rw.subDirs[path] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	go rw.run()
	return rw, nil
}

func (rw *RecursiveWatcher) Close() {
	rw.watcher.Close()
	close(rw.Messages)
	<-rw.doneC
}

func (rw *RecursiveWatcher) run() {
	defer close(rw.doneC)
	for {
		select {
		case e, ok := <-rw.watcher.Events:
			if !ok {
				return
			}
			rw.handleMessage(e)
		case err, ok := <-rw.watcher.Errors:
			if !ok {
				return
			}
			rw.handleError(err, "")
		}
	}
}

func (rw *RecursiveWatcher) handleError(err error, path string) {
	rw.Messages <- Error{
		Path:  path,
		Error: err,
	}
}

func (rw *RecursiveWatcher) handleMessage(e fsnotify.Event) {
	if e.Op == fsnotify.Remove {
		if _, contains := rw.subDirs[e.Name]; contains {
			delete(rw.subDirs, e.Name)
			rw.Messages <- Deleted{
				Path: e.Name,
			}
		}
		return
	}
	fi, err := os.Stat(e.Name)
	if err != nil {
		rw.handleError(err, e.Name)
		return
	}
	switch e.Op {
	case fsnotify.Create:
		rw.Messages <- Created{
			Path: e.Name,
			File: fi,
		}
		if fi.IsDir() {
			err := rw.watcher.Add(e.Name)
			if err != nil {
				rw.handleError(err, e.Name)
			} else {
				rw.subDirs[e.Name] = struct{}{}
			}
		}
	case fsnotify.Write:
		rw.Messages <- Changed{
			Path: e.Name,
			File: fi,
		}
	case fsnotify.Rename:
		rw.Messages <- Renamed{
			Path: e.Name,
			File: fi,
		}
	case fsnotify.Chmod:
		rw.Messages <- Chmoded{
			Path: e.Name,
			File: fi,
		}
	default:
		rw.handleError(errors.Errorf("unknown op (%s)", e.Op), "")
	}
}
