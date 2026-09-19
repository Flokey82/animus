package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// AgentStatusProvider provides introspection data to web clients.
type AgentStatusProvider interface {
	AgentHandler
	BuildSystemPrompt() string
}

// WebServer hosts a modern web chat UI and REST/SSE API for any Animus agent.
type WebServer struct {
	port    int
	handler AgentStatusProvider
	hub     *Hub
	server  *http.Server
}

// NewWebServer initializes the web chat server.
func NewWebServer(port int, handler AgentStatusProvider, hub *Hub) *WebServer {
	if port <= 0 {
		port = 8080
	}
	ws := &WebServer{
		port:    port,
		handler: handler,
		hub:     hub,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", ws.handleIndex)
	mux.HandleFunc("/api/chat", ws.handleChat)
	mux.HandleFunc("/api/events", ws.handleEvents)
	mux.HandleFunc("/api/status", ws.handleStatus)

	ws.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return ws
}

// Start runs the HTTP server in a background goroutine.
func (ws *WebServer) Start() error {
	go func() {
		_ = ws.server.ListenAndServe()
	}()
	return nil
}

// Stop shuts down the HTTP server gracefully.
func (ws *WebServer) Stop(ctx context.Context) error {
	return ws.server.Shutdown(ctx)
}

func (ws *WebServer) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Message string `json:"message"`
		Sender  string `json:"sender"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	sender := req.Sender
	if sender == "" {
		sender = "User"
	}

	reply, err := ws.handler.Chat(r.Context(), strings.TrimSpace(req.Message), sender)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"reply": reply})
}

func (ws *WebServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "online",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (ws *WebServer) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	sub := ws.hub.Subscribe()
	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return
		case msg := <-sub:
			data, _ := json.Marshal(msg)
			fmt.Fprintf(w, "data: %s\n\n", string(data))
			flusher.Flush()
		}
	}
}

func (ws *WebServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>🌟 Animus Agent Terminal</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg: #0b0f19;
      --card-bg: rgba(18, 24, 38, 0.75);
      --border: rgba(255, 255, 255, 0.08);
      --accent: #6366f1;
      --accent-glow: rgba(99, 102, 241, 0.35);
      --text: #f3f4f6;
      --text-dim: #9ca3af;
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
      background: var(--bg);
      color: var(--text);
      height: 100vh;
      display: flex;
      flex-direction: column;
    }
    header {
      padding: 1rem 2rem;
      background: var(--card-bg);
      backdrop-filter: blur(12px);
      border-bottom: 1px solid var(--border);
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
    .brand { font-weight: 700; font-size: 1.15rem; color: #fff; display: flex; align-items: center; gap: 0.5rem; }
    .status-badge {
      display: inline-flex; align-items: center; gap: 0.4rem;
      background: rgba(16, 185, 129, 0.15); color: #10b981;
      padding: 0.25rem 0.75rem; border-radius: 9999px; font-size: 0.8rem; font-weight: 600;
    }
    .status-dot { width: 8px; height: 8px; border-radius: 50%; background: #10b981; box-shadow: 0 0 8px #10b981; }
    main {
      flex: 1;
      display: flex;
      flex-direction: column;
      max-width: 900px;
      width: 100%;
      margin: 0 auto;
      padding: 1.5rem;
      gap: 1rem;
      overflow: hidden;
    }
    #chat-log {
      flex: 1;
      overflow-y: auto;
      display: flex;
      flex-direction: column;
      gap: 1rem;
      padding-right: 0.5rem;
    }
    .bubble {
      max-width: 80%;
      padding: 0.9rem 1.25rem;
      border-radius: 16px;
      line-height: 1.5;
      font-size: 0.95rem;
      word-break: break-word;
      animation: fadeIn 0.2s ease;
    }
    .bubble.user {
      align-self: flex-end;
      background: var(--accent);
      color: #fff;
      border-bottom-right-radius: 4px;
      box-shadow: 0 4px 14px var(--accent-glow);
    }
    .bubble.agent {
      align-self: flex-start;
      background: var(--card-bg);
      border: 1px solid var(--border);
      color: var(--text);
      border-bottom-left-radius: 4px;
      backdrop-filter: blur(8px);
    }
    .bubble.system {
      align-self: center;
      background: rgba(255, 255, 255, 0.04);
      color: var(--text-dim);
      font-size: 0.8rem;
      padding: 0.4rem 0.8rem;
    }
    .input-area {
      display: flex;
      gap: 0.75rem;
      background: var(--card-bg);
      border: 1px solid var(--border);
      border-radius: 14px;
      padding: 0.5rem 0.75rem;
      backdrop-filter: blur(12px);
    }
    input[type="text"] {
      flex: 1;
      background: transparent;
      border: none;
      outline: none;
      color: #fff;
      font-size: 1rem;
      padding: 0.5rem;
    }
    input::placeholder { color: #6b7280; }
    button {
      background: var(--accent);
      color: #fff;
      border: none;
      padding: 0.6rem 1.25rem;
      border-radius: 10px;
      font-weight: 600;
      cursor: pointer;
      transition: all 0.2s ease;
    }
    button:hover { filter: brightness(1.1); transform: translateY(-1px); }
    @keyframes fadeIn { from { opacity: 0; transform: translateY(6px); } to { opacity: 1; transform: translateY(0); } }
  </style>
</head>
<body>
  <header>
    <div class="brand">✨ Animus Agent</div>
    <div class="status-badge"><div class="status-dot"></div> Connected</div>
  </header>
  <main>
    <div id="chat-log">
      <div class="bubble system">System initialized. Connected to Animus Autonomous Gateway.</div>
      <div class="bubble agent">Hello! I am online and ready to explore, research, and converse. How can I assist you today?</div>
    </div>
    <form class="input-area" id="chat-form">
      <input type="text" id="msg-input" placeholder="Type a message or ask a question..." autocomplete="off" autofocus />
      <button type="submit">Send</button>
    </form>
  </main>
  <script>
    const log = document.getElementById('chat-log');
    const form = document.getElementById('chat-form');
    const input = document.getElementById('msg-input');

    function appendBubble(text, role) {
      const b = document.createElement('div');
      b.className = 'bubble ' + role;
      b.textContent = text;
      log.appendChild(b);
      log.scrollTop = log.scrollHeight;
    }

    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      const val = input.value.trim();
      if (!val) return;

      appendBubble(val, 'user');
      input.value = '';

      const typing = document.createElement('div');
      typing.className = 'bubble agent';
      typing.textContent = 'Thinking...';
      log.appendChild(typing);
      log.scrollTop = log.scrollHeight;

      try {
        const res = await fetch('/api/chat', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ message: val, sender: 'WebUser' })
        });
        const data = await res.json();
        typing.textContent = data.reply || (data.error ? 'Error: ' + data.error : 'No response');
      } catch (err) {
        typing.textContent = 'Failed to communicate with agent: ' + err.message;
      }
      log.scrollTop = log.scrollHeight;
    });

    // Listen to real-time events via Server-Sent Events (SSE)
    const evtSource = new EventSource('/api/events');
    evtSource.onmessage = (e) => {
      try {
        const msg = JSON.parse(e.data);
        if (msg.channel === 'agent_chat') {
          // Can sync messages from external channels (e.g. Discord, Telegram)
        }
      } catch (_) {}
    };
  </script>
</body>
</html>`
