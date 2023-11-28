package main

import (
	"bufio"
	"log"
	"main/ffmpeg"
	"main/ffmpeg/OS"
	"os"
)

func main() {
	cl := ffmpeg.NewFFClient(5000, log.New(os.Stdout, "", log.LstdFlags))
	SDP_string := OS.Generate_SDP_From_FFMPEG("./ffmpeg/examples/test.mp4")
	cl.AddInstance("./ffmpeg/examples/test.mp4", SDP_string, false)
	SDP_string1 := OS.Generate_SDP_From_FFMPEG("./ffmpeg/examples/test1.mp4")
	cl.AddInstance("./ffmpeg/examples/test1.mp4", SDP_string1, true)
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cl.Stop()
}