package shell

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Shell struct {
	name string
	commands map[string]func(args []string)
	logger  *log.Logger
	Lock *sync.RWMutex
}

func NewShell(logger *log.Logger, name string) *Shell {
	shell := new(Shell)
	shell.name = name
	shell.commands = make(map[string]func(args []string))
	shell.logger = logger
	shell.Lock = &sync.RWMutex{}
	shell.RegisterCommand("help", shell.shellHelp)
	return shell
}

func (s *Shell) RegisterCommand(command string, handler func(args []string)) {
	s.commands[command] = handler
}

func (s *Shell) shellHelp(args []string) {
	fmt.Println("Possible commands:")
	for command := range s.commands {
		fmt.Println(command)
	}
}

func (s *Shell) Run() {
	fmt.Printf("%s - Type 'exit' to quit\n", s.name)

	for {
		s.Lock.RLock()
		fmt.Printf("%s> ", s.name)
		s.Lock.RUnlock()
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()

		if input == "exit" {
			break
		}

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		if handler, ok := s.commands[command]; ok {
			handler(args)
		} else {
			fmt.Println("Unknown command:", command)
			s.commands["help"](nil)
		}
	}
}

func CheckIfIsIp(host string) bool {
	parts := strings.Split(host, ".")

	if len(parts) < 4 {
		return false
	}
	
	for _,x := range parts {
		if i, err := strconv.Atoi(x); err == nil {
			if i < 0 || i > 255 {
			return false
		}
		} else {
			return false
		}

	}
	return true
}