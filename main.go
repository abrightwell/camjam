package main

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/jpeg"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/blackjack/webcam"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"
)

// command options
var (
	logLevel   string
	configFile string
)

//
var (
	cameras  []*Camera
	selected *Camera
	frames   chan []byte
	lock     *sync.Mutex
	quit     chan struct{}
)

const (
	V4L2_PIX_FMT_PJPG = 0x47504A50
	V4L2_PIX_FMT_YUYV = 0x56595559
)

type Camera struct {
	device string
	wc     *webcam.Webcam
	format webcam.PixelFormat
	width  uint32
	height uint32
}

func (c *Camera) ReadFrame() ([]byte, error) {
	return c.wc.ReadFrame()
}

func (c *Camera) WaitForFrame(timeout uint32) error {
	return c.wc.WaitForFrame(timeout)
}

func (c *Camera) StopStreaming() error {
	return c.wc.StopStreaming()
}

func (c *Camera) StartStreaming() error {
	return c.wc.StartStreaming()
}

func init() {
	flag.StringVarP(&configFile, "config", "", "", "path to config file")
	flag.StringVarP(&logLevel, "log-level", "", "info", "logging level")

	config.SetConfigType("yaml")
	config.SetConfigName("config")
	config.AddConfigPath("/etc/camjam")
	config.AddConfigPath("$HOME/.camjam")
	config.AddConfigPath(".")

	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	log.SetOutput(os.Stderr)
}

func initCameras(c Config) {
	cameras = make([]*Camera, 0)

	for _, cam := range c.Cameras {
		var format webcam.PixelFormat

		wc, err := webcam.Open(cam.Device)

		log.Infof("opening camera device: %s", cam.Device)
		if err != nil {
			log.Errorf("could not open camera device: %s", err.Error())
			continue
		}

		// Determine the format code for the configured pixel format. If the
		// configured format isn't supported then log an error and move on to
		// the next camera.
		//
		// TODO: Check that the camera actually supports the configured pixel
		// format.
		switch cam.Format {
		case "MJPG":
			format = V4L2_PIX_FMT_PJPG
		case "YUYV":
			format = V4L2_PIX_FMT_YUYV
		default:
			log.Errorf("invalid pixel format: %s", cam.Format)
			continue
		}

		// Set the camera's image format. The device driver might change these
		// values, so we'll reassign them so that they are ensured to be
		// correct for later usage.
		format, width, height, err := wc.SetImageFormat(format, cam.Width, cam.Height)

		// If the image format cannot be set, then log an error and move on to
		// the next camera.
		if err != nil {
			log.Error(err)
			continue
		}

		log.Infof("Camera(%s, %d, %d, %d", cam.Device, format, width, height)

		// Set the buffer count for the webcam to 1. Otherwise, there might be
		// frames sent that are no longer relevant. This helps keep the image
		// stream closer to real time than it would be otherwise.
		if err := wc.SetBufferCount(1); err != nil {
			log.Error("could not set buffer count: %s", err.Error())
			continue
		}

		// Enable the cameras capture process.
		if err := wc.StartStreaming(); err != nil {
			log.Error("could not enable camera: %s", err.Error())
			continue
		}
		log.Infof("enabled camera: %s", cam.Device)

		camera := &Camera{
			device: cam.Device,
			wc:     wc,
			format: format,
			width:  width,
			height: height,
		}

		cameras = append(cameras, camera)
	}
}

// convertYUYV encodes an image frame that was captured from a camera using the
// YUYV format to JPEG format.
func convertYUYV(width int, height int, frame []byte) []byte {
	rect := image.Rect(0, 0, width, height)
	yuyv := image.NewYCbCr(rect, image.YCbCrSubsampleRatio422)

	for i := range yuyv.Cb {
		ii := i * 4
		yuyv.Y[i*2] = frame[ii]
		yuyv.Y[i*2+1] = frame[ii+2]
		yuyv.Cb[i] = frame[ii+1]
		yuyv.Cr[i] = frame[ii+3]
	}

	buffer := &bytes.Buffer{}

	if err := jpeg.Encode(buffer, yuyv, nil); err != nil {
		log.Errorf("error converting YUYV frame: %s", err.Error())
		return []byte{}
	}

	return buffer.Bytes()
}

func main() {
	var c Config

	flag.Parse()

	// Set the logging level.
	level, _ := log.ParseLevel(logLevel)
	log.SetLevel(level)

	// Read in the configuration file.
	if configFile != "" {
		config.SetConfigFile(configFile)
	}

	if err := config.ReadInConfig(); err != nil {
		log.Fatalf("could not read in config: %s", err.Error())
	}

	if err := config.Unmarshal(&c); err != nil {
		log.Fatalf("could not unmarshal config: %s", err.Error())
	}

	// Initialize all of the configured cameras.
	initCameras(c)

	// If no cameras were successfully configured, then fail startup completely.
	if len(cameras) == 0 {
		log.Fatal("no cameras were configured")
	}

	// Setup stop signal channel. This channel will handle listening for a
	// SIGTERM signal in order to gracefully shutdown.
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// set the selected camera
	// TODO: make this configurable.
	selected = cameras[0]

	// initialize the frames channel
	frames = make(chan []byte)
	lock = &sync.Mutex{}

	// start capturing camera frames
	//
	// TODO: refactor this to a named function.
	go func() {
		ticker := time.NewTicker(c.Server.Interval)
		for {
			select {
			case <-ticker.C:
				// Acquire a lock on the selected camera.
				lock.Lock()

				// Wait for the frame to be readable.
				if err := selected.WaitForFrame(5); err != nil {
					switch {
					case errors.Is(err, &webcam.Timeout{}):
						log.Warnf("camera timed out: %s", err.Error())
						continue
					default:
						log.Fatal(err.Error())
					}
				}

				// Read the current frame from the camera.
				frame, err := selected.ReadFrame()

				// Release the lock on the camera before handling any errors
				// that might have occurred while reading the frame.
				lock.Unlock()

				// If an error occured reading the from the camera, then log it
				// and move on.
				if err != nil {
					log.Errorf("an error occured reading camera frame: %s", err.Error())
					continue
				}

				// If the camera pixel format is configured as YUYV, then we
				// need to convert the image frame to JPEG.
				if selected.format == V4L2_PIX_FMT_YUYV {
					frame = convertYUYV(int(selected.width), int(selected.height), frame)
				}

				// Send the frame to the frames channel so that those things
				// which care about it can use it.
				frames <- frame

			case <-quit:
				break
			}
		}
	}()

	// Setup all of the HTTP service routes.
	router := mux.NewRouter()
	router.HandleFunc("/stream", stream).Methods("GET")
	router.HandleFunc("/cameras", listCameras).Methods("GET")
	router.HandleFunc("/cameras/{id:[0-9]+}", getCamera).Methods("GET")
	router.HandleFunc("/switch/{id:[0-9]+}", switchCamera).Methods("PUT")

	// Create the HTTP server.
	server := &http.Server{Addr: c.Server.Address, Handler: router}

	// Start the HTTP Server, listening on the configured port.
	go func() {
		log.Infof("listening on %s", c.Server.Address)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err.Error())
		}
	}()

	// Wait for the stop signal.
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("error shutting down server: %s", err.Error())
	} else {
		log.Info("http server shutdown.")
	}

	// Disable all of the cameras.
	for _, camera := range cameras {
		if err := camera.StopStreaming(); err != nil {
			log.Errorf("could not stop streaming camera: %s: %s", camera.device, err.Error())
		}
		log.Infof("shutdown camera: %s", camera.device)
	}
}
