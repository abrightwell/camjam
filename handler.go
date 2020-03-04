package main

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strconv"
	"syscall"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	boundary string = "frame"
)

func stream(w http.ResponseWriter, r *http.Request) {
	log.Infof("received stream request from %s", r.RemoteAddr)

	w.Header().Add(
		"Content-Type",
		fmt.Sprintf("multipart/x-mixed-replace;boundary=%s", boundary),
	)

	multipartWriter := multipart.NewWriter(w)
	multipartWriter.SetBoundary(boundary)

	for {
		// wait for a frame to be available
		frame := <-frames

		partWriter, err := multipartWriter.CreatePart(textproto.MIMEHeader{
			"Content-Type":   []string{"image/jpeg"},
			"Content-Length": []string{strconv.Itoa(len(frame))},
		})

		// Handle any errors that might have occur creating the multipart
		// writer. If an error occurs that is not a 'broken pipe' error, then
		// log it, otherwise just break out of the loop.
		//
		// TODO: handle errors with appropriate status code?
		if err != nil {
			if !errors.Is(err, syscall.EPIPE) {
				log.Errorf("error creating multipart: %s", err.Error())
			}
			break
		}

		// Write the frame to the stream. If an error occurs that is not a
		// 'broken pipe' error then log, otherwise just break out of the loop.
		//
		// TODO: handle errors with appropriate status code?
		if _, err := partWriter.Write(frame); err != nil {
			if !errors.Is(err, syscall.EPIPE) {
				log.Errorf("error writing to stream: %s", err.Error())
			}
			break
		}
	}

	log.Infof("closing stream to %s", r.RemoteAddr)
}

func listCameras(w http.ResponseWriter, r *http.Request) {
}

func getCamera(w http.ResponseWriter, r *http.Request) {
}

func switchCamera(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	lock.Lock()
	defer lock.Unlock()

	if id > len(cameras)-1 {
		return
	}

	selected = cameras[id]
}
