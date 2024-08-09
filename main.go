package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ipcPath          string
	debug            bool
	port             int
	blockNumberGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "assetchain_block_number",
			Help: "Current block number of the Asset Chain Opera blockchain",
		},
	)
)

func init() {
	// Register the gauge with Prometheus's default registry.
	prometheus.MustRegister(blockNumberGauge)
}

func getBlockNumber() (int64, error) {
	// Run the command to get the block number from the assetchain Opera node.
	cmd := exec.Command("sh", "-c", fmt.Sprintf(`echo "ftm.blockNumber" | opera attach "%s" | tail -n 2 | grep -o -E '[0-9]+'`, ipcPath))
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	if debug {
		log.Printf("Raw output from opera attach: %s", output)
	}

	// Parse the block number from the command output.
	blockNumberStr := strings.TrimSpace(string(output))
	blockNumber, err := strconv.ParseInt(blockNumberStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return blockNumber, nil
}

func updateBlockNumber() {
	blockNumber, err := getBlockNumber()
	if err != nil {
		log.Printf("Error getting block number: %v", err)
		return
	}
	blockNumberGauge.Set(float64(blockNumber))

	if debug {
		log.Printf("Updated block number: %d", blockNumber)
	}
}

func main() {
	// Define flags for the config path and debug mode.
	flag.StringVar(&ipcPath, "config.ipcpath", "/home/ubuntu/opera.ipc", "Path to the opera.ipc file")
	flag.BoolVar(&debug, "config.debug", false, "Enable debug mode")
	flag.IntVar(&port, "config.port", 8080, "Port for metrics")
	flag.Parse()

	if debug {
		log.Printf("Starting with IPC path: %s", ipcPath)
	}

	// Start a goroutine to update the block number periodically.
	go func() {
		for {
			updateBlockNumber()
			// Update every 30 seconds
			time.Sleep(60 * time.Second)
		}
	}()

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%d", port), nil))
	log.Printf("Listening on port %d", port)
}
