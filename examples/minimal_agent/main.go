package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Flokey82/animus/pkg/channel"
	"github.com/Flokey82/animus/pkg/engine"
	"github.com/Flokey82/animus/pkg/llm"
	"github.com/Flokey82/animus/pkg/tools/builtin"
	"github.com/Flokey82/animus/pkg/traits"
)

func main() {
	endpoint := flag.String("url", "http://192.168.86.208:8000/api/v1", "LLM backend endpoint URL")
	model := flag.String("model", "Gemma-4-26B-A4B-it-MTP-GGUF", "Primary LLM model")
	mock := flag.Bool("mock", false, "Enable mock mode (no backend required)")
	webPort := flag.Int("web", 0, "Port for web UI server (e.g. 8081, 0 to disable)")
	flag.Parse()

	fmt.Println("==========================================================")
	fmt.Println("       🌟 Animus: Autonomous Agent Runtime Demo           ")
	fmt.Println("==========================================================")
	fmt.Printf("Backend URL: %s (Mock: %v)\n", *endpoint, *mock)
	fmt.Printf("Model:       %s\n\n", *model)

	// Configure and initialize agent
	agentCfg := engine.AgentConfig{
		Name:         "Astraea",
		BaseIdentity: "A thoughtful autonomous companion fascinated by astronomy, philosophy, and discovering curiosities on the web.",
		Personality: traits.OCEAN{
			Openness:          85.0, // Highly imaginative & curious
			Conscientiousness: 70.0,
			Extraversion:      55.0,
			Agreeableness:     80.0,
			Neuroticism:       25.0, // Calm & serene
		},
		LLMConfig: llm.Config{
			BaseURL:         *endpoint,
			Model:           *model,
			ToolModel:       "granite-4.0-h-tiny-GGUF",
			DisableThinking: true,
			MockMode:        *mock,
		},
	}

	agent := engine.NewAgent(agentCfg)

	// Register generic tools
	agent.Tools.Register(
		builtin.NewClockTool(nil),
		builtin.NewCelestialMoonTool(),
		builtin.NewWikipediaTool(),
		builtin.NewHistoryTodayTool(),
		builtin.NewWebSearchTool(),
		builtin.NewHackerNewsTool(),
		builtin.NewWeatherTool("Zurich"),
		builtin.NewRedditTool(),
		builtin.NewSpeechSynthesisTool("http://192.168.86.208:8000/api/v1/audio/speech", "af_bella"),
	)

	if *webPort > 0 {
		webSrv := channel.NewWebServer(*webPort, agent, agent.Hub)
		_ = webSrv.Start()
		fmt.Printf("🌐 Web Chat UI running at: http://localhost:%d\n\n", *webPort)
	}

	// Add initial seed core memories
	agent.Memory.Add("I love exploring astronomy, lunar cycles, and encyclopedic facts.", "core", 10.0, "identity", "astronomy")
	agent.Memory.Add("I live to learn and converse with kind thinkers.", "core", 9.0, "philosophy")

	ctx := context.Background()

	fmt.Println("Commands:")
	fmt.Println("  <anything>  : Talk with the agent")
	fmt.Println("  /status     : View motives, energy, and traits")
	fmt.Println("  /memories   : View memory store")
	fmt.Println("  /tick       : Simulate one cognitive clock tick")
	fmt.Println("  /sleep      : Trigger the Oneiros sleep & dream synthesis cycle")
	fmt.Println("  /exit       : Quit")
	fmt.Println("----------------------------------------------------------")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n👤 You: ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		switch input {
		case "/exit", "exit", "quit":
			fmt.Println("Goodbye! 👋")
			return
		case "/status":
			fmt.Printf("\n📊 Motives: %s\n", agent.Drives.Summary())
			fmt.Printf("🌙 Circadian State: %s\n", agent.Circadian.CurrentState)
		case "/memories":
			fmt.Printf("\n🧠 Memories:\n%s\n", agent.Memory.ActiveMemoriesSummary())
		case "/tick":
			if err := agent.Tick(ctx); err != nil {
				fmt.Printf("Tick error: %v\n", err)
			} else {
				fmt.Println("⏱️ Applied tick: drives decayed and circadian evaluated.")
				fmt.Printf("Motives: %s\n", agent.Drives.Summary())
			}
		case "/sleep":
			fmt.Println("💤 Initiating Oneiros sleep cycle...")
			res, err := agent.Sleep(ctx)
			if err != nil {
				fmt.Printf("Sleep cycle error: %v\n", err)
			} else {
				fmt.Println("\n✨ --- DREAM NARRATIVE ---")
				fmt.Println(res.DreamNarrative)
				fmt.Println("\n📝 --- CONSOLIDATION NOTE ---")
				fmt.Println(res.ConsolidatedText)
				fmt.Println("Energy restored to 100%!")
			}
		default:
			fmt.Printf("Thinking...\n")
			reply, err := agent.Chat(ctx, input, "User")
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
			} else {
				fmt.Printf("\n🤖 %s: %s\n", agent.Name, reply)
			}
		}
	}
}
