APP      := halo
VERSION  := 1.0.0
BIN_DIR  := bin

# ====== CONFIGURATION ======
# You can set these once here instead of in your shell

# Cloud keys
OPENAI_API_KEY    := key_here
ANTHROPIC_API_KEY := key_here

# Local URLs
OLLAMA_URL   := http://localhost:11434/api/generate
MISTRAL_URL  := http://localhost:11434/api/generate

# ====== COMMON LDFLAGS TEMPLATE ======
define LDFLAGS
-X 'halo/models.DefaultAIProvider=$(1)' \
-X 'halo/models.DefaultAIModel=$(2)' \
-X 'halo/models.DefaultAPIURL=$(3)' \
-X 'halo/models.DefaultAPIKey=$(4)'
endef

# ---------------- CLOUD PROVIDERS ----------------

chatgpt_linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(call LDFLAGS,chatgpt,gpt-4.1,https://api.openai.com/v1/chat/completions,$(OPENAI_API_KEY))" \
		-o $(BIN_DIR)/$(APP)_chatgpt_linux

chatgpt_windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(call LDFLAGS,chatgpt,gpt-4.1,https://api.openai.com/v1/chat/completions,$(OPENAI_API_KEY))" \
		-o $(BIN_DIR)/$(APP)_chatgpt_windows.exe

anthropic_linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(call LDFLAGS,anthropic,claude-3-opus,https://api.anthropic.com/v1/complete,$(ANTHROPIC_API_KEY))" \
		-o $(BIN_DIR)/$(APP)_anthropic_linux

anthropic_windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(call LDFLAGS,anthropic,claude-3-opus,https://api.anthropic.com/v1/complete,$(ANTHROPIC_API_KEY))" \
		-o $(BIN_DIR)/$(APP)_anthropic_windows.exe

# ---------------- LOCAL PROVIDERS ----------------

ollama_linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(call LDFLAGS,ollama,llama3,$(OLLAMA_URL),)" \
		-o $(BIN_DIR)/$(APP)_ollama_linux

ollama_windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(call LDFLAGS,ollama,llama3,$(OLLAMA_URL),)" \
		-o $(BIN_DIR)/$(APP)_ollama_windows.exe

mistral_linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(call LDFLAGS,mistral,mistral,$(MISTRAL_URL),)" \
		-o $(BIN_DIR)/$(APP)_mistral_linux

mistral_windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(call LDFLAGS,mistral,mistral,$(MISTRAL_URL),)" \
		-o $(BIN_DIR)/$(APP)_mistral_windows.exe

# ---------------- GROUP TARGETS ----------------

linux: chatgpt_linux anthropic_linux ollama_linux mistral_linux
windows: chatgpt_windows anthropic_windows ollama_windows mistral_windows
all: linux windows
