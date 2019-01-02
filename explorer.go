// Copyright (c) 2018, Migwi Ndung'u
// See LICENSE for details.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/decred/dcrd/rpcclient"
	"github.com/gorilla/mux"
	"github.com/raedahgroup/dcrchainanalysis/v1/analytics"
	"github.com/raedahgroup/dcrchainanalysis/v1/rpcutils"
)

// explorer defines all the content needed to effectively serve http requests.
type explorer struct {
	Client      *rpcclient.Client
	RPCVersion  *rpcutils.RPCVersion
	Params      *config
	OtherParams *extraParams
}

// healthHandler helps checks if the system is up and running.
func (exp *explorer) HealthHandler(w http.ResponseWriter, r *http.Request) {
	healthMsg := []byte(`{"health": "Thanks for checking. I am still alive.",` +
		`"probability": "/api/v1/{tx-hash}", "solutions": "/api/v1/{tx-hash}/all" }`)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(healthMsg)
}

// StatusHandler handles the various system statuses supported.
func (exp *explorer) StatusHandler(w http.ResponseWriter, r *http.Request, err error) {
	log.Error(err)
	logErrorMsg := []byte(`{"error": "Oops! Something went wrong, try different` +
		` inputs or contact system maintainers if problem persist."}`)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnprocessableEntity)
	w.Write(logErrorMsg)
}

// AllTxSolutionsHandler fetches analyzed transactions inputs and outputs returning
// all the possible solutions generated.
func (exp *explorer) AllTxSolutionsHandler(w http.ResponseWriter, r *http.Request) {
	transactionX := mux.Vars(r)["tx"]

	txData, err := analytics.RetrieveTxData(exp.Client, transactionX)
	if err != nil {
		exp.StatusHandler(w, r, err)
		return
	}

	rawTxSolution, _, _, err := analytics.TransactionFundsFlow(txData)
	if err != nil {
		exp.StatusHandler(w, r, err)
		return
	}

	strData, err := json.Marshal(rawTxSolution)
	if err != nil {
		exp.StatusHandler(w, r, fmt.Errorf("error occured: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(strData)
}

// TxProbabilityHandler from the fetched analyzed solutions, it returns the solution
// with the lowest granularity as the best solution.
func (exp *explorer) TxProbabilityHandler(w http.ResponseWriter, r *http.Request) {
	transactionX := mux.Vars(r)["tx"]

	solProbability, _, err := analytics.RetrieveTxProbability(exp.Client, transactionX)
	if err != nil {
		exp.StatusHandler(w, r, err)
		return
	}

	strData, err := json.Marshal(solProbability)
	if err != nil {
		exp.StatusHandler(w, r, fmt.Errorf("error occured: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(strData)
}

// ChainHandler reconstructs the probability solution to create funds flow paths.
func (exp *explorer) ChainHandler(w http.ResponseWriter, r *http.Request) {
	transactionX := mux.Vars(r)["tx"]

	chain, err := analytics.ChainDiscovery(exp.Client, transactionX)
	if err != nil {
		exp.StatusHandler(w, r, err)
		return
	}

	strData, err := json.Marshal(chain)
	if err != nil {
		exp.StatusHandler(w, r, fmt.Errorf("error occured: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(strData)
}

// PprofHandler fetches the correct pprof handler needed.
func (exp *explorer) PprofHandler(w http.ResponseWriter, r *http.Request) {
	handlerType := mux.Vars(r)["name"]
	switch strings.ToLower(handlerType) {
	case "pprof":
		pprof.Index(w, r)
	case "trace":
		pprof.Trace(w, r)
	case "profile":
		pprof.Profile(w, r)
	default:
		pprof.Handler(handlerType).ServeHTTP(w, r)
	}
}
