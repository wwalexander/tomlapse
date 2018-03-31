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
		if !strings.HasSuffix(name, ".jpg") {
			continue
		}
		frames = append(frames, name)
	}
	return frames, nil
}

func nameToTime(name string) (time.Time, error) {
	name = strings.TrimSuffix(name, ".jpg")
	t, err := time.Parse(format, name)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func LastFrameTime() (time.Time, error) {
	frames, err := FrameList()
	if err != nil {
		return time.Time{}, err
	}
	if len(frames) == 0 {
		return time.Time{}, nil
	}
	lastTime, err := nameToTime(frames[len(frames)-1])
	if err != nil {
		return time.Time{}, err
	}
	return lastTime, nil
}

func GetFrame() error {
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
	if !frameTime.After(lastFrameTime) {
		return nil
	}
	name := frameTime.Format(format) + ".jpg"
	log.Println(name)
	frame, err := os.Create(name)
	if err != nil {
		return err
	}
	defer frame.Close()
	if _, err := io.Copy(frame, resp.Body); err != nil {
		return err
	}
	return nil
}

func MakeListFile() error {
	frames, err := FrameList()
	if err != nil {
		return err
	}
	var start int
	for i := len(frames) - 1; i >= 0; i-- {
		t, err := nameToTime(frames[i])
		if err != nil {
			return err
		}
		if time.Since(t) >= 24*time.Hour {
			start = i
			break
		}
	}
	list, err := os.Create("mylist.txt")
	if err != nil {
		return err
	}
	defer list.Close()
	for _, frame := range frames[start:] {
		fmt.Fprintf(list, "file '%s'\n", frame)
	}
	return nil
}

func Update() error {
	if err := GetFrame(); err != nil {
		return err
	}
	if err := MakeListFile(); err != nil {
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
		"-movflags", "+faststart",
		"tomlapse.mp4.tmp",
	)
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := os.Rename("tomlapse.mp4.tmp", "tomlapse.mp4"); err != nil {
		return err
	}
	if err := os.Remove("mylist.txt"); err != nil {
		return err
	}
	return nil
}

func main() {
	for {
		go func() {
			if err := Update(); err != nil {
				log.Println(err)
			}
		}()
		time.Sleep(30 * time.Second)
	}
}
