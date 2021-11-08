package consumers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/vee2xx/camtron"
)

var vidFilesToForwardChan chan string = make(chan string, 10)
var tempVidDir string = "videos"
var targetUrls []string

func SaveVideo(streamChan chan []byte) {
	if _, err := os.Stat(tempVidDir); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(tempVidDir, os.ModePerm)
		}
	}
	var vidCount int = 1
	var data []byte
	for {
		select {
		//receive bytes pushed onto channel by camtron library
		case packet, ok := <-streamChan:
			if !ok {
				log.Print("WARNING: Failed to get packet")
			}

			data = append(data, packet...)
			// save the video to a file every once in a while
			if len(data) > 1000000 {
				var index string = strconv.Itoa(vidCount)
				tempFileName := "tempvid-" + index + ".webm"

				// stop recording so that we get a new video
				// after saving this one
				camtron.StopRecording()
				if !saveTempVid(tempFileName, data) {
					return
				}
				// push file name onto a channel to
				// notify the forwarding process that
				// it is available
				vidFilesToForwardChan <- tempFileName
				// clear everything
				data = nil
				vidCount = vidCount + 1
				// start recording to get a new video
				camtron.StartRecording()
			}
		case val, _ := <-camtron.Context:
			if val == "stop" {
				close(streamChan)
				log.Println("INFO: Shutting streaming to clients")
				return
			}
		}
	}
}

//Standard function to save a file
func saveTempVid(fname string, video []byte) bool {
	tempVidFile := tempVidDir + "/" + fname
	vidFile, fileOpenErr := os.OpenFile(tempVidFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if fileOpenErr != nil {
		log.Println(fileOpenErr)
	}
	defer vidFile.Close()
	_, statErr := vidFile.Stat()
	if statErr != nil {
		log.Println(statErr)
	}
	_, writeErr := vidFile.Write(video)
	if writeErr != nil {
		log.Println(writeErr)
		return false
	}
	return true
}

func ForwardStream() {
	for {
		select {
		case vidFIleToForward, ok := <-vidFilesToForwardChan:
			if !ok {
				log.Print("WARNING: Something went wrong getting file to forward")
			}
			sourveVidFile := tempVidDir + "/" + vidFIleToForward
			for _, targetUrl := range targetUrls {
				runFFMPEGToTCP(sourveVidFile, targetUrl)
			}
			// clean up the temp file
			err := os.Remove(sourveVidFile)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
func runFFMPEGToTCP(sourceFile string, targetURL string) {
	cmd := exec.Command("ffmpeg", "-i", sourceFile, "-f", "mjpeg", targetURL)
	err := cmd.Start()
	if err != nil {
		log.Panic(err)
	}
	err = cmd.Wait()
	if err != nil {
		log.Println(err)
	}
}

func StartForwardStreamConsumer() {
	file, _ := os.Open("consumers/conf.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	targetUrls = configuration.TargetUrls
	vidStream := make(chan []byte, 10)
	camtron.RegisterStream(vidStream)
	go SaveVideo(vidStream)
	go ForwardStream()
}

type Configuration struct {
	TargetUrls []string
}
