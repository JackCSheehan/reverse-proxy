package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/JackCSheehan/reverse-proxy/proxy"
)

func main() {
	if len(os.Args) < 2 {
		panic("Usage: ./reverse-proxy <path/to/config.yaml>")
	}

	logFile, err := os.OpenFile("reverseproxy.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0775)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	// It's useful to see log output to the console and to a file, so we'll use a multi-writer to
	// write to both the file and to stdout
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)


	// Path to the config YAML
	configFilePath := os.Args[1]

	config, err := proxy.NewConfig(configFilePath)
	if err != nil {
		panic(err)
	}

	proxy.RegisterMetricsEndpoint()

	proxy.RegisterEndpoints(config)
	log.Println("Registered endpoints")

	log.Printf("Starting reverse proxy on port %d\n", config.Port)
	err = http.ListenAndServe(":"+strconv.Itoa(config.Port), nil)
	if err != nil {
		panic(err)
	}
}
