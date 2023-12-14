package main

import (
	"io"
	"log"
	"net/http"
)

const moduleContent = `
declare "myModule" {
	argument "scrape_endpoint" {}
  
	argument "forward_to" {}
  
	argument "scrape_interval" {
	  optional = true
	  default  = "1s"
	}
  
	prometheus.scrape "scrape_prom_metrics_module_file" {
	  targets = [
		{"__address__" = argument.scrape_endpoint.value},
	  ]
	  forward_to = argument.forward_to.value
	  scrape_classic_histograms = true
	  enable_protobuf_negotiation = true
	  scrape_interval = argument.scrape_interval.value
	  scrape_timeout = "500ms"
	}
  }
`

func main() {
	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, moduleContent)
	})
	log.Fatal(http.ListenAndServe(":8090", nil))
}
