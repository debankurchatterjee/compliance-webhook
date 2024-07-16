package app

import (
	"context"
	"crypto/tls"
	"flag"
	"github.com/compliance-webhook/internal/logutil/log"
	"github.com/compliance-webhook/pkg/handler"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func RunUnsecure(ctx context.Context) {
	http.HandleFunc("/mutate", handler.WebhookHandler)
	logger := log.From(ctx).WithName("webhook-server")
	server := &http.Server{
		Addr: ":8001",
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		logger.Info("starting webhook server")
		err := server.ListenAndServe()
		if err != nil {
			logger.Error(err, "failed to start webhook server")
			return
		}
	}()
	<-stop
	// Create a context with a timeout for the shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Attempt a graceful shutdown
	logger.Info("Shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error(err, "Server forced to shutdown")
	}

	logger.Info("Server exiting")
}

func Run(ctx context.Context) {
	http.HandleFunc("/mutate", handler.WebhookHandler)
	logger := log.From(ctx).WithName("webhook-server")
	server := &http.Server{
		Addr: ":8443",
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		var tlsKey, tlsCert string
		flag.StringVar(&tlsKey, "tlsKey", "/certs/tls.key", "Path to the TLS key")
		flag.StringVar(&tlsCert, "tlsCert", "/certs/tls.crt", "Path to the TLS certificate")
		logger.Info("starting webhook server")
		//decodeString1, err := base64.StdEncoding.DecodeString(tlsKey)
		//if err != nil {
		//	return
		//}
		//decodeString2, err := base64.StdEncoding.DecodeString(tlsCert)
		//if err != nil {
		//	return
		//}
		err := server.ListenAndServeTLS(tlsCert, tlsKey)
		if err != nil {
			logger.Error(err, "failed to start webhook server")
			return
		}
	}()
	<-stop
	// Create a context with a timeout for the shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt a graceful shutdown
	logger.Info("Shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error(err, "Server forced to shutdown")
	}

	logger.Info("Server exiting")
}
