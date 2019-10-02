package ui

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"

	"github.com/nanu-c/textsecure"
	"github.com/nanu-c/textsecure-qml/app/config"
	"github.com/nanu-c/textsecure-qml/app/settings"
	"github.com/nanu-c/textsecure-qml/app/store"
	"github.com/nanu-c/textsecure-qml/app/webserver"
	log "github.com/sirupsen/logrus"
	"github.com/zserge/lorca"
)

// var Win *qml.Window
// var Engine *qml.Engine

func GroupUpdateMsg(tels []string, title string) string {
	s := ""
	if len(tels) > 0 {
		for _, t := range tels {
			s += store.TelToName(t) + ", "
		}
		s = s[:len(s)-2] + " joined the group. "
	}

	return s + "Title is now '" + title + "'."
}
func RegistrationDone() {
	log.Infoln("Registered")
	// Win.Root().Call("registered")
	textsecure.WriteConfig(config.ConfigFile, config.Config)
	settings.SettingsModel.Registered = true
	webserver.RegistrationDone()
}
func SetComponent() error {
	// component, err := Engine.LoadFile(config.MainQml)
	// if err != nil {
	// 	log.Println(err)
	// 	return err
	// }
	// Win = component.CreateWindow(nil)
	return nil
}
func SetEngine() {
	// Engine = qml.NewEngine()
}
func InitModels() {
	var err error
	settings.SettingsModel, err = settings.LoadSettings()
	if err != nil {
		log.Println(err)
	} else {
		if settings.SettingsModel.Registered {
			log.Debugf("Already registered")
		}
	}
	// Engine.Context().SetVar("contactsModel", store.ContactsModel)
	// Engine.Context().SetVar("settingsModel", settings.SettingsModel)
	// Engine.Context().SetVar("sessionsModel", store.SessionsModel)
	// textsecure.LinkedDevices()
	// Engine.Context().SetVar("linkedDevicesModel", store.LinkedDevicesModel)
	// Engine.Context().SetVar("storeModel", store.DS)

	go store.UpdateTimestamps()
}
func RunUi(e string) {
	// cmd := exec.Command("webapp-container", "http://[::1]:8080/")
	if e == "ut" || e == "me" {
		runUIUbuntuTouch(e)
	} else if e == "lorca" {
		fmt.Println("start lorca")
		ui, err := lorca.New("", "", 480, 720, "--hide-scrollbars")
		if err != nil {
			log.Debugln("lorca error")
			log.Fatal(err)
		}
		defer ui.Close()

		// A simple way to know when UI is ready (uses body.onload event in JS)
		ui.Bind("start", func() {
			log.Println("UI is ready")
		})
		ui.Load(fmt.Sprintf("http://localhost:9080"))

		// Wait until the interrupt signal arrives or browser window is closed
		sigc := make(chan os.Signal)
		signal.Notify(sigc, os.Interrupt)
		select {
		case <-sigc:
		case <-ui.Done():
		}

		log.Println("exiting...")
	}
}
func runUIUbuntuTouch(e string) {
	var cmd *exec.Cmd
	log.Infof("Axolotl-gui starting for sys: %v", config.Gui)

	if config.Gui == "ut" {
		cmd = exec.Command("qmlscene", "--scaling", "guis/qml/ut/MainUt.qml")
	} else if config.Gui == "me" {
		cmd = exec.Command("/home/nanu/Qt/5.13.0/gcc_64/bin/qmlscene", "--scaling", "guis/qml/Main.qml")

	} else {
		cmd = exec.Command("qmlscene", "--scaling", "guis/qml/Main.qml")
	}
	var stdout, stderr []byte
	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() failed with '%s'\n", err)
	}

	// cmd.Wait() should be called only after we finish reading
	// from stdoutIn and stderrIn.
	// wg ensures that we finish
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
		wg.Done()
	}()

	stderr, errStderr = copyAndCapture(os.Stderr, stderrIn)

	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	if errStdout != nil || errStderr != nil {
		log.Fatal("failed to capture stdout or stderr\n")
	}
	outStr, errStr := string(stdout), string(stderr)
	log.Infof("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
	log.Infof("Axolotl-gui finished with error: %v", err)
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
