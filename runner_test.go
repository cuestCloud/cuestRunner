package main

import (
	"nagarunner/docker"
	"testing"
)

func TestDocker(t *testing.T) {
	docker.LoadStoppedConatinersInfo()
	containerId := docker.StartOrRun("localhost:5000/adder")
	docker.Attach2ContainterAndInvoke(containerId, "33")
	//
	//docker.StopOldContainers()
	//docker.StartOrRun("localhost:5000/adder")
}
