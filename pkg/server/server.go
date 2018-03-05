package server

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strconv"
	"sync"

	"github.com/blackjack/webcam"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/abrightwell/camjam/pkg/config"
)

type Server struct {
	Router  *mux.Router
	Cameras []*webcam.Webcam
	config  config.Config
	current *webcam.Webcam
	frames  map[*webcam.Webcam]chan []byte
	lock    *sync.Mutex
}

var quit chan (bool) = make(chan (bool), 1)

func (s *Server) Initialize(c config.Config) {
	s.Router = mux.NewRouter()
	s.config = c
	s.lock = &sync.Mutex{}
	s.initializeRoutes()
	s.initializeCameras()
}

func (s *Server) initializeRoutes() {
	s.Router.HandleFunc("/stream", s.streamHandler).Methods("GET")
	s.Router.HandleFunc("/switch/{id:[0-9]+}", s.switchCamHandler).Methods("PUT")
}

func formatToCode(f string) uint32 {
	return uint32(f[0]) | uint32(f[1])<<8 | uint32(f[2])<<16 | uint32(f[3])<<24
}

func (s *Server) initializeCameras() {
	numCameras := len(s.config.Cameras)
	s.frames = make(map[*webcam.Webcam]chan []byte, numCameras)
	s.Cameras = make([]*webcam.Webcam, numCameras)

	for index, cam := range s.config.Cameras {
		wc, err := webcam.Open(cam.Device)
		s.frames[wc] = make(chan []byte)

		if err != nil {
			log.Fatalf("Error opening camera: %s - %s", cam.Device, err.Error())
		}

		code := formatToCode(cam.Format)

		wc.SetImageFormat(webcam.PixelFormat(code), uint32(cam.Width), uint32(cam.Height))
		wc.StartStreaming()
		go s.startCamera(wc)

		s.Cameras[index] = wc
	}

	s.current = s.Cameras[0]
}

func (s *Server) Start() {
	log.Infof("Starting server on %s", s.config.Server.Address)
	if err := http.ListenAndServe(s.config.Server.Address, s.Router); err != nil {
		log.Fatal(err.Error())
	}
}

func (s *Server) startCamera(cam *webcam.Webcam) {
	for {
		cam.WaitForFrame(5)
		frame, err := cam.ReadFrame()

		if err != nil {
			log.Error(err.Error())
		}

		s.frames[cam] <- frame
	}
}

func (s *Server) streamHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("Received stream request from %s", r.RemoteAddr)
	mimeWriter := multipart.NewWriter(w)

	contentType := fmt.Sprintf("multipart/x-mixed-replace;boundary=%s",
		mimeWriter.Boundary())

	w.Header().Add("Content-Type", contentType)

	partHeader := make(textproto.MIMEHeader)
	partHeader.Add("Content-Type", "image/jpeg")

	for {
		partWriter, err := mimeWriter.CreatePart(partHeader)

		if err != nil {
			break
		}

		frame := <-s.frames[s.current]

		partWriter.Write(frame)
	}
}

func (s *Server) switchCamHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	s.lock.Lock()
	defer s.lock.Unlock()

	s.current = s.Cameras[id]
}
