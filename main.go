/*
   2020年4月17日 14:19:02 by elvis
   kubernetes nginx configmap reload

*/
package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
)

const (
	nginxProcessName     = "nginx"
	defaultNginxConfPath = "/etc/nginx"
	defaultNginxPidPath  = "/usr/local/nginx/logs/nginx.pid"
	watchPidEnvVarName   = "WATCH_NGINX_PID_PATH"
	watchPathEnvVarName  = "WATCH_NGINX_CONF_PATH"
)

var stderrLogger = log.New(os.Stderr, "error: ", log.Lshortfile)
var stdoutLogger = log.New(os.Stdout, "", log.Lshortfile)

func getMasterNginxPid() (int, error) {

	nginxPidPath, ok := os.LookupEnv(watchPidEnvVarName)
	if !ok {
		nginxPidPath = defaultNginxPidPath
	}

	//获取nginx的进程ID
	pfile, err := os.Open(nginxPidPath)
	defer pfile.Close()

	pidData, _ := ioutil.ReadAll(pfile)
	masterNginxPid := string(pidData)
	masterNginxPid = strings.Replace(pid, "\n", "", -1)
	return strconv.Atoi(masterNginxPid), nil
}

func signalNginxReload(pid int) error {
	stdoutLogger.Printf("signaling master nginx process (pid: %d) -> SIGHUP\n", pid)
	nginxProcess, nginxProcessErr := os.FindProcess(pid)

	if nginxProcessErr != nil {
		return nginxProcessErr
	}

	return nginxProcess.Signal(syscall.SIGHUP)
}

func main() {
	watcher, watcherErr := fsnotify.NewWatcher()
	if watcherErr != nil {
		stderrLogger.Fatal(watcherErr)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					if filepath.Base(event.Name) == "..data" {
						stdoutLogger.Println("config map updated")

						nginxPid, nginxPidErr := getMasterNginxPid()
						if nginxPidErr != nil {
							stderrLogger.Printf("getting master nginx pid failed: %s", nginxPidErr.Error())

							continue
						}

						if err := signalNginxReload(nginxPid); err != nil {
							stderrLogger.Printf("signaling master nginx process failed: %s", err)
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				stderrLogger.Printf("received watcher.Error: %s", err)
			}
		}
	}()

	pathToWatch, ok := os.LookupEnv(watchPathEnvVarName)
	if !ok {
		pathToWatch = defaultNginxConfPath
	}

	stdoutLogger.Printf("adding path: `%s` to watch\n", pathToWatch)

	if err := watcher.Add(pathToWatch); err != nil {
		stderrLogger.Fatal(err)
	}
	<-done
}
