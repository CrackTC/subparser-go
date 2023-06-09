package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"sora.zip/subparser-go/config"
)

func readConf(confPath string) config.Config {
	bytes, err := os.ReadFile(confPath)
	if err != nil {
		log.Fatalln("[error] cannot read config file:", err)
	}

	conf, err := config.LoadString(string(bytes))
	if err != nil {
		log.Fatalln("[error] cannot parse config file:", err)
	}

	return conf
}

func watchConf(confPath string, conf *atomic.Value) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln("[error] cannot create file watcher:", err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if ok == false {
					return
				}

				if event.Has(fsnotify.Write) {
					if event.Name == confPath {
						log.Println("[info] detected config change")

						conf.Store(readConf(confPath))
					}
				}

			case err, ok := <-watcher.Errors:
				if ok == false {
					return
				}

				log.Fatalln("[error] file watcher reported an error:", err)
			}
		}
	}()

	err = watcher.Add(filepath.Dir(confPath))
	if err != nil {
		log.Fatalln("[error] cannot watch config file:", err)
	}

	<-make(chan struct{})
}

func newHandler(confPath string) func(http.ResponseWriter, *http.Request) {
	var conf atomic.Value
	conf.Store(readConf(confPath))

	go watchConf(confPath, &conf)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Header().Add("Allow", "GET")
			return
		}

		query := r.URL.Query()

		if query.Has("url") == false {
			w.Write([]byte("usage: /?url=<url>"))
			return
		}

		url := query.Get("url")

		resp, err := http.Get(url)
		if err != nil {
			w.Write([]byte(err.Error()))
			log.Println("[error] cannot get remote config:", err)
			return
		}
		defer resp.Body.Close()

		remoteConf, err := config.Load(resp.Body.(io.Reader))
		if err != nil {
			w.Write([]byte(err.Error()))
			log.Println("[error] cannot parse remote config:", err)
			return
		}

		remoteConf.Merge(conf.Load().(config.Config))

		proxyNames := make([]string, len(remoteConf["proxies"].([]any)))
		for i, value := range remoteConf["proxies"].([]any) {
			proxyNames[i] = value.(config.Config)["name"].(string)
		}

		for _, value := range remoteConf["proxy-groups"].([]any) {
			group := value.(config.Config)
			if _, ok := group["proxies"]; ok == false {
				if _, ok := group["use"]; ok == false {
					group["proxies"] = proxyNames
				}
			}
		}

		str, err := remoteConf.String()
		if err != nil {
			w.Write([]byte(err.Error()))
			log.Println("[error] cannot parse remote conf to string:", err)
			return
		}
		w.Write([]byte(str))
	}
}
