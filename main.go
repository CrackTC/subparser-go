package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func getConfPath() string {
	confPath := os.Getenv("CONFIG_PATH")
	if confPath != "" {
		return confPath
	}

	exe, err := os.Executable()
	if err != nil {
		log.Fatalln("[error] cannot get executable path:", err)
	}

	confPath = filepath.Join(filepath.Dir(exe), "config/config.yml")
	return confPath
}

func getPort() int {
	const defaultPort = 4447

	portString := os.Getenv("PORT")
	if port, err := strconv.Atoi(portString); err != nil {
		return defaultPort
	} else if port <= 0 || port >= 65536 {
		return defaultPort
	} else {
		return port
	}
}

func main() {
	confPath := getConfPath()
	port := getPort()

	http.HandleFunc("/", newHandler(confPath))

	log.Println("[info] listening on port", port)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
