package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

func findProcessByName(name string) []map[string]interface{} {
	listProcess := []map[string]interface{}{}
	seen := make(map[string]bool)

	processes, _ := process.Processes()
	for _, proc := range processes {
		procName, _ := proc.Name()
		if strings.Contains(procName, name) {
			pid := proc.Pid
			connections, _ := proc.Connections()
			for _, conn := range connections {
				if conn.Status == "LISTEN" {
					key := fmt.Sprintf("%d-%d", pid, conn.Laddr.Port)
					if !seen[key] {
						seen[key] = true
						listProcess = append(listProcess, map[string]interface{}{
							"pid":  pid,
							"name": procName,
							"port": conn.Laddr.Port,
						})
					}
				}
			}
		}
	}

	return listProcess
}

func killDeadServers(pid int32) {
	fmt.Printf("正在结束本机进程 pid %d\n", pid)
	cmd := exec.Command("taskkill", "/F", "/PID", strconv.Itoa(int(pid)))
	cmd.Run()
	for pidExists(pid) {
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Printf("pid %d kill success\n", pid)
}

func pidExists(pid int32) bool {
	_, err := process.NewProcess(pid)
	return err == nil
}

func main() {
	fmt.Println("正在查找ios_webkit_debug_proxy进程是否存在")
	pList := findProcessByName("ios_webkit_debug_proxy")
	fmt.Println(pList)
	if len(pList) > 0 {
		fmt.Println("发现ios_webkit_debug_proxy进程，准备结束")
		for _, p := range pList {
			killDeadServers(p["pid"].(int32))
		}
	} else {
		fmt.Println("未发现ios_webkit_debug_proxy进程")
	}

	var port string
	fmt.Print("【选填】请输入监听端口号（默认9000）：")
	fmt.Scanln(&port)
	if port == "" {
		port = "9000"
	}

	currentPath, _ := os.Executable()
	currentDir := filepath.Dir(currentPath)
	cmd := exec.Command(filepath.Join(currentDir, "remotedebug_ios_webkit_adapter.exe"), "--port="+port)

	// 设置标准输出和标准错误
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("开始运行ios_webkit_debug_proxy.exe，监听端口号：%s\n", port)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("进程运行出错：%v\n", err)
	}
	fmt.Printf("进程已结束，退出代码为 %d\n", cmd.ProcessState.ExitCode())
}
