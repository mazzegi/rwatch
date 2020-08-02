package rwatch

import (
	"fmt"
	"os"
)

type Message interface{}

type Error struct {
	Path  string
	Error error
}

type Created struct {
	Path string
	File os.FileInfo
}

type Changed struct {
	Path string
	File os.FileInfo
}

type Deleted struct {
	Path string
}

type Renamed struct {
	Path string
	File os.FileInfo
}

type Chmoded struct {
	Path string
	File os.FileInfo
}

func (m Error) String() string {
	return fmt.Sprintf("ERROR (%s): %v", m.Path, m.Error)
}

func (m Created) String() string {
	return fmt.Sprintf("CREATED (%s) (dir=%t)", m.Path, m.File.IsDir())
}

func (m Changed) String() string {
	return fmt.Sprintf("CHANGED (%s) (dir=%t)", m.Path, m.File.IsDir())
}

func (m Deleted) String() string {
	return fmt.Sprintf("DELETED (%s)", m.Path)
}

func (m Renamed) String() string {
	return fmt.Sprintf("RENAMED (%s) (dir=%t)", m.Path, m.File.IsDir())
}

func (m Chmoded) String() string {
	return fmt.Sprintf("CHMODED (%s) (dir=%t)", m.Path, m.File.IsDir())
}
