package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/vee2xx/camtron"
)

var config Config

type Config struct {
	Video struct {
		MaxSize  string `json:"maxSize"`
		Encoding string `json:"encoding"`
		Format   string `json:"format"`
	} `json:"video"`
}

func Configure() {

	file, fileErr := os.Open("config.json")

	if fileErr != nil {
		fmt.Println("Unable to open configuration file", fileErr.Error)
	}

	defer file.Close()
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&config)

	if err != nil {
		fmt.Println("Error parsing configuration.", err.Error)
	}
}

func main() {

	// fmt.Println(os.Environ())
	// goPath, _ := os.UserHomeDir()
	// log.Println(goPath + "/go/src")

	Configure()
	if _, err := os.Stat("vids"); os.IsNotExist(err) {
		os.Mkdir("vids", os.ModePerm)
	}

	var consumers map[string]camtron.StreamConsumer = make(map[string]camtron.StreamConsumer)
	var vidToFileStream = make(chan []byte, 10)
	options := make(map[string]string)
	options["filePath"] = "./vids/vid-" + time.Now().Format("2006_01_02_15_04_05") + "." + config.Video.Format
	options["maxSize"] = config.Video.MaxSize
	streamConsumer := camtron.StreamConsumer{Stream: vidToFileStream, Context: make(chan string), Handler: camtron.StreamToFile, Options: options}
	consumers["file"] = streamConsumer

	camtron.StartCam(consumers)
}
