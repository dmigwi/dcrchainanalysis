// Copyright (c) 2018, Migwi Ndung'u
// See LICENSE for details.
package main

import (
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/raedahgroup/dcrchainanalysis/v1/rpcutils"
)

// start sets up the explorer.
func start() (*explorer, error) {
	cfg, otherCfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	log.Info("Starting up the Chain Analysis Tool")

	client, rpcVersion, err := rpcutils.ConnectRPCNode(cfg.DcrdServ, cfg.DcrdUser,
		cfg.DcrdPass, cfg.DcrdCert, cfg.DisableDaemonTLS, nil)
	if err != nil {
		return nil, err
	}

	log.Infof("Connected to a dcrd node successfully: %s, %s",
		otherCfg.ActiveNet.String(), rpcVersion.String())

	exp := &explorer{
		Client:      client,
		RPCVersion:  rpcVersion,
		Params:      cfg,
		OtherParams: otherCfg,
	}

	return exp, nil
}

// main initaites program execution.
func main() {
	expl, err := start()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", expl.HealthHandler)
	r.HandleFunc("/api/v1/{tx}", expl.TxProbabilityHandler)
	r.HandleFunc("/api/v1/{tx}/all", expl.AllTxSolutionsHandler)

	if expl.Params.CPUProfile {
		log.Debug("CPU profiling Activated. Setting up the Profiling.")

		r.HandleFunc("/debug/{name}", expl.PprofHandler)
	}
	// Return the health page for all 404s
	r.NotFoundHandler = http.HandlerFunc(expl.HealthHandler)

	server := &http.Server{
		Handler:      r,
		Addr:         expl.Params.DCAHost,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	log.Info("Server running :", expl.Params.DCAHost)

	// start server in a go routine.
	go func() {
		if err = server.ListenAndServe(); err != nil {
			log.Error(err)
			os.Exit(1)
		}
	}()

	c := make(chan os.Signal, 1)
	defer close(c)

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	log.Info("(Ctrl+C) pressed")
	log.Info("Bye, System shutting down")
	os.Exit(0)
}
