package OS

import (
	"fmt"
	"os/exec"
)


func get_SDP_command(input string) *exec.Cmd {
	command := []string{"-an", "-vcodec", "copy", "-f", "rtp", "rtp://127.0.0.1:11111", "-vn", "-acodec", "copy", "-f", "rtp", "rtp://127.0.0.1:11111", "-hide_banner"}
	return exec.Command("ffmpeg", append([]string{"-fflags", "+genpts", "-i", input}, command...)...)
}

func get_streaming_command(input string, ports []int, startAt int64) *exec.Cmd {
	command := []string{"-an", "-vcodec", "copy", "-f", "rtp", "rtp://127.0.0.1:" + fmt.Sprint(ports[0]), "-vn", "-acodec", "copy", "-f", "rtp", "rtp://127.0.0.1:" + fmt.Sprint(ports[2]), "-hide_banner"}
	return exec.Command("ffmpeg", append([]string{"-re", "-stream_loop", "-1", "-fflags", "nobuffer", "-fflags", "+genpts", "-ss", fmt.Sprint(startAt), "-i", input}, command...)...)
}