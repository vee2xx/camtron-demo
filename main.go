package main

/*
#include<stdio.h>
#include <stdlib.h>
extern char * inCFile(char *str);
*/
import "C"
import (
	"camtron-demo/consumers"

	"github.com/vee2xx/camtron"
)

func main() {

	consumers.StartBroadcastStreamConsumer()
	camtron.StartCam()

}
