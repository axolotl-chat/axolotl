package worker

import (
	"io"
	"os"
	"os/exec"
	"sync"

	log "github.com/sirupsen/logrus"
)

func RunRustBackend() {
	var cmd *exec.Cmd
	log.Infoln("[axolotl] Starting crayfish-backend")
	if _, err := os.Stat("./crayfish"); err == nil {
		cmd = exec.Command("./crayfish")
	} else {
		cmd = exec.Command("./backend/target/debug/crayfish")
	}
	var stdout, stderr []byte
	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		log.Fatalf("[axolotl] Starting crayfish-backend cmd.Start() failed with '%s'\n", err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
		wg.Done()
	}()

	stderr, errStderr = copyAndCapture(os.Stderr, stderrIn)

	wg.Wait()

	err = cmd.Wait()
	if errStdout != nil || errStderr != nil {
		log.Fatal("[axolotl] failed to capture stdout or stderr\n")
	}
	outStr, errStr := string(stdout), string(stderr)
	log.Infof("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
	log.Infof("[axolotl] Crayfish-backend finished with error: %v", err)

}
func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}

func StartWebsocket() {

}
