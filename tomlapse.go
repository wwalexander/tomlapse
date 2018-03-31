package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const format = "20060102T150405Z0700"
const ext = ".jpg"

func FrameList() ([]string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var frames []string
	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, ext) && name != "tomlapse.jpg" {
			frames = append(frames, name)
		}
	}
	return frames, nil
}

func LastFrameTime() (*time.Time, error) {
	frames, err := FrameList()
	if err != nil {
		return nil, err
	}
	if len(frames) == 0 {
		return &time.Time{}, nil
	}
	lastName := strings.TrimSuffix(frames[len(frames)-1], ext)
	lastTime, err := time.Parse(format, lastName)
	if err != nil {
		return nil, err
	}
	return &lastTime, nil
}

func GenerateList() error {
	frames, err := FrameList()
	if err != nil {
		return err
	}
	list, err := os.Create("mylist.txt")
	if err != nil {
		return err
	}
	defer list.Close()
	for _, frame := range frames {
		fmt.Fprintf(list, "file '%s'\n", frame)
	}
	return nil
}

func SyncFrame() error {
	resp, err := http.Get("https://www.usc.edu/cameras/tommycam.jpg")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	lastModified := resp.Header.Get("Last-Modified")
	frameTime, err := time.Parse(time.RFC1123, lastModified)
	if err != nil {
		return err
	}
	lastFrameTime, err := LastFrameTime()
	if err != nil {
		return err
	}
	fmt.Println(frameTime, lastFrameTime)
	if !frameTime.After(*lastFrameTime) {
		return nil
	}
	current, err := os.Create("tomlapse.jpg")
	if err != nil {
		return err
	}
	defer current.Close()
	name := frameTime.Format(format) + ".jpg"
	frame, err := os.Create(name)
	if err != nil {
		return err
	}
	defer frame.Close()
	w := io.MultiWriter(current, frame)
	if _, err := io.Copy(w, resp.Body); err != nil {
		return err
	}
	fmt.Println(name)
	if err := GenerateList(); err != nil {
		return err
	}
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-safe", "0",
		"-r", "30",
		"-f", "concat",
		"-i", "mylist.txt",
		"-c:v", "libx264",
		"-crf", "18",
		"-f", "mp4",
		"tomlapse.mp4.tmp",
	)
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := os.Rename("tomlapse.mp4.tmp", "tomlapse.mp4"); err != nil {
		return err
	}
	return nil
}

func Update() {
	if err := SyncFrame(); err != nil {
		log.Println(err)
		return
	}
}

func main() {
	ticker := time.NewTicker(55 * time.Second)
	go Update()
	for range ticker.C {
		go Update()
	}
}
