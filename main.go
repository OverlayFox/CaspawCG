package main

import (
	"log"

	cp "github.com/OverlayFox/CaspawCG/casperProxy"
)

const (
	proxyPort    = "5251"
	casparCGHost = "127.0.0.1"
	casparCGPort = "5250"
)

func main() {
	proxy, err := cp.NewProxy(proxyPort, casparCGHost, casparCGPort)
	if err != nil {
		log.Fatalf("Failed to create proxy: %v", err)
	}

	log.Printf("Starting CasparCG AMCP Proxy on port %s, forwarding to %s:%s", proxyPort, casparCGHost, casparCGPort)

	if err := proxy.Start(); err != nil {
		log.Fatalf("Error starting proxy: %v", err)
	}
}

// interceptAndModifyCommand is where your core logic lives
// func interceptAndModifyCommand(command string) string {
// 	// Example: Intercept CG ADD commands and modify their data
// 	if strings.HasPrefix(command, "CG ADD") {
// 		parts := strings.SplitN(command, " ", 5) // Split into "CG", "ADD", channel-layer, cg_layer, template_path, data
// 		if len(parts) >= 5 {
// 			templatePathAndPlayFlag := parts[4] // e.g., "template_name" 1 "data_string"

// 			if strings.Contains(command, "\"") {
// 				lastQuoteIndex := strings.LastIndex(command, "\"")
// 				firstQuoteIndex := strings.LastIndex(command[:lastQuoteIndex], "\"")

// 				if firstQuoteIndex != -1 && lastQuoteIndex != -1 && lastQuoteIndex > firstQuoteIndex {
// 					originalDataStr := command[firstQuoteIndex+1 : lastQuoteIndex]
// 					log.Printf("Detected original data string: %s", originalDataStr)

// 					// Attempt to parse as JSON
// 					if strings.HasPrefix(originalDataStr, "{") && strings.HasSuffix(originalDataStr, "}") {
// 						fmt.Sprintln(originalDataStr)
// 						return command[:firstQuoteIndex+1] + modifiedDataStr + command[lastQuoteIndex:]
// 					} else if strings.HasPrefix(originalDataStr, "<") && strings.HasSuffix(originalDataStr, ">") {
// 						// This would require XML parsing and modification
// 						log.Println("XML data detected, but XML modification logic not implemented in this example.")
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return command // Return original command if not intercepted or not a CG ADD with modifiable data
// }
