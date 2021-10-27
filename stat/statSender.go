package stat

import (
    "github.com/gonutz/w32/v2"
    "fmt"
    "time"	
	"github.com/magiconair/properties"
	"bytes"
    "encoding/json"   
	"net/http"
	"net"
	"os"
	"github.com/mackerelio/go-osstat/memory"
)

 var gwURL = properties.MustLoadFile("./config.properties", properties.UTF8).MustGetString("gateway")	
	
func SendStat() {		
	fmt.Printf("Sending Stats\n")	
	//Encode the data
	var st StatDTO
	st.FreeMemory = getFreeMemory()
	st.Cpu = getFreeCPU()
	st.HostURL = getMyIp()
	fmt.Println(st)
    var jsonData []byte
	jsonData, err := json.Marshal(st)
    responseBody := bytes.NewBuffer(jsonData)
	fmt.Println(string(jsonData))
	//Leverage Go's HTTP Post function to make request
    _, err = http.Post(gwURL+"/stat", "application/json", responseBody)
	//Handle Error
    if err != nil {
      fmt.Printf("An Error Occured %v", err)
    }
    //defer resp.Body.Close()	
}

func getFreeMemory() uint64 {
	memory, err := memory.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return 0
	}	
	return memory.Free
}

func getFreeCPU() float64 {
	var idle, kernel, user w32.FILETIME

    w32.GetSystemTimes(&idle, &kernel, &user)
    idleFirst   := idle.DwLowDateTime   | (idle.DwHighDateTime << 32)
    kernelFirst := kernel.DwLowDateTime | (kernel.DwHighDateTime << 32)
    userFirst   := user.DwLowDateTime   | (user.DwHighDateTime << 32)

    time.Sleep(time.Second)

    w32.GetSystemTimes(&idle, &kernel, &user)
    idleSecond   := idle.DwLowDateTime   | (idle.DwHighDateTime << 32)
    kernelSecond := kernel.DwLowDateTime | (kernel.DwHighDateTime << 32)
    userSecond   := user.DwLowDateTime   | (user.DwHighDateTime << 32)

    totalIdle   := float64(idleSecond - idleFirst)
    totalKernel := float64(kernelSecond - kernelFirst)
    totalUser   := float64(userSecond - userFirst)
    totalSys    := float64(totalKernel + totalUser)
    return  (totalSys - totalIdle) * 100 / totalSys    
}

func getMyIp() string {

	conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        fmt.Printf("An Error Occured %v", err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP.String()+":10000"
}

