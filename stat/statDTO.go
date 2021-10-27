package stat

type StatDTO struct {
    FreeMemory uint64 `json:"freeMemory"`
    HostURL string `json:"hostURL"` 
	Cpu float64 `json:"cpu"` 	
}

