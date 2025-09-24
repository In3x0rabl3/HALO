# HALO — Hybrid Autonomous Logic Operator

HALO is a Windows-focused research framework that demonstrates conditional execution of encrypted payloads based on AI-assisted telemetry evaluation. The system integrates host environment analysis with OpenAI GPT models to determine whether an embedded payload should be executed. Payloads are encrypted, embedded at build time, and executed inline through Windows API calls.

---

## Design Overview

HALO consists of four major subsystems:

1. **Telemetry**  
   Collects runtime host data such as process information, parent process, system uptime, idle metrics, and device states. Establishes a baseline snapshot and compares subsequent samples to detect deviations or suspicious conditions.

2. **AI Decision Engine**  
   Sends structured telemetry data to an OpenAI GPT model with a system prompt describing decision rules. The AI responds with JSON containing:
   - `allow` (boolean) — whether execution should proceed  
   - `reason` (string) — justification for the decision  
   - `confidence` (float, 0.0–1.0) — estimated confidence  

3. **Payload Encryption / Embedding**  
   Payloads are pre-encrypted using RC4. Both the encrypted payload (`loader_encrypted.bin`) and key (`key.txt`) are embedded directly into the Go binary at compile time via `//go:embed`. No external files are required at runtime.

4. **Shellcode Execution**  
   If AI conditions are met, the shellcode is decrypted, allocated in executable memory via `VirtualAlloc`, written in place, and executed in a new thread with `CreateThread`. The host blocks on execution using `WaitForSingleObject`.

---

---

## Requirements

- Go 1.24 or newer  
- OpenAI API key (`OPENAI_API_KEY` environment variable)  
- Windows target environment (x86-64)  

---

## Build Process

HALO is intended for Windows (x86-64). Cross-compilation from Linux requires MinGW:

```bash
CGO_ENABLED=1 \
CC=x86_64-w64-mingw32-gcc \
GOOS=windows GOARCH=amd64 \
go build -o halo.exe .
```
During build:

loader_encrypted.bin and key.txt (in shellcode/) are embedded into the binary as byte slices.

The resulting halo.exe is a standalone artifact; it does not require external payload or key files.

Payload Encryption Utility (enc.go)
The enc.go utility transforms a raw payload into the format consumed by HALO.

```bash
go build -o encryptor enc.go
./encryptor <input_payload.bin> <key.txt> <output_encrypted.bin>
```

### Runtime Behavior

#### Initialization
- The program initializes logging and verifies the `OPENAI_API_KEY` environment variable.  

#### Baseline Collection
- The telemetry subsystem records process and system state, storing a baseline for later comparison.  

#### Decision Loop
1. A telemetry sample is collected.  
2. Telemetry is serialized to JSON and sent to the GPT model with the system prompt.  
3. The model’s response is parsed.  
4. Execution proceeds only if:  
   - `allow = true`  
   - `confidence ≥ 0.8`  

#### Decryption and Execution
- The embedded payload is decrypted with RC4 using the embedded key.  
- Memory is allocated with `VirtualAlloc` (`MEM_COMMIT | MEM_RESERVE`, `PAGE_EXECUTE_READWRITE`).  
- Shellcode is copied into the allocated region.  
- `CreateThread` is invoked with the shellcode entrypoint.  
- Execution is blocked until completion via `WaitForSingleObject`.  


Example Output
```bash
[*] AI-powered stealth payload started. Self: halo.exe, Parent: explorer.exe
[*] Collected telemetry. Querying GPT...
[*] Decision: true (confidence 0.91)
[*] Reason: Host state matches baseline, no monitoring tools present
[+] Execution approved. Launching payload inline.
[+] Shellcode execution returned cleanly.
```

### Customization

#### Model Selection
- Modify the model string in `ai/ai.go` (default: `gpt-4.1-mini`).  

#### Confidence Threshold
- Adjust the conditional check in `main.go` to change the minimum required confidence value.  

#### Telemetry Signals
- Extend `telemetry.go` to collect additional host environment attributes as needed.  

#### Payload Replacement
- Generate a new `loader_encrypted.bin` and `key.txt` with `enc.go`, then rebuild the project.  

