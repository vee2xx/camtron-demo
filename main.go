package main

import (
	"camtron-demo/consumers"

	"github.com/vee2xx/camtron"
)

func main() {
	consumers.StartForwardStreamConsumer()
	camtron.StartCam()
}
