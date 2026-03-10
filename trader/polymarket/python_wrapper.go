// /Users/tom/Desktop/nofi/trader/polymarket/python_wrapper.go
package polymarket

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

type PythonWrapper struct {
	cmd        *exec.Cmd
	pythonPath string
	scriptPath string
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	scanner    *bufio.Scanner
	mutex      sync.Mutex
	isRunning  bool
}

func NewPythonWrapper(pythonPath, scriptPath string) (*PythonWrapper, error) {
	wrapper := &PythonWrapper{
		pythonPath: pythonPath,
		scriptPath: scriptPath,
	}

	if err := wrapper.start(); err != nil {
		return nil, err
	}

	// 启动健康检查
	go wrapper.healthCheck()

	return wrapper, nil
}

func (p *PythonWrapper) start() error {
	// 启动Python进程
	cmd := exec.Command(p.pythonPath, p.scriptPath)
	
	// Pipe stderr to main process stderr for debugging
	cmd.Stderr = exec.Command("cat").Stderr 

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Python process: %w", err)
	}

	p.cmd = cmd
	p.stdin = stdin
	p.stdout = stdout
	p.scanner = bufio.NewScanner(stdout)
	p.isRunning = true
	
	return nil
}

func (p *PythonWrapper) Call(command map[string]interface{}) (map[string]interface{}, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isRunning {
		return nil, fmt.Errorf("Python process is not running")
	}

	// 发送命令
	cmdJSON, err := json.Marshal(command)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}

	if _, err := p.stdin.Write(append(cmdJSON, '\n')); err != nil {
		p.isRunning = false
		return nil, fmt.Errorf("failed to write to stdin: %w", err)
	}

	// Use a channel to handle timeout
	type responseResult struct {
		data map[string]interface{}
		err  error
	}
	resultChan := make(chan responseResult, 1)

	// Capture scanner to avoid race condition if restart happens
	scanner := p.scanner

	go func() {
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				resultChan <- responseResult{nil, fmt.Errorf("failed to read from stdout: %w", err)}
			} else {
				resultChan <- responseResult{nil, fmt.Errorf("stdout closed")}
			}
			return
		}

		responseJSON := scanner.Text()
		var response map[string]interface{}
		if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
			resultChan <- responseResult{nil, fmt.Errorf("failed to unmarshal response: %w", err)}
			return
		}
		resultChan <- responseResult{response, nil}
	}()

	select {
	case res := <-resultChan:
		if res.err != nil {
			p.isRunning = false
			return nil, res.err
		}
		return res.data, nil
	case <-time.After(10 * time.Second): // 10s timeout
		p.isRunning = false
		// Kill process to unblock the goroutine (scanner.Scan)
		if p.cmd != nil && p.cmd.Process != nil {
			_ = p.cmd.Process.Kill()
		}
		return nil, fmt.Errorf("timeout waiting for Python response")
	}
}

func (p *PythonWrapper) healthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Check if running without lock first to avoid holding it too long
		p.mutex.Lock()
		if !p.isRunning {
			p.mutex.Unlock()
			continue // Or return if we want to stop checking
		}
		p.mutex.Unlock()

		// Send ping command (Call handles its own locking)
		_, err := p.Call(map[string]interface{}{
			"command":   "ping",
			"timestamp": time.Now().Unix(),
		})

		if err != nil {
			// Restart logic needs lock to update state safely
			p.mutex.Lock()
			p.isRunning = false
			p.mutex.Unlock()
			
			// Attempt restart
			if restartErr := p.restart(); restartErr != nil {
				fmt.Printf("Failed to restart Python process: %v\n", restartErr)
			}
		}
	}
}

func (p *PythonWrapper) restart() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
	}

	return p.start()
}

func (p *PythonWrapper) Close() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.isRunning {
		p.cmd.Process.Kill()
		p.isRunning = false
	}
}
