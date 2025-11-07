package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"gopkg.in/yaml.v3"
	"sync"

	"github.com/adith2005-20/ratelim-go/internal/limiter"
)

type AppConfig struct {
	Rate int	`yaml:"rate"`
	Slack int	`yaml:"slack"`
	Per time.Duration `yaml:"per"`
	Port string	`yaml:"port"`
}

type Config struct {
	Apps map[string]AppConfig `yaml:"apps"`
	Port string	`yaml:"port"`
}


func loadConfig() Config {
	file, err := os.Open("config.yaml")
	if err != nil {
		log.Println("config.yaml not found, reverting to defaults.",err)
		return Config{
			Apps: map[string]AppConfig{
				"default": {Rate: 5, Slack: 10, Per: time.Second},
			},
			Port: ":8080",
		}
	}
	defer file.Close()

	var cfg Config
	if err := yaml.NewDecoder(file).Decode(&cfg); err != nil {
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

	limiters:=make(map[string]limiter.Limiter)
	var mu sync.Mutex


	getLimiter:= func(appID string) limiter.Limiter {
		mu.Lock()
		defer mu.Unlock()

		if lim, ok := limiters[appID]; ok {
			return lim
		}

		appCfg, ok:= config.Apps[appID]

		if !ok {
			appCfg = config.Apps["default"]
			log.Printf("No specific config for '%s', using default", appID)
		}
		lim:= limiter.New(appCfg.Rate, limiter.WithSlack(appCfg.Slack), limiter.Per(appCfg.Per))
		limiters[appID] = lim
		return lim
	}

	

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		appID := r.Header.Get("X-App-ID")
		if appID=="" {
			http.Error(w,"Missing X-App-ID header. Try again.", http.StatusBadRequest)
			return
		}
		
		lim:= getLimiter(appID)
		t:=lim.Take()

		fmt.Fprintf(w, "App %s Request allowed at %v\n",appID, t.Format(time.RFC3339Nano))
		log.Printf("[%s] Request from %s allowed at %v",appID, r.RemoteAddr, t)
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
