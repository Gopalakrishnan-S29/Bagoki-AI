import { useEffect, useRef, useState } from "react";
import "./App.css";

type Message = {
  id: number;
  role: "user" | "assistant";
  content: string;
  metadata?: {
    provider?: string;
    model?: string;
    taskType?: string;
    complexity?: string;
    latencyMs?: number;
  };
};

type ChatResponse = {
  reply: string;
  provider: string;
  model: string;
  task_type: string;
  complexity: string;
  risk_tier: string;
  reason: string;
  tokens_used: number;
  latency_ms: number;
};

const API_BASE_URL = "http://localhost:8080";

function App() {
  const [prompt, setPrompt] = useState("");
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(false);
  const [backendOnline, setBackendOnline] = useState(false);
  const [activePage, setActivePage] = useState("Dashboard");
  const [showDetails, setShowDetails] = useState(false);

  const textareaRef = useRef<HTMLTextAreaElement>(null);

  /*
   * ------------------------------------------------------------
   * Backend health check
   * ------------------------------------------------------------
   */

  const checkBackend = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/health`);

      if (response.ok) {
        setBackendOnline(true);
      } else {
        setBackendOnline(false);
      }
    } catch {
      setBackendOnline(false);
    }
  };

  useEffect(() => {
    checkBackend();

    const interval = setInterval(checkBackend, 10000);

    return () => clearInterval(interval);
  }, []);

  /*
   * ------------------------------------------------------------
   * Auto resize textarea
   * ------------------------------------------------------------
   */

  const resizeTextarea = () => {
    const textarea = textareaRef.current;

    if (!textarea) {
      return;
    }

    textarea.style.height = "auto";
    textarea.style.height = `${Math.min(textarea.scrollHeight, 220)}px`;
  };

  useEffect(() => {
    resizeTextarea();
  }, [prompt]);

  /*
   * ------------------------------------------------------------
   * Send prompt to Go backend
   * ------------------------------------------------------------
   */

  const sendPrompt = async () => {
    const trimmedPrompt = prompt.trim();

    if (!trimmedPrompt || loading) {
      return;
    }

    const userMessage: Message = {
      id: Date.now(),
      role: "user",
      content: trimmedPrompt,
    };

    setMessages((previous) => [...previous, userMessage]);
    setPrompt("");
    setLoading(true);

    try {
      const response = await fetch(
        `${API_BASE_URL}/api/router/chat`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            prompt: trimmedPrompt,
          }),
        }
      );

      const responseText = await response.text();

      let data: ChatResponse | null = null;

      try {
        data = JSON.parse(responseText);
      } catch {
        data = null;
      }

      if (!response.ok) {
        throw new Error(
          data && "reply" in data
            ? data.reply
            : responseText || "Backend request failed."
        );
      }

      if (!data) {
        throw new Error("Invalid response from backend.");
      }

      const assistantMessage: Message = {
        id: Date.now() + 1,
        role: "assistant",
        content: data.reply,
        metadata: {
          provider: data.provider,
          model: data.model,
          taskType: data.task_type,
          complexity: data.complexity,
          latencyMs: data.latency_ms,
        },
      };

      setMessages((previous) => [...previous, assistantMessage]);
      setBackendOnline(true);
    } catch (error) {
      const errorMessage =
        error instanceof Error
          ? error.message
          : "Something went wrong while contacting the AI Router.";

      const assistantMessage: Message = {
        id: Date.now() + 1,
        role: "assistant",
        content: `Unable to complete the request.\n\n${errorMessage}`,
      };

      setMessages((previous) => [...previous, assistantMessage]);

      checkBackend();
    } finally {
      setLoading(false);
    }
  };

  /*
   * ------------------------------------------------------------
   * Keyboard handling
   * ------------------------------------------------------------
   */

  const handleKeyDown = (
    event: React.KeyboardEvent<HTMLTextAreaElement>
  ) => {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      sendPrompt();
    }
  };

  /*
   * ------------------------------------------------------------
   * New chat
   * ------------------------------------------------------------
   */

  const newChat = () => {
    setMessages([]);
    setPrompt("");
    setActivePage("Dashboard");
    setShowDetails(false);

    setTimeout(() => {
      textareaRef.current?.focus();
    }, 50);
  };

  /*
   * ------------------------------------------------------------
   * Example prompts
   * ------------------------------------------------------------
   */

  const examplePrompts = [
    "Explain what a firewall is in simple terms.",
    "Compare TCP and UDP.",
    "What is SQL injection?",
  ];

  /*
   * ------------------------------------------------------------
   * Render
   * ------------------------------------------------------------
   */

  return (
    <div className="app-shell">
      {/* ======================================================
          SIDEBAR
          ====================================================== */}

      <aside className="sidebar">
        <div className="sidebar-top">
          <div className="brand">
            <div className="brand-mark">AI</div>

            <div className="brand-text">
              <div className="brand-name">AI Router</div>
              <div className="brand-subtitle">
                Intelligent Model Routing
              </div>
            </div>
          </div>

          <button
            className="new-chat-button"
            onClick={newChat}
            type="button"
          >
            <span className="new-chat-plus">+</span>
            <span>New chat</span>
          </button>

          <nav className="navigation">
            {[
              {
                name: "Dashboard",
                icon: "⌂",
              },
              {
                name: "Router",
                icon: "ϟ",
              },
              {
                name: "Models",
                icon: "◉",
              },
              {
                name: "Analytics",
                icon: "▣",
              },
            ].map((item) => (
              <button
                key={item.name}
                type="button"
                className={`nav-item ${
                  activePage === item.name ? "active" : ""
                }`}
                onClick={() => setActivePage(item.name)}
              >
                <span className="nav-icon">{item.icon}</span>
                <span>{item.name}</span>
              </button>
            ))}
          </nav>
        </div>

        <div className="sidebar-bottom">
          <div className="system-card">
            <div className="system-card-header">
              <span
                className={`status-dot ${
                  backendOnline ? "online" : "offline"
                }`}
              />

              <span className="system-label">
                BACKEND
              </span>
            </div>

            <div
              className={`system-status ${
                backendOnline ? "online-text" : "offline-text"
              }`}
            >
              {backendOnline ? "Online" : "Offline"}
            </div>

            <div className="system-api">
              Go API :8080
            </div>
          </div>

          <button
            type="button"
            className="connection-button"
            onClick={checkBackend}
          >
            Check connection
          </button>

          <div className="sidebar-footer">
            <span className="footer-avatar">G</span>

            <div className="footer-user">
              <span className="footer-name">
                AI Router
              </span>
              <span className="footer-role">
                Local workspace
              </span>
            </div>

            <span className="footer-menu">•••</span>
          </div>
        </div>
      </aside>

      {/* ======================================================
          MAIN AREA
          ====================================================== */}

      <main className="main-content">
        <header className="top-bar">
          <div className="mobile-brand">
            <div className="brand-mark small">AI</div>
            <span>AI Router</span>
          </div>

          <div className="backend-indicator">
            <span
              className={`status-dot ${
                backendOnline ? "online" : "offline"
              }`}
            />

            <span>
              {backendOnline
                ? "Backend connected"
                : "Backend offline"}
            </span>
          </div>
        </header>

        <div className="chat-container">
          {/* ==================================================
              EMPTY STATE
              ================================================== */}

          {messages.length === 0 && (
            <section className="welcome-section">
              <div className="welcome-mark">
                ✦
              </div>

              <h1>
                Good afternoon
              </h1>

              <p className="welcome-description">
                How can I help you today?
              </p>

              <div className="welcome-context">
                AI Router intelligently selects the
                appropriate model for your request.
              </div>
            </section>
          )}

          {/* ==================================================
              CHAT MESSAGES
              ================================================== */}

          {messages.length > 0 && (
            <section className="messages-container">
              {messages.map((message) => (
                <div
                  className={`message-row ${message.role}`}
                  key={message.id}
                >
                  {message.role === "user" ? (
                    <div className="user-message">
                      {message.content}
                    </div>
                  ) : (
                    <div className="assistant-message">
                      <div className="assistant-header">
                        <div className="assistant-avatar">
                          AI
                        </div>

                        <span className="assistant-name">
                          AI Router
                        </span>
                      </div>

                      <div className="assistant-content">
                        {message.content}
                      </div>

                      {message.metadata && (
                        <div className="response-meta">
                          <div className="meta-summary">
                            <span>
                              {message.metadata.provider}
                            </span>

                            <span>•</span>

                            <span>
                              {message.metadata.model}
                            </span>

                            <span>•</span>

                            <span>
                              {message.metadata.latencyMs} ms
                            </span>

                            <button
                              type="button"
                              onClick={() =>
                                setShowDetails(
                                  !showDetails
                                )
                              }
                            >
                              {showDetails
                                ? "Hide details"
                                : "Routing details"}
                            </button>
                          </div>

                          {showDetails && (
                            <div className="routing-details">
                              <div>
                                <span>Provider</span>
                                <strong>
                                  {message.metadata.provider}
                                </strong>
                              </div>

                              <div>
                                <span>Model</span>
                                <strong>
                                  {message.metadata.model}
                                </strong>
                              </div>

                              <div>
                                <span>Task</span>
                                <strong>
                                  {message.metadata.taskType}
                                </strong>
                              </div>

                              <div>
                                <span>Complexity</span>
                                <strong>
                                  {message.metadata.complexity}
                                </strong>
                              </div>

                              <div>
                                <span>Latency</span>
                                <strong>
                                  {message.metadata.latencyMs} ms
                                </strong>
                              </div>
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              ))}

              {loading && (
                <div className="message-row assistant">
                  <div className="assistant-message">
                    <div className="assistant-header">
                      <div className="assistant-avatar">
                        AI
                      </div>

                      <span className="assistant-name">
                        AI Router
                      </span>
                    </div>

                    <div className="loading-response">
                      <span />
                      <span />
                      <span />
                    </div>
                  </div>
                </div>
              )}
            </section>
          )}

          {/* ==================================================
              EXAMPLES
              ================================================== */}

          {messages.length === 0 && (
            <div className="examples">
              {examplePrompts.map((example) => (
                <button
                  key={example}
                  type="button"
                  className="example-card"
                  onClick={() => {
                    setPrompt(example);
                    textareaRef.current?.focus();
                  }}
                >
                  <span className="example-icon">
                    ✦
                  </span>

                  <span>{example}</span>
                </button>
              ))}
            </div>
          )}

          {/* ==================================================
              COMPOSER
              ================================================== */}

          <div className="composer-wrapper">
            <div className="composer">
              <textarea
                ref={textareaRef}
                value={prompt}
                onChange={(event) =>
                  setPrompt(event.target.value)
                }
                onKeyDown={handleKeyDown}
                placeholder="How can I help you today?"
                rows={1}
                disabled={loading}
              />

              <div className="composer-bottom">
                <div className="composer-left">
                  <button
                    type="button"
                    className="composer-icon-button"
                    title="New chat"
                    onClick={newChat}
                  >
                    +
                  </button>

                  <div className="mode-selector">
                    <button
                      type="button"
                      className="mode active"
                    >
                      Chat
                    </button>

                    <button
                      type="button"
                      className="mode"
                    >
                      Router
                    </button>
                  </div>
                </div>

                <div className="composer-right">
                  <span className="model-name">
                    llama3.2:latest
                  </span>

                  <button
                    type="button"
                    className={`send-button ${
                      prompt.trim() && !loading
                        ? "ready"
                        : ""
                    }`}
                    onClick={sendPrompt}
                    disabled={
                      !prompt.trim() || loading
                    }
                    title="Send"
                  >
                    {loading ? "..." : "↑"}
                  </button>
                </div>
              </div>
            </div>

            <div className="composer-note">
              AI Router may make mistakes. Check
              important information.
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}

export default App;