package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mazzegi/rwatch"
)

func main() {
	dir := flag.String("d", ".", "directory to watch recursively")
	flag.Parse()
	rw, err := rwatch.NewRecursiveWatcher(*dir)
	if err != nil {
		panic(err)
	}
	fmt.Println("watching", *dir)
	go func() {
		for m := range rw.Messages {
			fmt.Println(m)
		}
	}()

	sigC := make(chan os.Signal)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	<-sigC
	rw.Close()
}
