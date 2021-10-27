package main

import (
	"encoding/json"
	"fmt"
	"log"
	"nagarunner/docker"
	"nagarunner/stat"
	"net/http"
	"time"

	"github.com/magiconair/properties"
)

func health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
	fmt.Println("Endpoint Hit: health")
	// json.NewEncoder(w).Encode(Articles)
}

func invoke(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		fmt.Println("Endpoint Hit: invoke")
		var p stat.InvokeDTO
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		containerId := docker.StartOrRun(p.ImageName)
		fmt.Println("Invoking with container Id " + containerId)
		res := docker.Attach2ContainterAndInvoke(containerId, p.Payload)
		fmt.Fprintln(w, res)

	} else {
		fmt.Fprintf(w, "Only POST is supported")
	}
}

// 190.79 mekadem for 60, 185 in 60.5
func handleRequests() {
	http.HandleFunc("/health", health)
	http.HandleFunc("/invoke", invoke)
	log.Fatal(http.ListenAndServe(":10000", nil))
}

func handleStatTimer() {
	ticker := time.NewTicker(50000 * time.Millisecond)
	for range ticker.C {
		fmt.Println("Tick! sending stats...")
		stat.SendStat()
	}
}

/**
1. Load available containers in this machine (once)
2. A scheduled task to stop long running and unused containers
**/
func handleStopContainers() {
	docker.LoadStoppedConatinersInfo()
	//
	ticker := time.NewTicker(10000 * time.Millisecond)
	for range ticker.C {
		//fmt.Println("Tick! Killing old containers...")
		docker.StopOldContainers()
	}
}

func main() {
	// init from a file
	p := properties.MustLoadFile("./config.properties", properties.UTF8)
	fmt.Println("GW address: " + p.MustGetString("gateway"))
	go handleStatTimer()
	go handleStopContainers()
	handleRequests()

}
