package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const IDLE_TIME = 30 // 30 secs

/**
	Load available containers on this machine
**/
func LoadStoppedConatinersInfo() {
	LoadStoppedContainersIntoMap()
}

/**
 Stopping containers with functions that were not invoked in the past IDLE_TIME seconds
**/
func StopOldContainers() {
	//now := time.Now().UnixNano() / int64(time.Millisecond)
	now := time.Now().Unix()
	// Get running containers
	runners := GetRunningContainers()
	// collect containers to stop
	for _, r := range runners {
		// Stop them
		fmt.Println("Now is " + fmt.Sprint(now) + " and last used+idle " + fmt.Sprint(r.lastUsed+IDLE_TIME))
		if now > r.lastUsed+IDLE_TIME {
			StopContainer(&r)
		}
	}
}

/**
	Attach to a running container
**/
func Attach2ContainterAndInvoke(containerId string, payload string) string {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	waiter, errAttach := cli.ContainerAttach(ctx, containerId, types.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})
	if errAttach != nil {
		// ContainerAttach returns an ErrPersistEOF (connection closed)
		// means server met an error and put it in Hijacked connection
		// keep the error and read detailed error message from hijacked connection later
		panic(errAttach)
	}
	defer waiter.Close()
	// Test

	fmt.Fprintf(waiter.Conn, payload+"\n")
	line, _, err := waiter.Reader.ReadLine()

	if err != nil { // check errorpanic(err)
		panic(err)
	}
	ret := string(line[8:])
	fmt.Printf("Response: %s \n", ret)
	return ret
}

/**
	start a stopped container (if available) or run a new one (unless
	already running, of course). return the running container id
**/
func StartOrRun(imageName string) string {
	fmt.Println("dockerHandler::StartOrRun")
	// Do we have a running container for that image
	container, err := GetRunningContainerByImageName(imageName)
	if err == nil {
		fmt.Println("Found running container for " + imageName)
		return container.containerId
	}
	// Do we have a stopped container for that image
	container, err = GetStoppedContainerByImageName(imageName)
	if err == nil {
		fmt.Println("Found a stopped container for " + imageName)
		StartAStoppedContainer(&container)
		fmt.Println("last used after is " + fmt.Sprint(container.lastUsed))
		return container.containerId
	}
	// If not, use docker run to get the image from the repository and
	// start a new container
	fmt.Println("Running a NEW container for " + imageName)
	return RunContainer(imageName)
}

/*func doDocker() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	reader, err := cli.ImagePull(ctx, "localhost:5000/adder", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "localhost:5000/adder",
		Cmd:   []string{"5"},
	}, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
}*/
