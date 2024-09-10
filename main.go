package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func main() {
	log.Println("Starting telemetry service...")

	r := gin.Default()

	// Initialize routes
	initializeRoutes(r)

	log.Println("Routes initialized. Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initializeRoutes(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		log.Println("Received request to /")
		c.JSON(http.StatusOK, gin.H{
			"system": "ok",
		})
	})

	r.GET("/cpu-info", getCPUInfo)
	r.GET("/system-load", getSystemLoad)
	r.GET("/gpu-info", getGPUInfo)
}

func getCPUInfo(c *gin.Context) {
	cpuUtil, err := cpu.Percent(0, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get CPU utilization"})
		return
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get memory information"})
		return
	}

	response := gin.H{
		"cpu_utilization":    fmt.Sprintf("%.2f%%", cpuUtil[0]),
		"memory_utilization": fmt.Sprintf("%.2f%%", memInfo.UsedPercent),
		"total_memory":       fmt.Sprintf("%d MB", memInfo.Total/1024/1024),
		"used_memory":        fmt.Sprintf("%d MB", memInfo.Used/1024/1024),
		"free_memory":        fmt.Sprintf("%d MB", memInfo.Free/1024/1024),
	}

	c.JSON(http.StatusOK, response)
}

func getSystemLoad(c *gin.Context) {
	cpuUtil, err := cpu.Percent(0, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get CPU utilization"})
		return
	}

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get memory information"})
		return
	}

	load := (cpuUtil[0] + memInfo.UsedPercent) / 2

	gpuInfo, err := getGPUInfoInternal()
	if err != nil {
		log.Printf("Error getting GPU info: %v", err)
	} else if gpuInfo != nil {
		load = (load + gpuInfo.GPUUtilization) / 2
	}

	c.JSON(http.StatusOK, gin.H{
		"load": fmt.Sprintf("%.2f%%", load),
	})
}

func getGPUInfoInternal() (*GPUInfo, error) {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("Failed to initialize NVML")
	}
	defer nvml.Shutdown()

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("Unable to get device count")
	}

	if count == 0 {
		return nil, nil
	}

	device, ret := nvml.DeviceGetHandleByIndex(0)
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("Unable to get device handle")
	}

	memInfo, ret := device.GetMemoryInfo()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("Unable to get memory info")
	}

	utilization, ret := device.GetUtilizationRates()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("Unable to get utilization rates")
	}

	return &GPUInfo{
		MemoryUtilization: float64(memInfo.Used) / float64(memInfo.Total) * 100,
		GPUUtilization:    float64(utilization.Gpu),
	}, nil
}

type GPUInfo struct {
	MemoryUtilization float64
	GPUUtilization    float64
}

func getGPUInfo(c *gin.Context) {
	gpuInfo, err := getGPUInfoInternal()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if gpuInfo == nil {
		c.JSON(http.StatusOK, gin.H{"message": "No GPU information available"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"memory_utilization": fmt.Sprintf("%.2f%%", gpuInfo.MemoryUtilization),
		"gpu_utilization":    fmt.Sprintf("%.2f%%", gpuInfo.GPUUtilization),
	})
}
