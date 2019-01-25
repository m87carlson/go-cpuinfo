package main

import (
	"github.com/json-iterator/go"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type Cpu struct {
	Processor       int      `json:"-"`
	VendorId        string   `json:"vendor_id"`
	CpuFamily       int      `json:"cpu_family"`
	Model           int      `json:"model"`
	ModelName       string   `json:"mode_name"`
	Stepping        int      `json:"stepping"`
	CpuMhz          float32  `json:"cpu_mhz"`
	CacheSize       string   `json:"cache_size"`
	PhysicalId      int      `json:"physical_id"`
	Siblings        int      `json:"siblings"`
	CoreId          int      `json:"core_id"`
	CpuCores        int      `json:"cpu_cores"`
	Apicid          int      `json:"apicid"`
	InitialApicid   int      `json:"initial_apicid"`
	Fpu             string   `json:"fpu"`
	FpuException    string   `json:"fpu_exception"`
	CpuidLevel      int      `json:"cpuid_level"`
	Wp              string   `json:"wp"`
	Flags           []string `json:"flags"`
	Bogomips        float32  `json:"bogomips"`
	ClflushSize     int      `json:"clflush_size"`
	CacheAlignment  int      `json:"cache_alignment"`
	AddressSizes    string   `json:"address_sizes"`
	PowerManagement string   `json:"power_management"`
}

type CpuInfo struct {
	Cpus  map[int]Cpu `json:"cpu"`
	Total int         `json:"total"`
	Real  int         `json:"real"`
	Cores int         `json:"cores"`
}

func main() {
	http.HandleFunc("/", Info)
	http.HandleFunc("/status", HealthCheck)
	http.ListenAndServe(":8080", nil)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"status": "okay"}`)
}

func GetReal(c *CpuInfo) int {
	realProcessors := 0
	for _, v := range c.Cpus {
		if realProcessors < v.PhysicalId {
			realProcessors = v.PhysicalId
		}
	}
	return realProcessors + 1
}

func GetSiblings(c *CpuInfo) int {
	siblings := 0
	for _, v := range c.Cpus {
		siblings = v.Siblings
	}
	return siblings
}

func GetCpuCores(c *CpuInfo) int {
	cpuCores := 0
	for _, v := range c.Cpus {
		cpuCores = v.CpuCores
	}
	return cpuCores
}

func ParseSystemCpus(f string, c *CpuInfo) {
	cpufile, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}

	cpu := Cpu{}
	cpuArray := make(map[int]Cpu)

	lines := strings.Split(string(cpufile), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			cpuArray[cpu.Processor] = cpu
			continue
		}

		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])

		switch key {
		case "processor":
			cpu.Processor, _ = strconv.Atoi(value)
		case "vendor_id":
			cpu.VendorId = value
		case "cpu family":
			cpu.CpuFamily, _ = strconv.Atoi(value)
		case "model":
			cpu.Model, _ = strconv.Atoi(value)
		case "model name":
			cpu.ModelName = value
		case "stepping":
			cpu.Stepping, _ = strconv.Atoi(value)
		case "cpu MHz":
			f, _ := strconv.ParseFloat(value, 32)
			cpu.CpuMhz = float32(f)
		case "cache size":
			cpu.CacheSize = value
		case "physical id":
			cpu.PhysicalId, _ = strconv.Atoi(value)
		case "siblings":
			cpu.Siblings, _ = strconv.Atoi(value)
		case "core id":
			cpu.CoreId, _ = strconv.Atoi(value)
		case "cpu cores":
			cpu.CpuCores, _ = strconv.Atoi(value)
		case "apicid":
			cpu.Apicid, _ = strconv.Atoi(value)
		case "initial apicid":
			cpu.InitialApicid, _ = strconv.Atoi(value)
		case "fpu":
			cpu.Fpu = value
		case "fpu_exception":
			cpu.FpuException = value
		case "cpuid level":
			cpu.CpuidLevel, _ = strconv.Atoi(value)
		case "wp":
			cpu.Wp = value
		case "flags":
			cpu.Flags = strings.Split(value, " ")
		case "bogomips":
			f, _ := strconv.ParseFloat(value, 32)
			cpu.Bogomips = float32(f)
		case "clflush size":
			cpu.ClflushSize, _ = strconv.Atoi(value)
		case "cache_alignment":
			cpu.CacheAlignment, _ = strconv.Atoi(value)
		case "address sizes":
			cpu.AddressSizes = value
		case "power management":
			cpu.PowerManagement = value
		}
	}
	c.Cpus = cpuArray
}

func Info(w http.ResponseWriter, r *http.Request) {
	c, t, i := 0, 0, 0

	cpuInfo := CpuInfo{}
	ParseSystemCpus("/proc/cpuinfo", &cpuInfo)

	cpuInfo.Real = GetReal(&cpuInfo)
	for i < cpuInfo.Real {
		t = t + GetSiblings(&cpuInfo)
		c = c + GetCpuCores(&cpuInfo)
		i++
	}
	cpuInfo.Total = t
	cpuInfo.Cores = c

	js, err := jsoniter.Marshal(cpuInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
