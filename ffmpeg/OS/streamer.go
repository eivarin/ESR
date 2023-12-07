package OS

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
)

type Streamer struct {
	media_path string
	process *os.Process
	logger *log.Logger
	debug bool
}

func get_generated_SDP(input io.ReadCloser) string {
	SDP_bytes := new(bytes.Buffer)
	reader := bufio.NewReader(input)
	line, _ := reader.ReadString('\n')
	if line != "SDP:\n" {
		SDP_bytes.WriteString(line)
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		SDP_bytes.WriteString(line)
	}
	return SDP_bytes.String()
}

func Generate_SDP_From_FFMPEG(input string) string {
	cmnd := get_SDP_command(input)
	output_pipe,_ := cmnd.StdoutPipe()
	cmnd.Start()
	SDP_string := get_generated_SDP(output_pipe)
	cmnd.Wait()
	return SDP_string
}

func NewStreamer(media_path string, logger *log.Logger, debug bool, ports []int, startAtSeconds int64) *Streamer {
	s := new(Streamer)
	s.media_path = media_path
	s.logger = logger
	s.debug = debug
	cmd := get_streaming_command(media_path, ports, startAtSeconds)
	s.logger.Printf("Starting STREAMER with command: %s\n", cmd.String())
	//handle error
	errorPipe, _ := cmd.StderrPipe()
	go s.start_logger("STREAMER", errorPipe, '\r')
	cmd.Start()
	s.process = cmd.Process
	return s
}


func (s Streamer) start_logger(descriptor string, pipe io.ReadCloser, delimiter byte) {
	reader := bufio.NewReader(pipe)
	for {
		line, err := reader.ReadString(delimiter)
		if err != nil {
			s.logger.Printf("[%s]: Streamer Stopped\n", descriptor)
			break
		}
		if s.debug {
			s.logger.Printf("[%s]: %s\n", descriptor, line)
		}
	}
}

func (s Streamer) Stop() {
	s.process.Kill()
	s.process.Wait()
}
