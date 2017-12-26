package util

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/abrightwell/webcam"
)

type Camera struct {
	Name   string
	Device string
}

func isV4LDevice(name string) bool {
	return strings.HasPrefix(name, "video") ||
		strings.HasPrefix(name, "radio") ||
		strings.HasPrefix(name, "vbi") ||
		strings.HasPrefix(name, "v4l-subdev")
}

func convertCardName(s []uint8) string {
	b := make([]byte, 0)
	for _, v := range s {
		if v != 0 {
			b = append(b, byte(v))
		}
	}
	return string(b)
}

func ListCameras() []Camera {
	devDir, _ := os.Open("/dev")

	files, _ := devDir.Readdir(0)

	cameras := make([]Camera, 0)

	for _, fileInfo := range files {
		if isV4LDevice(fileInfo.Name()) {
			path := fmt.Sprintf("/dev/%s", fileInfo.Name())
			file, err := os.OpenFile(path, unix.O_RDWR|unix.O_NONBLOCK, 0666)

			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			caps, _ := webcam.GetCapabilities(file.Fd())

			cam := Camera{
				Name:   convertCardName(caps.Card[:]),
				Device: path,
			}

			cameras = append(cameras, cam)
		}
	}

	return cameras
}
