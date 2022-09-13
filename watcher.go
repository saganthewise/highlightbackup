package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
	"github.com/harry1453/go-common-file-dialog/cfd"
)

type Config struct {
	Clippath string
	Savepath string
}

func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func main() {
	var (
		config Config
		buf    = new(bytes.Buffer)
	)
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		log.Fatal(err)
	}
	if config.Clippath == "" {
		pickFolderDialog, err := cfd.NewSelectFolderDialog(cfd.DialogConfig{
			Title: "Pick Clip Folder",
			Role:  "Pick Clip Folder",
		})
		if err != nil {
			log.Fatal(err)
		}
		if err := pickFolderDialog.Show(); err != nil {
			log.Fatal(err)
		}
		result, err := pickFolderDialog.GetResult()
		config.Clippath = result
	}
	if config.Savepath == "" {
		pickFolderDialog, err := cfd.NewSelectFolderDialog(cfd.DialogConfig{
			Title: "Pick Save Folder",
			Role:  "Pick Save Folder",
		})
		if err != nil {
			log.Fatal(err)
		}
		if err := pickFolderDialog.Show(); err != nil {
			log.Fatal(err)
		}
		result, err := pickFolderDialog.GetResult()
		config.Savepath = result
	}
	err = toml.NewEncoder(buf).Encode(config)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("config.toml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString(buf.String())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(buf.String())
	if err == cfd.ErrorCancelled {
		log.Fatal("Dialog was cancelled by the user.")
	} else if err != nil {
		log.Fatal(err)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					filename := strings.Split(event.Name, "\\")
					Copy(event.Name, config.Savepath+"//"+filename[len(filename)-1])

				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(config.Clippath)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
