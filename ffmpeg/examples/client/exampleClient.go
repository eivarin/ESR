package main

import (
	"bufio"
	"log"
	"main/ffmpeg"
	"main/ffmpeg/OS"
	"os"
)

func main() {
	cl := ffmpeg.NewFFClient(5000, log.New(os.Stdout, "", log.LstdFlags), false)
	SDP_string := OS.Generate_SDP_From_FFMPEG("./ffmpeg/examples/test.mp4")
	cl.AddInstance("test", SDP_string, false)
	SDP_string1 := OS.Generate_SDP_From_FFMPEG("./ffmpeg/examples/test1.mp4")
	cl.AddInstance("test1", SDP_string1, true)
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cl.Stop()
}