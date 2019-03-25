//go:generate go run -tags generate gen.go

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/schollz/jsonstore"
	"github.com/zserge/lorca"
)

var ks = new(jsonstore.JSONStore)
var configFile string

func initStore() {
	u, e := user.Current()
	if e == nil {
		h := u.HomeDir
		path := filepath.Join(h, "Library/Application Support/lorca-boilerplate")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.Mkdir(path, os.ModePerm)
		}
		configFile = filepath.Join(path, "config.json.gz")
	}
}

func main() {
	initStore()
	env := flag.String("env", "production", "")
	flag.Parse()
	log.Println(*env)

	var args []string
	args = append(args, "--disable-web-security")
	args = append(args, "--enable-file-cookies")
	if runtime.GOOS == "linux" {
		args = append(args, "--class=Lorca")
	}
	ui, err := lorca.New("", "", 1024, 728, args...)
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	_ = ui.Bind("start", func() {
		log.Println("UI is ready")
	})

	// Create and bind Go object to the UI

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	go http.Serve(ln, http.FileServer(FS))
	if *env == "dev" {
		ui.Load("http://localhost:3000")
	} else {
		ui.Load(fmt.Sprintf("http://%s", ln.Addr()))
	}

	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt)
	select {
	case <-sigc:
	case <-ui.Done():
	}

	log.Println("exiting...")
}
