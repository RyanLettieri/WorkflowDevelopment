package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
)

// Placeholder string for the task queue
const BankTaskQueue = "BankTaskQueue"

// Placeholder string for the task queue
const LongTaskQueue = "LongTaskQueue"

var temporalClient client.Client

func main() {
	var dir string

	// Create the workflow client
	cOptions := client.Options{HostPort: "localhost:7233"}
	c, err := client.Dial(cOptions)
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	temporalClient = c

	flag.StringVar(&dir, "dir", ".", "the directory to serve files from. Defaults to the current dir")
	flag.Parse()
	r := mux.NewRouter()

	// This will serve files under http://localhost:8000/static/<filename>
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(dir))))

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	setHandlers(r)

	// Create the worker
	w := worker.New(temporalClient, BankTaskQueue, worker.Options{Identity: "workerOne"})
	// Register a workflow and activities for the worker
	w.RegisterWorkflow(BankTransfer)
	w.RegisterActivity(GetBankBalance)
	w.RegisterActivityWithOptions(WriteFunc, activity.RegisterOptions{Name: "writeFunc"})
	err = w.Start()
	if err != nil {
		return
	}

	w2 := worker.New(temporalClient, LongTaskQueue, worker.Options{Identity: "workerTwo"})
	// Register a workflow and activities for the worker
	w2.RegisterWorkflow(longRunning)
	w2.RegisterActivityWithOptions(longRunningAct, activity.RegisterOptions{Name: "LongRunning"})
	err = w2.Start()
	if err != nil {
		return
	}
	log.Fatal(srv.ListenAndServe())
}

// POST http://localhost:3500/v1.0/workflows/{workflowType}/{instanceId}/start

func setHandlers(r *mux.Router) {
	r.HandleFunc("/workflows/{workflowID}/{instanceId}/start", startWorkflow).Methods("POST")
	r.HandleFunc("/workflows/{workflowID}/{instanceId}/terminate", terminateWorkFlow).Methods("POST", "DELETE")
	r.HandleFunc("/workflows/{workflowID}/{instanceId}/register", registerWorkFlow).Methods("POST")
	r.HandleFunc("/workflows/{workflowID}/{instanceId}/getWorkflowStats", registerWorkFlow).Methods("GET")
}

func startWorkflow(rw http.ResponseWriter, r *http.Request) {

	rw.Write([]byte("starting workflow\n"))
	// Tokenize url
	x := strings.Split(r.URL.Path, "/")
	switch x[2] {
	case "BankTransfer":
		// Set ID of workflow to run
		options1 := client.StartWorkflowOptions{
			TaskQueue:           BankTaskQueue,
			WorkflowTaskTimeout: (time.Duration(100) * time.Second),
			RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 0},
		}
		we, err := temporalClient.ExecuteWorkflow(context.Background(), options1, x[2], "Bank1", "Bank2", 5)
		if err != nil {
			log.Fatalln("An Error occurred when executing the workflow: ", err)
		}
		rw.Write([]byte("\nwe.GetRunID(): "))
		rw.Write([]byte(we.GetRunID()))
	case "longRunning":
		// Set options of workflow to run
		options := client.StartWorkflowOptions{
			TaskQueue:           LongTaskQueue,
			WorkflowTaskTimeout: (time.Duration(100) * time.Second),
			RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 0},
		}
		we, err := temporalClient.ExecuteWorkflow(context.Background(), options, x[2])
		if err != nil {
			log.Fatalln("An Error occurred when executing the workflow: ", err)
		}
		rw.Write([]byte("\nwe.GetRunID(): "))
		rw.Write([]byte(we.GetRunID()))
		rw.Write([]byte("\nwe.GetID(): "))
		rw.Write([]byte(we.GetID()))
	default:
	}

}

func terminateWorkFlow(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("test2\n"))
	x := strings.Split(r.URL.Path, "/")
	rw.Write([]byte(r.URL.Path))
	rw.Write([]byte("\n"))
	rw.Write([]byte(x[3]))

	err := temporalClient.CancelWorkflow(context.Background(), x[3], x[4])
	if err != nil {
		fmt.Println("ERROR during terminate: ", err.Error())
	}
}

func registerWorkFlow(rw http.ResponseWriter, r *http.Request) {

}

func getWorkflowStats(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("Get Workflow Stats\n"))
	// temporalClient.QueryWorkflow()
}
