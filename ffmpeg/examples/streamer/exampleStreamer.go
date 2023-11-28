package main

import (
	"log"
	"main/ffmpeg"
	"os"
	"bufio"
)

func main() {
	st := ffmpeg.NewFFStreamer(5000, log.New(os.Stdout, "", log.LstdFlags), "streamer1", "127.0.0.1", true)
	st.AddStream("./ffmpeg/examples/test.mp4")
	st.AddStream("./ffmpeg/examples/test1.mp4")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	st.Stop()
}