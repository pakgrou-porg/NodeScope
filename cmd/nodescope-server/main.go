package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pakgrou-porg/nodescope/internal/app"
	"github.com/pakgrou-porg/nodescope/internal/controlapi"
	"github.com/pakgrou-porg/nodescope/internal/mcpserver"
	"github.com/pakgrou-porg/nodescope/internal/proxy"
	"github.com/pakgrou-porg/nodescope/internal/telemetry"
)

func main() {
	config, err := app.LoadConfig(os.Getenv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "NodeScope server configuration error:", err)
		os.Exit(2)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger.Info("starting NodeScope server", "config", config.RedactedSummary(), "version", app.BuildVersion)

	options := make([]app.ServerOption, 0, 3)
	var runtimeDatabase *app.RuntimeDatabase
	if config.RuntimeDatabaseURL != "" {
		database, openError := app.OpenRuntimeDatabase(context.Background(), config.RuntimeDatabaseURL)
		if openError != nil {
			logger.Error("open NodeScope runtime database", "error", openError)
			os.Exit(1)
		}
		defer database.Close()
		runtimeDatabase = database
		options = append(options, app.WithDatabase(database), app.WithIngestor(telemetry.NewPostgresStore(database.Pool())))
	}
	if config.APIConfigPath != "" {
		if runtimeDatabase == nil {
			logger.Error("control API requires the NodeScope runtime database")
			os.Exit(2)
		}
		apiConfiguration, apiError := mcpserver.LoadHTTPConfiguration(config.APIConfigPath)
		if apiError != nil {
			logger.Error("load protected control API configuration", "error", apiError)
			os.Exit(2)
		}
		options = append(options, app.WithControlAPI(controlapi.Handler{
			Service:       mcpserver.NewPostgresService(runtimeDatabase.Pool()),
			Authenticator: apiConfiguration.Authenticator(),
		}))
	}
	if config.MCPConfigPath != "" {
		if runtimeDatabase == nil {
			logger.Error("MCP server requires the NodeScope runtime database")
			os.Exit(2)
		}
		mcpConfiguration, mcpError := mcpserver.LoadHTTPConfiguration(config.MCPConfigPath)
		if mcpError != nil {
			logger.Error("load protected MCP configuration", "error", mcpError)
			os.Exit(2)
		}
		mcpTools := mcpserver.Server{Service: mcpserver.NewPostgresService(runtimeDatabase.Pool())}.New()
		options = append(options, app.WithMCP(mcpserver.NewHTTPHandler(mcpTools, mcpConfiguration.Authenticator())))
	}
	if config.ProxyConfigPath != "" {
		if runtimeDatabase == nil {
			logger.Error("inference proxy requires the NodeScope runtime database")
			os.Exit(2)
		}
		proxyConfiguration, proxyError := proxy.LoadFileConfiguration(config.ProxyConfigPath)
		if proxyError != nil {
			logger.Error("load protected inference proxy configuration", "error", proxyError)
			os.Exit(2)
		}
		options = append(options, app.WithInferenceProxy(&proxy.Handler{
			Registry:      proxyConfiguration.Registry(),
			Authenticator: proxyConfiguration.Authenticator(),
			Recorder:      proxy.NewPostgresRecorder(runtimeDatabase.Pool()),
		}))
	}

	tlsConfig, tlsError := app.BuildServerTLSConfig(config)
	if tlsError != nil {
		logger.Error("configure NodeScope TLS", "error", tlsError)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              config.ListenAddress,
		Handler:           app.NewServer(config, logger, options...).Handler(),
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errorsChannel := make(chan error, 1)
	go func() {
		if config.CertificatePath != "" {
			errorsChannel <- server.ListenAndServeTLS(config.CertificatePath, config.PrivateKeyPath)
			return
		}
		errorsChannel <- server.ListenAndServe()
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case signalError := <-errorsChannel:
		if !errors.Is(signalError, http.ErrServerClosed) {
			logger.Error("NodeScope server stopped unexpectedly", "error", signalError)
			os.Exit(1)
		}
	case <-shutdownSignal.Done():
		logger.Info("shutting down NodeScope server")
		shutdownContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if shutdownError := server.Shutdown(shutdownContext); shutdownError != nil {
			logger.Error("NodeScope server graceful shutdown failed", "error", shutdownError)
			os.Exit(1)
		}
	}
}
