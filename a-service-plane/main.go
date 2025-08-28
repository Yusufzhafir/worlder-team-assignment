package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

type StatsResponse struct {
	Data struct {
		ConfiguredFreqMs float64 `json:"Configured_FreqMs"`
		IsRunning        bool    `json:"Is_Running"`
		OverallRPS       float64 `json:"Overall_RPS"`
		TotalFailed      float64 `json:"Total_Failed"`
		TotalRequests    float64 `json:"Total_Requests"`
		TotalSent        float64 `json:"Total_Sent"`
		UptimeSeconds    float64 `json:"Uptime_Seconds"`
	} `json:"data"`
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type AggregatedStats struct {
	ConfiguredFreqMs struct{ Min, Max, Avg float64 }
	IsRunning        struct{ Running, Total float64 }
	OverallRPS       struct{ Min, Max, Avg, Total float64 }
	TotalFailed      struct{ Min, Max, Avg, Total float64 }
	TotalRequests    struct{ Min, Max, Avg, Total float64 }
	TotalSent        struct{ Min, Max, Avg, Total float64 }
	UptimeSeconds    struct{ Min, Max, Avg, Total float64 }
}

var rootCmd = &cobra.Command{
	Use:   "a-plane",
	Short: "CLI tool for managing API endpoints across multiple ports",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(frequencyCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statsCmd)
}

var frequencyCmd = &cobra.Command{
	Use:   "frequency [duration]",
	Short: "Set frequency for all endpoints",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		duration := args[0]

		payload := map[string]string{
			"timeout": duration,
		}

		makeRequestToAllPorts("POST", "/api/v1/frequency", payload)
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start all endpoints",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		makeRequestToAllPorts("POST", "/api/v1/start", nil)
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop all endpoints",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		makeRequestToAllPorts("POST", "/api/v1/stop", nil)
	},
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Get aggregated stats from all endpoints",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		getAggregatedStats()
	},
}

func makeRequestToAllPorts(method, endpoint string, payload interface{}) {
	var wg sync.WaitGroup

	for port := 9000; port <= 9009; port++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()

			url := fmt.Sprintf("http://localhost:%d%s", p, endpoint)

			var body io.Reader
			if payload != nil {
				jsonData, err := json.Marshal(payload)
				if err != nil {
					log.Printf("Error marshaling payload for port %d: %v", p, err)
					return
				}
				body = bytes.NewBuffer(jsonData)
			}

			req, err := http.NewRequest(method, url, body)
			if err != nil {
				log.Printf("Error creating request for port %d: %v", p, err)
				return
			}

			req.Header.Set("Accept", "application/json")
			if payload != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Error making request to port %d: %v", p, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				fmt.Printf("✓ Port %d: Success\n", p)
			} else {
				fmt.Printf("✗ Port %d: HTTP %d\n", p, resp.StatusCode)
			}
		}(port)
	}

	wg.Wait()
	fmt.Println("All requests completed.")
}

func getAggregatedStats() {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var responses []StatsResponse

	for port := 9000; port <= 9009; port++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()

			url := fmt.Sprintf("http://localhost:%d/api/v1/stats", p)

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Printf("Error creating request for port %d: %v", p, err)
				return
			}

			req.Header.Set("Accept", "application/json")

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Error making request to port %d: %v", p, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				log.Printf("Port %d returned HTTP %d", p, resp.StatusCode)
				return
			}

			var statsResp StatsResponse
			if err := json.NewDecoder(resp.Body).Decode(&statsResp); err != nil {
				log.Printf("Error decoding response from port %d: %v", p, err)
				return
			}

			mu.Lock()
			responses = append(responses, statsResp)
			mu.Unlock()
		}(port)
	}

	wg.Wait()

	if len(responses) == 0 {
		fmt.Println("No successful responses received")
		return
	}

	aggregated := aggregateStats(responses)
	printAggregatedStats(aggregated, len(responses))
}

