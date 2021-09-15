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

var vidFIlesToForwardChan chan string = make(chan string, 10)

var streams []chan []byte
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
		case packet, ok := <-streamChan: //packet is a byte array received from the camtron ui by the websocket

			if !ok {
				log.Print("WARNING: Failed to get packet")
			}

			data = append(data, packet...)

			if len(data) > 1000000 {
				var index string = strconv.Itoa(vidCount)
				tempFileName := "tempvid-" + index + ".webm"
				camtron.StopRecording()
				//workaround until I figure out if this can be done using stdin and stdout.
				if !saveTempVid(tempFileName, data) {
					return
				}

				vidFIlesToForwardChan <- tempFileName

				data = nil
				vidCount = vidCount + 1

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

func BroadcastStream() {
	vidCount := 1
	for {
		select {
		case vidFIleToForward, ok := <-vidFIlesToForwardChan:
			if !ok {
				log.Print("WARNING: Something went wrong getting file to forward")
			}
			sourveVidFile := tempVidDir + "/" + vidFIleToForward
			for _, targetUrl := range targetUrls {
				runFFMPEGToTCP(sourveVidFile, targetUrl)
			}
			err := os.Remove(sourveVidFile)
			if err != nil {
				fmt.Println(err)
			}
			vidCount++
		}
	}

}

func runFFMPEGToTCP(sourceFile string, targetURL string) {
	cmd := exec.Command("ffmpeg", "-i", sourceFile, "-f", "mjpeg", targetURL)
	fmt.Println(cmd)
	err := cmd.Start()
	if err != nil {
		log.Panic(err)
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Println(err)
	}
}

func StartBroadcastStreamConsumer() {
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
	go BroadcastStream()
}

type Configuration struct {
	TargetUrls []string
}
