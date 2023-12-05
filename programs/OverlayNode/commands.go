package main

import (
	"main/shell"
)
func (rp *OverlayNode) registerCommands(s *shell.Shell) {
	s.RegisterCommand("leave", func(args []string) {
		rp.LeaveOverlay()
	})
}