func aggregateStats(responses []StatsResponse) AggregatedStats {
	var agg AggregatedStats

	if len(responses) == 0 {
		return agg
	}

	// Initialize min/max values
	first := responses[0].Data
	agg.ConfiguredFreqMs.Min = first.ConfiguredFreqMs
	agg.ConfiguredFreqMs.Max = first.ConfiguredFreqMs
	agg.OverallRPS.Min = first.OverallRPS
	agg.OverallRPS.Max = first.OverallRPS
	agg.TotalFailed.Min = first.TotalFailed
	agg.TotalFailed.Max = first.TotalFailed
	agg.TotalRequests.Min = first.TotalRequests
	agg.TotalRequests.Max = first.TotalRequests
	agg.TotalSent.Min = first.TotalSent
	agg.TotalSent.Max = first.TotalSent
	agg.UptimeSeconds.Min = first.UptimeSeconds
	agg.UptimeSeconds.Max = first.UptimeSeconds

	// Aggregate all values
	for _, resp := range responses {
		data := resp.Data

		// ConfiguredFreqMs
		if data.ConfiguredFreqMs < agg.ConfiguredFreqMs.Min {
			agg.ConfiguredFreqMs.Min = data.ConfiguredFreqMs
		}
		if data.ConfiguredFreqMs > agg.ConfiguredFreqMs.Max {
			agg.ConfiguredFreqMs.Max = data.ConfiguredFreqMs
		}
		agg.ConfiguredFreqMs.Avg += data.ConfiguredFreqMs

		// IsRunning
		if data.IsRunning {
			agg.IsRunning.Running++
		}
		agg.IsRunning.Total++

		// OverallRPS
		if data.OverallRPS < agg.OverallRPS.Min {
			agg.OverallRPS.Min = data.OverallRPS
		}
		if data.OverallRPS > agg.OverallRPS.Max {
			agg.OverallRPS.Max = data.OverallRPS
		}
		agg.OverallRPS.Avg += data.OverallRPS
		agg.OverallRPS.Total += data.OverallRPS

		// TotalFailed
		if data.TotalFailed < agg.TotalFailed.Min {
			agg.TotalFailed.Min = data.TotalFailed
		}
		if data.TotalFailed > agg.TotalFailed.Max {
			agg.TotalFailed.Max = data.TotalFailed
		}
		agg.TotalFailed.Avg += data.TotalFailed
		agg.TotalFailed.Total += data.TotalFailed

		// TotalRequests
		if data.TotalRequests < agg.TotalRequests.Min {
			agg.TotalRequests.Min = data.TotalRequests
		}
		if data.TotalRequests > agg.TotalRequests.Max {
			agg.TotalRequests.Max = data.TotalRequests
		}
		agg.TotalRequests.Avg += data.TotalRequests
		agg.TotalRequests.Total += data.TotalRequests

		// TotalSent
		if data.TotalSent < agg.TotalSent.Min {
			agg.TotalSent.Min = data.TotalSent
		}
		if data.TotalSent > agg.TotalSent.Max {
			agg.TotalSent.Max = data.TotalSent
		}
		agg.TotalSent.Avg += data.TotalSent
		agg.TotalSent.Total += data.TotalSent

		// UptimeSeconds
		if data.UptimeSeconds < agg.UptimeSeconds.Min {
			agg.UptimeSeconds.Min = data.UptimeSeconds
		}
		if data.UptimeSeconds > agg.UptimeSeconds.Max {
			agg.UptimeSeconds.Max = data.UptimeSeconds
		}
		agg.UptimeSeconds.Avg += data.UptimeSeconds
		agg.UptimeSeconds.Total += data.UptimeSeconds
	}

	// Calculate averages
	count := len(responses)
	agg.ConfiguredFreqMs.Avg /= float64(count)
	agg.OverallRPS.Avg /= float64(count)
	agg.TotalFailed.Avg /= float64(count)
	agg.TotalRequests.Avg /= float64(count)
	agg.TotalSent.Avg /= float64(count)
	agg.UptimeSeconds.Avg /= float64(count)

	return agg
}

func printAggregatedStats(agg AggregatedStats, responseCount int) {
	fmt.Printf("\n=== Aggregated Stats (from %d endpoints) ===\n\n", responseCount)

	fmt.Printf("Configured Frequency (ms):\n")
	fmt.Printf("  Min: %f, Max: %f, Avg: %f\n\n",
		agg.ConfiguredFreqMs.Min, agg.ConfiguredFreqMs.Max, agg.ConfiguredFreqMs.Avg)

	fmt.Printf("Running Status:\n")
	fmt.Printf("  Running: %f/%f endpoints\n\n", agg.IsRunning.Running, agg.IsRunning.Total)

	fmt.Printf("Overall RPS:\n")
	fmt.Printf("  Min: %.2f, Max: %.2f, Avg: %.2f, Total: %.2f\n\n",
		agg.OverallRPS.Min, agg.OverallRPS.Max, agg.OverallRPS.Avg, agg.OverallRPS.Total)

	fmt.Printf("Total Failed:\n")
	fmt.Printf("  Min: %f, Max: %f, Avg: %f, Total: %f\n\n",
		agg.TotalFailed.Min, agg.TotalFailed.Max, agg.TotalFailed.Avg, agg.TotalFailed.Total)

	fmt.Printf("Total Requests:\n")
	fmt.Printf("  Min: %f, Max: %f, Avg: %f, Total: %f\n\n",
		agg.TotalRequests.Min, agg.TotalRequests.Max, agg.TotalRequests.Avg, agg.TotalRequests.Total)

	fmt.Printf("Total Sent:\n")
	fmt.Printf("  Min: %f, Max: %f, Avg: %f, Total: %f\n\n",
		agg.TotalSent.Min, agg.TotalSent.Max, agg.TotalSent.Avg, agg.TotalSent.Total)

	fmt.Printf("Uptime (seconds):\n")
	fmt.Printf("  Min: %f, Max: %f, Avg: %f, Total: %f\n\n",
		agg.UptimeSeconds.Min, agg.UptimeSeconds.Max, agg.UptimeSeconds.Avg, agg.UptimeSeconds.Total)
}
