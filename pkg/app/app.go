package app

import (
	"context"
	"flag"
	"github.com/compliance-webhook/pkg/handler"
	"net/http"
)

func Run(ctx context.Context) {
	http.HandleFunc("/mutate", handler.WebhookHandler)
	var tlsKey, tlsCert string
	flag.StringVar(&tlsKey, "tlsKey", "tls.key", "Path to the TLS key")
	flag.StringVar(&tlsCert, "tlsCert", "tls.crt", "Path to the TLS certificate")
	err := http.ListenAndServeTLS(":8443", tlsCert, tlsKey, nil)
	if err != nil {
		return
	}
}
