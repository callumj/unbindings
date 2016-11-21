package core

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"

	shellwords "github.com/mattn/go-shellwords"
)

var (
	confReg = regexp.MustCompile("^cnf\\|(?P<key>[^|]+)\\|(?P<value>.+)\r?\n?$")
)

type Invocation struct {
	Resident bool
	StdOut   io.ReadCloser

	cmd           *exec.Cmd
	stdIn         io.WriteCloser
	parentOpts    map[string]string
	childOpts     map[string]string
	quitReceiving chan bool
}

func NewInvocation(path string, resident bool) (*Invocation, error) {
	p := shellwords.NewParser()
	args, err := p.Parse(path)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(args[0], args[1:]...)
	stdOut, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stdIn, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	return &Invocation{
		Resident:   resident,
		StdOut:     stdOut,
		cmd:        cmd,
		stdIn:      stdIn,
		parentOpts: make(map[string]string),
		childOpts:  make(map[string]string),
	}, nil
}

func (i *Invocation) SetOption(key, value string) error {
	i.parentOpts[key] = value

	// if running resident then we need to notify of a change
	if i.Resident && i.cmd.Process != nil {
		return i.writeConfig()
	}

	return nil
}

func (i *Invocation) Wait() error {
	if i.cmd.Process == nil {
		return errors.New("Please start me")
	}

	err := i.cmd.Wait()
	i.stdIn.Close()
	return err
}

func (i *Invocation) Start() error {
	i.cmd.Env = i.envOpts()
	if err := i.cmd.Start(); err != nil {
		return err
	}
	if !i.Resident {
		return nil
	}

	if err := i.writeConfig(); err != nil {
		return err
	}
	i.readIncomingConfig()
	return nil
}

func (i *Invocation) readIncomingConfig() {
	i.quitReceiving = make(chan bool)
	scanner := bufio.NewScanner(i.StdOut)
	go func() {
		for {
			select {
			case <-i.quitReceiving:
				return
			default:
				scanned := scanner.Scan()
				if scanned {
					parts := confReg.FindStringSubmatch(scanner.Text())
					if len(parts) != 0 {
						i.childOpts[parts[1]] = parts[2]
						fmt.Printf("%+v\n", i.childOpts)
					} else {

					}
				}
			}
		}
	}()
}

func (i *Invocation) writeConfig() error {
	if i.parentOpts == nil {
		return nil
	}

	for k, v := range i.parentOpts {
		_, err := i.stdIn.Write([]byte("cnf|" + k + "|" + v + "\n"))
		if err != nil {
			return err
		}
	}

	_, err := i.stdIn.Write([]byte("compl|\n"))
	if err != nil {
		return err
	}

	return nil
}

func (i *Invocation) envOpts() []string {
	opts := []string{}
	for k, v := range i.parentOpts {
		opts = append(opts, "UB_CONFIG_"+strings.ToUpper(k)+"="+v)
	}
	return opts
}
