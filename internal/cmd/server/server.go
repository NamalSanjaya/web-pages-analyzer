package server

import (
	"log"
	"net/http"
	wpac "web-pages-analyzer/internal/controllers/webpage_analyzer"
	dmhttp "web-pages-analyzer/internal/domain/clients/http"
	clihttp "web-pages-analyzer/internal/infrastructure/clients/http"
	wpa "web-pages-analyzer/internal/usecases/webpage_analyzer"
)

func Start() {

	cfg := &dmhttp.HttpClientCfg{
		Timeout:      10,
		MaxRedirects: 5,
	}

	// Create singleton instances
	httpclient := clihttp.New(cfg)
	wpaUsecase := wpa.New(httpclient)
	wpaCtrler := wpac.New(wpaUsecase)

	http.HandleFunc("/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			wpaCtrler.Analyze(w, r)
			return
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})

	hostPort := ":8080"
	log.Printf("Server starting on port %s\n", hostPort)
	log.Fatal(http.ListenAndServe(hostPort, nil))
}
