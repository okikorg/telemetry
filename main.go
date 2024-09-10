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
	r := gin.Default()

	r.Use(gin.Logger())

	r.GET("/", func(c *gin.Context) {
		log.Println("Received request to /")
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})

	r.GET("/cpu-info", func(c *gin.Context) {
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
	})

	r.GET("/system-load", func(c *gin.Context) {
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

		gpuInfo := getGPUInfo()
		if gpuInfo != nil {
			load = (load + gpuInfo.GPUUtilization) / 2
		}

		c.JSON(http.StatusOK, gin.H{
			"load": fmt.Sprintf("%.2f%%", load),
		})
	})

	r.GET("/gpu-info", func(c *gin.Context) {
		gpuInfo := getGPUInfo()
		if gpuInfo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get GPU information"})
			return
		}

		response := gin.H{
			"gpu_memory_utilization": fmt.Sprintf("%.2f%%", gpuInfo.MemoryUtilization),
			"gpu_utilization":        fmt.Sprintf("%.2f%%", gpuInfo.GPUUtilization),
		}

		c.JSON(http.StatusOK, response)
	})

	log.Println("Starting server on :8080")
	r.Run(":8080")
}

type GPUInfo struct {
	MemoryUtilization float64
	GPUUtilization    float64
}

func getGPUInfo() *GPUInfo {
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		log.Println("Failed to initialize NVML")
		return nil
	}
	defer nvml.Shutdown()

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		log.Println("Unable to get device count")
		return nil
	}

	if count == 0 {
		return nil
	}

	device, ret := nvml.DeviceGetHandleByIndex(0)
	if ret != nvml.SUCCESS {
		log.Println("Unable to get device handle")
		return nil
	}

	memInfo, ret := device.GetMemoryInfo()
	if ret != nvml.SUCCESS {
		log.Println("Unable to get memory info")
		return nil
	}

	utilization, ret := device.GetUtilizationRates()
	if ret != nvml.SUCCESS {
		log.Println("Unable to get utilization rates")
		return nil
	}

	return &GPUInfo{
		MemoryUtilization: float64(memInfo.Used) / float64(memInfo.Total) * 100,
		GPUUtilization:    float64(utilization.Gpu),
	}
}
