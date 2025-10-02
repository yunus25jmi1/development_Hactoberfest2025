package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // pending, in-progress, completed
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SystemMetrics struct {
	CPUUsage     float64
	Memory       MemoryInfo
	Disk         DiskInfo
	Network      NetworkInfo
	System       SystemInfo
	LastUpdate   time.Time
}

type MemoryInfo struct {
	UsedPercent float64
	Used        uint64
	Total       uint64
}

type DiskInfo struct {
	UsedPercent float64
	Used        uint64
	Total       uint64
}

type NetworkInfo struct {
	BytesSent uint64
	BytesRecv uint64
	Count     int
}

type SystemInfo struct {
	OS       string
	Platform string
	Arch     string
	Cores    int
	Uptime   time.Duration
}

type TaskManager struct {
	Tasks []Task
}

func main() {
	manager := &TaskManager{}
	manager.loadTasks()
	
	fmt.Println("=== TaskTracker CLI with System Metrics ===")
	
	for {
		fmt.Println("\n=== Main Menu ===")
		fmt.Println("1. Task Management")
		fmt.Println("2. System Metrics")
		fmt.Println("3. Exit")
		fmt.Print("Choose option (1-3): ")
		
		var choice int
		fmt.Scanf("%d", &choice)
		
		switch choice {
		case 1:
			taskMenu(manager)
		case 2:
			metricsMenu()
		case 3:
			manager.saveTasks()
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func taskMenu(manager *TaskManager) {
	for {
		fmt.Println("\n=== Task Management ===")
		fmt.Println("1. Add Task")
		fmt.Println("2. List Tasks")
		fmt.Println("3. Update Task")
		fmt.Println("4. Delete Task")
		fmt.Println("5. Back to Main Menu")
		fmt.Print("Choose option (1-5): ")
		
		var choice int
		fmt.Scanf("%d", &choice)
		
		switch choice {
		case 1:
			manager.addTask()
		case 2:
			manager.listTasks()
		case 3:
			manager.updateTask()
		case 4:
			manager.deleteTask()
		case 5:
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func metricsMenu() {
	for {
		fmt.Println("\n=== System Metrics ===")
		fmt.Println("1. View Current Metrics")
		fmt.Println("2. Monitor Continuously")
		fmt.Println("3. Back to Main Menu")
		fmt.Print("Choose option (1-3): ")
		
		var choice int
		fmt.Scanf("%d", &choice)
		
		switch choice {
		case 1:
			viewMetrics()
		case 2:
			monitorMetrics()
		case 3:
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func (tm *TaskManager) addTask() {
	scanner := bufio.NewScanner(os.Stdin)
	
	fmt.Print("Task title: ")
	scanner.Scan()
	title := scanner.Text()
	
	fmt.Print("Description: ")
	scanner.Scan()
	description := scanner.Text()
	
	task := Task{
		ID:          len(tm.Tasks) + 1,
		Title:       title,
		Description: description,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	tm.Tasks = append(tm.Tasks, task)
	fmt.Println("Task added successfully!")
}

func (tm *TaskManager) listTasks() {
	if len(tm.Tasks) == 0 {
		fmt.Println("No tasks available")
		return
	}
	
	fmt.Println("\nYour Tasks:")
	for _, task := range tm.Tasks {
		status := getStatusSymbol(task.Status)
		fmt.Printf("[%d] %s %s\n", task.ID, status, task.Title)
		fmt.Printf("    Status: %s\n", task.Status)
		fmt.Printf("    Description: %s\n", task.Description)
		fmt.Printf("    Created: %s\n", task.CreatedAt.Format("2006-01-02 15:04"))
		fmt.Printf("    Updated: %s\n", task.UpdatedAt.Format("2006-01-02 15:04"))
		fmt.Println()
	}
}

func getStatusSymbol(status string) string {
	switch status {
	case "completed":
		return "✓"
	case "in-progress":
		return "→"
	default:
		return "○"
	}
}

func (tm *TaskManager) updateTask() {
	var id int
	fmt.Print("Enter task ID to update: ")
	fmt.Scanf("%d", &id)
	
	for i := range tm.Tasks {
		if tm.Tasks[i].ID == id {
			fmt.Print("New status (pending/in-progress/completed): ")
			var status string
			fmt.Scanf("%s", &status)
			
			switch status {
			case "pending", "in-progress", "completed":
				tm.Tasks[i].Status = status
				tm.Tasks[i].UpdatedAt = time.Now()
				fmt.Println("Task updated successfully!")
				return
			default:
				fmt.Println("Invalid status")
				return
			}
		}
	}
	
	fmt.Println("Task not found")
}

func (tm *TaskManager) deleteTask() {
	var id int
	fmt.Print("Enter task ID to delete: ")
	fmt.Scanf("%d", &id)
	
	for i, task := range tm.Tasks {
		if task.ID == id {
			tm.Tasks = append(tm.Tasks[:i], tm.Tasks[i+1:]...)
			fmt.Println("Task deleted!")
			return
		}
	}
	
	fmt.Println("Task not found")
}

func (tm *TaskManager) loadTasks() {
	data, err := os.ReadFile("tasks.json")
	if err != nil {
		return // File doesn't exist yet
	}
	
	json.Unmarshal(data, &tm.Tasks)
}

func (tm *TaskManager) saveTasks() {
	data, err := json.MarshalIndent(tm.Tasks, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling tasks: %v\n", err)
		return
	}
	if err := os.WriteFile("tasks.json", data, 0644); err != nil {
		fmt.Printf("Error writing tasks to file: %v\n", err)
	}
}

func viewMetrics() {
	metrics, err := getSystemMetrics()
	if err != nil {
		fmt.Printf("Error getting metrics: %v\n", err)
		return
	}
	
	printMetrics(metrics)
}

func monitorMetrics() {
	var duration int
	fmt.Print("Enter monitoring duration in seconds (0 for infinite): ")
	fmt.Scanf("%d", &duration)
	
	endTime := time.Now().Add(time.Duration(duration) * time.Second)
	
	if duration == 0 {
		endTime = time.Time{} // Infinite monitoring
	}
	
	for {
		if !endTime.IsZero() && time.Now().After(endTime) {
			break
		}
		
		metrics, err := getSystemMetrics()
		if err != nil {
			fmt.Printf("Error getting metrics: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}
		
		clearScreen()
		printMetrics(metrics)
		time.Sleep(2 * time.Second)
	}
}

func getSystemMetrics() (*SystemMetrics, error) {
	var metrics SystemMetrics
	
	// CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}
	metrics.CPUUsage = cpuPercent[0]
	
	// Memory info
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}
	metrics.Memory = MemoryInfo{
		UsedPercent: vmStat.UsedPercent,
		Used:        vmStat.Used,
		Total:       vmStat.Total,
	}
	
	// Disk info (root partition)
	diskStat, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("failed to get disk stats: %w", err)
	}
	metrics.Disk = DiskInfo{
		UsedPercent: diskStat.UsedPercent,
		Used:        diskStat.Used,
		Total:       diskStat.Total,
	}
	
	// Network info
	netStats, err := net.IOCounters(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get network stats: %w", err)
	}
	
	if len(netStats) > 0 {
		metrics.Network = NetworkInfo{
			BytesSent: netStats[0].BytesSent,
			BytesRecv: netStats[0].BytesRecv,
			Count:     len(netStats),
		}
	}
	
	// Host info
	hostStat, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}
	metrics.System = SystemInfo{
		OS:       hostStat.OS,
		Platform: hostStat.Platform,
		Arch:     runtime.GOARCH,
		Cores:    runtime.NumCPU(),
		Uptime:   time.Duration(hostStat.Uptime) * time.Second,
	}
	
	metrics.LastUpdate = time.Now()
	
	return &metrics, nil
}

func printMetrics(metrics *SystemMetrics) {
	fmt.Printf("=== System Metrics ===\n")
	fmt.Printf("Last Updated: %s\n\n", metrics.LastUpdate.Format("2006-01-02 15:04:05"))
	
	// CPU
	fmt.Printf("CPU Usage: %.2f%%\n", metrics.CPUUsage)
	
	// Memory
	fmt.Printf("RAM Usage: %.2f%% (%s / %s)\n",
		metrics.Memory.UsedPercent,
		formatBytes(metrics.Memory.Used),
		formatBytes(metrics.Memory.Total))
	
	// Disk
	fmt.Printf("Disk Usage: %.2f%% (%s / %s)\n",
		metrics.Disk.UsedPercent,
		formatBytes(metrics.Disk.Used),
		formatBytes(metrics.Disk.Total))
	
	// Network
	fmt.Printf("Network: %d interfaces, %s sent, %s received\n",
		metrics.Network.Count,
		formatBytes(metrics.Network.BytesSent),
		formatBytes(metrics.Network.BytesRecv))
	
	// System Info
	fmt.Printf("OS: %s (%s)\n", metrics.System.OS, metrics.System.Platform)
	fmt.Printf("Architecture: %s\n", metrics.System.Arch)
	fmt.Printf("CPU Cores: %d\n", metrics.System.Cores)
	fmt.Printf("Uptime: %s\n", metrics.System.Uptime)
	
	// Visual indicators
	fmt.Printf("\nCPU: %s\n", getVisualBar(metrics.CPUUsage))
	fmt.Printf("RAM: %s\n", getVisualBar(metrics.Memory.UsedPercent))
	fmt.Printf("DISK: %s\n", getVisualBar(metrics.Disk.UsedPercent))
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func getVisualBar(percent float64) string {
	const barLength = 20
	filled := int(percent / 100 * barLength)
	
	bar := ""
	for i := 0; i < barLength; i++ {
		if i < filled {
			bar += "="
		} else {
			bar += "-"
		}
	}
	
	return fmt.Sprintf("[%s] %.2f%%", bar, percent)
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
