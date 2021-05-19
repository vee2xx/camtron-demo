package main

import (
	"github.com/vee2xx/camtron"
)

func main() {

	camtron.StartStreamToFileConsumer()

	camtron.StartCam()

}
