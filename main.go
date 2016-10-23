package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
)

type Trigger struct {
	Name    string `json:name,omitempty`
	URL     string `json:url`
	Command string `json:command`
}

var (
	port          = 5138
	logLevel      = ""
	configPath    = ""
	triggers      = []*Trigger{}
	gitCommitHash = "non-git"
	gitBranch     = "non-git"
	compileDate   = "not set"
)

func main() {
	// parse flags
	flag.IntVar(&port, "p", port, "Port to use (<1024 needs according priviledges)")
	flag.StringVar(&logLevel, "l", "info", "LogLevel (fatal, error, warn, info, debug")
	flag.StringVar(&configPath, "c", "remote-trigger.conf", "Path to config file")
	flag.Parse()

	// configure logging
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	var level log.Level
	switch logLevel {
	case "fatal":
		level = log.FatalLevel
	case "error":
		level = log.ErrorLevel
	case "warn":
		level = log.WarnLevel
	case "info":
		level = log.InfoLevel
	case "debug":
		level = log.DebugLevel
	}
	log.SetLevel(level)

	// print version identifier
	fmt.Printf(">> remote-trigger %s (branch %s, compiled at %s)\n\n", gitCommitHash, gitBranch, compileDate)

	// initialize
	loadTriggers()
	http.HandleFunc("/", handler)
	log.WithFields(log.Fields{
		"port": port,
		"time": time.Now(),
	}).Info("Starting server...")
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in handler", r)
		}
	}()

	triggerValue := r.URL.Path[1:]

	// key existing?
	trigger, err := getTrigger(triggerValue)
	if err != nil {
		// 404
		w.WriteHeader(404)
		return
	}

	cmd := exec.Command(trigger.Command)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error creating StdoutPipe for Cmd")
		return
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s\n", scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error starting Cmd")
		return
	}

	err = cmd.Wait()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error waiting for Cmd")
		return
	}

	fmt.Fprintf(w, "Access to %s", triggerValue)
	log.WithFields(log.Fields{
		"url":  triggerValue,
		"name": trigger.Name,
		"time": time.Now(),
	}).Info("URL triggered")
}

func loadTriggers() {
	absPath, _ := filepath.Abs(configPath)
	log.WithFields(log.Fields{
		"path":          configPath,
		"absolute path": absPath,
	}).Info("Loading triggers from config file...")
	config, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.WithFields(log.Fields{
			"path":          configPath,
			"absolute path": absPath,
			"err":           err,
		}).Fatal("Config file not readable")
	}
	err = json.Unmarshal(config, &triggers)
	if err != nil {
		log.WithFields(log.Fields{
			"path":          configPath,
			"absolute path": absPath,
			"err":           err,
		}).Fatal("Config file not parseable (JSON format)")
	}

	// log the triggers found
	for _, t := range triggers {
		log.WithFields(log.Fields{
			"name":    t.Name,
			"url":     t.URL,
			"command": t.Command,
		}).Debug("Trigger loaded")
	}
	log.WithFields(log.Fields{
		"count": len(triggers),
	}).Info("Triggers loaded")
}

func getTrigger(url string) (*Trigger, error) {
	for _, trigger := range triggers {
		if trigger.URL == url {
			return trigger, nil
		}
	}
	return nil, errors.New("Trigger not found")
}
