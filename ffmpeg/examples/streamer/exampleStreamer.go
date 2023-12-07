package main

import (
	"log"
	"main/ffmpeg"
	"os"
	"bufio"
)

func main() {
	st := ffmpeg.NewFFStreamer(5000, log.New(os.Stdout, "", log.LstdFlags), "streamer1", "10.0.3.20", true)
	st.AddStream("./ffmpeg/examples/test.mp4", "test", 0)
	st.AddStream("./ffmpeg/examples/test1.mp4", "test1", 0)
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	st.Stop()
}