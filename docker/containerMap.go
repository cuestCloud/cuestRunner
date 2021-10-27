package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// Image Name -> container info
var stoppedContainers = make(map[string]ContainerInfo)
var runningContainers = make(map[string]ContainerInfo)

func LoadStoppedContainersIntoMap() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		panic(err)
	}
	// Only add stopped containers
	for _, container := range containers {
		if container.State == "exited" {
			stoppedContainers[container.Image] = ContainerInfo{
				containerId: container.ID,
				imageName:   container.Image,
				// we don't care here, as it is already stopped
				lastUsed: time.Now().Unix(),
			}
		}
	}
}

/**
  Returns a new array with references to the running containers
**/
func GetRunningContainers() []ContainerInfo {
	if len(runningContainers) == 0 {
		return nil
	}
	v := make([]ContainerInfo, 0, len(runningContainers))
	for _, value := range runningContainers {
		v = append(v, value)
	}
	return v
}

/**
 Try to find and return info on a running container, updating its 'lastUsed' field
**/
func GetRunningContainerByImageName(imageName string) (ContainerInfo, error) {
	fmt.Println("Looking for a RUNNING container of the image: " + imageName)
	info, ok := runningContainers[imageName]
	if ok {
		info.lastUsed = time.Now().Unix()
		runningContainers[imageName] = info
		return info, nil
	}
	return info, errors.New("no running container found")
}

/**
 Try to find and return info on a stopped container
**/
func GetStoppedContainerByImageName(imageName string) (ContainerInfo, error) {
	fmt.Println("Looking for a STOPPED container of the image: " + imageName)
	info, ok := stoppedContainers[imageName]
	if ok {
		return info, nil
	}
	return info, errors.New("no container found")
}

/**
	Stop a running container and update the maps accordingly
**/
func StopContainer(containerInfo *ContainerInfo) {

	// verify it's in the running list? or simply fail?
	fmt.Println("Stopping container " + containerInfo.containerId + " image " + containerInfo.imageName)
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStop(ctx, containerInfo.containerId, nil); err != nil {
		panic(err)
	}
	// remove from running
	delete(runningContainers, containerInfo.imageName)
	// add to stopped
	stoppedContainers[containerInfo.imageName] = *containerInfo
}

/**
	Start a running container and update the maps accordingly
**/
func StartAStoppedContainer(containerInfo *ContainerInfo) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, containerInfo.containerId, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	// remove from stopped
	delete(stoppedContainers, containerInfo.imageName)
	// add to running
	now := time.Now().Unix()
	containerInfo.lastUsed = now
	runningContainers[containerInfo.imageName] = *containerInfo
}

/**
	Download an image and run a container from it. Updates the maps accordingly.
**/
func RunContainer(imageName string) string {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, out)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        imageName,
		Tty:          false,
		AttachStdin:  true,
		AttachStdout: true,
		OpenStdin:    true,
	}, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	ci := ContainerInfo{
		containerId: resp.ID,
		imageName:   imageName,
		lastUsed:    time.Now().Unix(),
	}
	runningContainers[imageName] = ci

	return resp.ID
}
