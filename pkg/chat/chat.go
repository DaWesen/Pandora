package chat

import (
	"bufio"
	"fmt"
	"os"

	"github.com/DaWesen/Pandora/pkg/core"
)

// ChatLoop 启动一个连续对话循环
// agent: 要使用的 Agent 实例
// prompt: 用户输入提示，默认为 "User: "
// exitCommand: 退出命令，默认为 "exit"
func ChatLoop(agent *core.AgentImpl, prompt, exitCommand string) {
	if prompt == "" {
		prompt = "User: "
	}
	if exitCommand == "" {
		exitCommand = "exit"
	}

	scanner := bufio.NewScanner(os.Stdin)
	
	fmt.Println("Pandora Agent started. Type", exitCommand, "to quit.")
	fmt.Println("---------------------------------------------)")
	
	for {
		// 读取用户输入
		fmt.Print(prompt)
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		
		// 退出条件
		if input == exitCommand {
			fmt.Println("Goodbye!")
			break
		}
		
		// 执行对话
		response, err := agent.Run(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		
		// 显示响应
		fmt.Printf("Agent: %s\n", response.Content)
		fmt.Println("---------------------------------------------)")
	}
}

// StreamingChatLoop 启动一个流式连续对话循环
// agent: 要使用的 Agent 实例
// prompt: 用户输入提示，默认为 "User: "
// exitCommand: 退出命令，默认为 "exit"
func StreamingChatLoop(agent *core.AgentImpl, prompt, exitCommand string) {
	if prompt == "" {
		prompt = "User: "
	}
	if exitCommand == "" {
		exitCommand = "exit"
	}

	scanner := bufio.NewScanner(os.Stdin)
	
	fmt.Println("Pandora Agent started. Type", exitCommand, "to quit.")
	fmt.Println("---------------------------------------------)")
	
	for {
		// 读取用户输入
		fmt.Print(prompt)
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		
		// 退出条件
		if input == exitCommand {
			fmt.Println("Goodbye!")
			break
		}
		
		// 执行流式对话
		fmt.Print("Agent: ")
		msgChan, errChan := agent.StreamRun(input)
		
		// 处理流式响应
		for {
			select {
			case msg, ok := <-msgChan:
				if !ok {
					fmt.Println()
					fmt.Println("---------------------------------------------)")
					goto nextTurn
				}
				fmt.Printf("%s", msg.Content)
			case err, ok := <-errChan:
				if !ok {
					goto nextTurn
				}
				fmt.Printf("Error: %v\n", err)
				goto nextTurn
			}
		}
		
	nextTurn:
		continue
	}
}
