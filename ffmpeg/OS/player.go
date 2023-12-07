package OS

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
)

type Player struct {
	sdp_file_path string
	process *os.Process
	logger *log.Logger
}


func NewPlayer(sdp_file_path string, logger *log.Logger, debug bool) *Player {
	p := new(Player)
	p.sdp_file_path = sdp_file_path
	p.logger = logger
	loglevel := ""
	if debug {
		loglevel = "level+32"
	} else {
		loglevel = "level+24"
	}
	cmd := exec.Command("ffplay", sdp_file_path, "-hide_banner", "-loglevel", loglevel, "-protocol_whitelist", "file,rtp,udp", "-fflags", "nobuffer", "-flags", "low_delay", "-framedrop", "-strict" ,"experimental", "-probesize", "32", "-analyzeduration", "0")
	logger.Printf("Starting ffplay with command: %s\n", cmd.String())
	//handle error
	errorPipe, _ := cmd.StderrPipe()
	go p.start_logger("ffplay", errorPipe, '\r')
	cmd.Start()
	p.process = cmd.Process
	return p
}

func (p Player) start_logger(descriptor string, pipe io.ReadCloser, delimiter byte) {
	reader := bufio.NewReader(pipe)
	for {
		line, err := reader.ReadString(delimiter)
		if err != nil {
			p.logger.Printf("[%s]: Player Stopped\n", descriptor)
			break
		}
		p.logger.Printf("[%s]: %s\n", descriptor, line)
	}
}

func  (p Player) Stop() {
	p.process.Kill()
	p.process.Wait()
}
