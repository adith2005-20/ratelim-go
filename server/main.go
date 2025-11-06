package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/adith2005-20/ratelim-go/internal/limiter"
)

type Config struct {
	Rate int	`json:"rate"`
	Slack int	`json:"slack"`
	Per time.Duration `json:"per"`
	Port string	`json:"port"`
}


func loadConfig() Config {
	file, err := os.Open("../../config.json")
	if err != nil {
		log.Println("Config not found, reverting to defaults.")
		return Config{Rate: 5, Slack: 10, Per: time.Second, Port:":8080"}
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		log.Fatalf("Failed to parse config.json: %v", err)
	}
	if cfg.Port == "" {
		cfg.Port = ":8080"
	}
	return cfg
}

func main () {
	config :=loadConfig()
	log.Printf("Ratelim starting...")
	lim:=limiter.New(config.Rate, limiter.WithSlack(config.Slack), limiter.Per(config.Per))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		t:=lim.Take()

		fmt.Fprintf(w, "Request allowed at %v\n", t.Format(time.RFC3339Nano))
		log.Printf("Request from %s allowed at %v", r.RemoteAddr, t)
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Printf("Listening on %s...", config.Port)
	if err := http.ListenAndServe(config.Port, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
