APP      := halo
VERSION  := 1.0.0
BIN_DIR  := bin

# ====== CONFIGURATION ======
OPENAI_API_KEY    := key_here
ANTHROPIC_API_KEY := key_here
OLLAMA_URL        := http://192.168.1.145:11434/api/generate

# Local models (from `ollama list`)
OLLAMA_MODELS := llama3 mistral neural-chat

# ====== LDFLAGS TEMPLATE ======
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

# ---------------- LOCAL OLLAMA MODELS ----------------
# Function to sanitize names (replace ':' and '-' with '_')
sanitize = $(subst -,_,$(subst :,_,$(1)))

# Function to build rules for each model
define BUILD_OLLAMA_RULES
$(call sanitize,$(1))_linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(call LDFLAGS,ollama,$(1),$(OLLAMA_URL),)" \
		-o $(BIN_DIR)/$(APP)_$(call sanitize,$(1))_linux

$(call sanitize,$(1))_windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(call LDFLAGS,ollama,$(1),$(OLLAMA_URL),)" \
		-o $(BIN_DIR)/$(APP)_$(call sanitize,$(1))_windows.exe
endef

# Expand for each model in OLLAMA_MODELS
$(foreach m,$(OLLAMA_MODELS),$(eval $(call BUILD_OLLAMA_RULES,$(m))))

# ---------------- GROUP TARGETS ----------------
linux: chatgpt_linux anthropic_linux $(foreach m,$(OLLAMA_MODELS),$(call sanitize,$(m))_linux)
windows: chatgpt_windows anthropic_windows $(foreach m,$(OLLAMA_MODELS),$(call sanitize,$(m))_windows)
all: linux windows

# ---------------- UTIL ----------------
list:
	@echo "Available targets:"
	@echo "  chatgpt_linux anthropic_linux"
	@echo "  chatgpt_windows anthropic_windows"
	@$(foreach m,$(OLLAMA_MODELS),echo "  $(call sanitize,$(m))_linux $(call sanitize,$(m))_windows";)

# ---------------- CLEAN ----------------
clean:
	rm -f $(BIN_DIR)/$(APP)_*
