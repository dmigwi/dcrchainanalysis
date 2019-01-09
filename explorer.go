// Copyright (c) 2018, Migwi Ndung'u
// See LICENSE for details.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/decred/dcrd/rpcclient"
	"github.com/gorilla/mux"
	"github.com/raedahgroup/dcrchainanalysis/v1/analytics"
	"github.com/raedahgroup/dcrchainanalysis/v1/rpcutils"
)

const (
	healthMsg = `{` +
		`"health": "Thanks for checking. Still alive.",` +
		`"probability": "/api/v1/{tx-hash}", ` +
		`"raw solutions": "/api/v1/{tx-hash}/all",` +
		`"all paths": "/api/v1/{tx}/chain",` +
		`"single path": "/api/v1/{tx}/chain/{index}"}`

	defaultErrorMsg = `{"error": "Oops! Something went wrong, try different ` +
		`inputs or contact system maintainers if problem persists.",` +
		`"duration":"%s"}`
)

// TimeData defines the time data type that holds the block time from the
// actual tx and time taken to process a given payload.
type TimeData struct {
	TxTime   int64  `json:",omitempty"`
	Duration string `json:",omitempty"`
}

// explorer defines all the content needed to effectively serve http requests.
type explorer struct {
	Client      *rpcclient.Client
	RPCVersion  *rpcutils.RPCVersion
	Params      *config
	OtherParams *extraParams
}

// rawSolution defines the full structure of final raw solution(single tx
// analyzed solution).
type rawSolution struct {
	TimeData
	Data []*analytics.AllFundsFlows
}

// probabilitySolution defines the full structure of the probability solution
// for a single tx that is based of the raw data solution.
type probabilitySolution struct {
	TimeData
	Data []*analytics.FlowProbability
}

// pathSolution is the funds flow solution that just a chain of probability
// solutions linked together.
type pathSolution struct {
	TimeData
	Data []*analytics.Hub
}

// healthHandler helps checks if the system is up and running.
func (exp *explorer) HealthHandler(w http.ResponseWriter, r *http.Request) {
	jsonWrite([]byte(healthMsg), http.StatusOK, w)
}

// StatusHandler handles the various system statuses supported.
func (exp *explorer) StatusHandler(w http.ResponseWriter, r *http.Request,
	startTime time.Time, err error) {
	log.Error(err)

	data := fmt.Sprintf(defaultErrorMsg, durationInSec(startTime))
	jsonWrite([]byte(data), http.StatusUnprocessableEntity, w)
}

// AllTxSolutionsHandler fetches analyzed transactions inputs and outputs returning
// all the possible solutions generated(raw tx solution).
func (exp *explorer) AllTxSolutionsHandler(w http.ResponseWriter, r *http.Request) {
	transactionX := mux.Vars(r)["tx"]
	t := time.Now()

	txData, err := analytics.RetrieveTxData(exp.Client, transactionX)
	if err != nil {
		exp.StatusHandler(w, r, t, err)
		return
	}

	rawTxSolution, _, _, err := analytics.TransactionFundsFlow(txData)
	if err != nil {
		exp.StatusHandler(w, r, t, err)
		return
	}

	exp.handleJSONWrite(
		rawSolution{
			Data: rawTxSolution,
			TimeData: TimeData{
				TxTime: txData.BlockTime, Duration: durationInSec(t),
			},
		},
		http.StatusOK, t, w, r)
}

// TxProbabilityHandler from the fetched analyzed solutions, it returns the solution
// with the lowest granularity as the best solution.
func (exp *explorer) TxProbabilityHandler(w http.ResponseWriter, r *http.Request) {
	transactionX := mux.Vars(r)["tx"]
	t := time.Now()

	solProbability, txData, err := analytics.RetrieveTxProbability(exp.Client, transactionX)
	if err != nil {
		exp.StatusHandler(w, r, t, err)
		return
	}

	exp.handleJSONWrite(
		probabilitySolution{
			Data: solProbability,
			TimeData: TimeData{
				TxTime: txData.BlockTime, Duration: durationInSec(t),
			},
		},
		http.StatusOK, t, w, r)
}

// ChainHandler reconstructs the probability solution to create funds flow paths.
func (exp *explorer) ChainHandler(w http.ResponseWriter, r *http.Request) {
	transactionX := mux.Vars(r)["tx"]
	t := time.Now()

	chain, TxTime, err := analytics.ChainDiscovery(exp.Client, transactionX)
	if err != nil {
		exp.StatusHandler(w, r, t, err)
		return
	}

	exp.handleJSONWrite(
		pathSolution{
			Data: chain,
			TimeData: TimeData{
				TxTime: TxTime, Duration: durationInSec(t),
			},
		},
		http.StatusOK, t, w, r)
}

// ChainPathHandler reconstructs the probability solution to create one funds
// flow path on the provided outputs index. If the index provided in greater than
// the available output index, the outputs path with the last index is returned.
func (exp *explorer) ChainPathHandler(w http.ResponseWriter, r *http.Request) {
	transactionX := mux.Vars(r)["tx"]
	t := time.Now()

	txIndex, err := strconv.Atoi(mux.Vars(r)["index"])
	if err != nil {
		exp.StatusHandler(w, r, t, err)
		return
	}

	chain, TxTime, err := analytics.ChainDiscovery(exp.Client, transactionX, txIndex)
	if err != nil {
		exp.StatusHandler(w, r, t, err)
		return
	}

	exp.handleJSONWrite(
		pathSolution{
			Data: chain,
			TimeData: TimeData{
				TxTime: TxTime, Duration: durationInSec(t),
			},
		},
		http.StatusOK, t, w, r)
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

// handleJSONWrite processes the payload data.
func (exp *explorer) handleJSONWrite(rawData interface{}, status int,
	t time.Time, w http.ResponseWriter, r *http.Request) {

	byteData, err := json.Marshal(rawData)
	if err != nil {
		exp.StatusHandler(w, r, t, fmt.Errorf("error occured: %v", err))
		return
	}

	jsonWrite(byteData, status, w)
}

// jsonWrite sends back the json payload.
func jsonWrite(data []byte, status int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(data)
}

// durationInSec calculates the duration in seconds.
func durationInSec(t time.Time) string {
	d := time.Since(t)
	return strconv.FormatFloat(d.Seconds(), 'f', -1, 64) + "s"
}
