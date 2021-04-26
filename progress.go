package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type progress struct {
	started   bool
	name      string
	end       chan bool
	written   int64
	total     *int64
	previous  time.Time
	maxStrLen int
}

func (prog *progress) Write(p []byte) (n int, err error) {
	n = len(p)
	prog.started = true
	prog.written += int64(n)
	return
}

func (prog *progress) Close() error {
	prog.end <- true
	return nil
}

func (prog *progress) Run() {
	begin := time.Now()
	prog.written = 0
	prog.previous = begin
	prog.end = make(chan bool)
	prog.maxStrLen = 0
	c := time.Tick(100 * time.Millisecond)
	var written int64
	var speed string
	for {
		diff := time.Since(prog.previous).Seconds()
		if diff > 0.5 {
			speed = "at " + humanize(int64(float64(prog.written-written)/diff)) + "/s"
			written = prog.written
			prog.previous = time.Now()
		}
		prog.print(fmt.Sprint("Processing `", prog.name, "` ", humanize(prog.written), prog.getTotal(), " ", speed))
		select {
		case <-prog.end:
			duration := time.Since(begin)
			speed = "at " + humanize(int64(float64(prog.written)/duration.Seconds())) + "/s"
			prog.print(fmt.Sprint("Done in ", duration.Truncate(time.Millisecond), " ", speed, ", waiting for response"))
			return
		case <-c:
			continue
		}
	}
}

func (prog *progress) getTotal() (total string) {
	if prog.total != nil {
		total = " of " + humanize(*prog.total)
	}
	return
}

func (prog *progress) print(line string) {
	if prog.started == false {
		return
	}
	if len(line) < prog.maxStrLen {
		line += strings.Repeat(" ", prog.maxStrLen-len(line))
	}
	prog.maxStrLen = len(line)
	fmt.Fprint(os.Stderr, line, "\r")
}
