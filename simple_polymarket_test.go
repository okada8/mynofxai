package main

import (
	"fmt"
	"os"
	"os/exec"
	"encoding/json"
)

func main() {
	fmt.Println("🧪 测试Polymarket Python桥接...")
	
	// 直接测试Python桥接
	scriptPath := "trader/polymarket/polymarket_bridge_fixed.py"
	
	// 检查脚本是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		fmt.Printf("❌ Python脚本不存在: %s\n", scriptPath)
		return
	}
	
	fmt.Printf("✅ Python脚本存在: %s\n", scriptPath)
	
	// 使用exec直接测试，不使用包装器
	cmd := exec.Command("python3", scriptPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("❌ 创建stdin管道失败: %v\n", err)
		return
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("❌ 创建stdout管道失败: %v\n", err)
		return
	}
	
	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ 启动Python进程失败: %v\n", err)
		return
	}
	
	// 发送ping命令
	pingCmd := map[string]interface{}{
		"command": "ping",
		"timestamp": 1234567890,
	}
	
	cmdJSON, err := json.Marshal(pingCmd)
	if err != nil {
		fmt.Printf("❌ JSON编码失败: %v\n", err)
		return
	}
	
	if _, err := stdin.Write(append(cmdJSON, '\n')); err != nil {
		fmt.Printf("❌ 写入stdin失败: %v\n", err)
		return
	}
	stdin.Close()
	
	// 读取响应
	buf := make([]byte, 1024)
	n, err := stdout.Read(buf)
	if err != nil {
		fmt.Printf("❌ 读取stdout失败: %v\n", err)
		return
	}
	
	response := string(buf[:n])
	fmt.Printf("✅ Python响应: %s\n", response)
	
	// 解析响应
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(response), &resp); err != nil {
		fmt.Printf("❌ JSON解码失败: %v\n", err)
		return
	}
	
	if status, ok := resp["status"].(string); ok && status == "pong" {
		fmt.Println("✅ Python桥接测试通过!")
	} else {
		fmt.Printf("❌ 响应状态异常: %v\n", resp)
	}
	
	cmd.Wait()
}
