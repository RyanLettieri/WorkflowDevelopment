package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// Placeholder string for the task queue
const BankTaskQueue = "BankTaskQueue"

// Placeholder string for the task queue
const LongTaskQueue = "LongTaskQueue"

var temporalClient client.Client

type WorkflowStruct struct {
	WorkflowId    string `json:"workflow_id"`
	WorkflowRunId string `json:"workflow_run_id"`
}

type StateResponse struct {
	WfInfo    WorkflowStruct
	StartTime string                        `json:"start_time"`
	TaskQueue string                        `json:"task_queue"`
	Status    enums.WorkflowExecutionStatus `json:"status"`
}

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
	w.RegisterWorkflowWithOptions(BankTransfer, workflow.RegisterOptions{Name: "BankTransferWF"})
	w.RegisterActivity(GetBankBalance)
	w.RegisterActivityWithOptions(WriteFunc, activity.RegisterOptions{Name: "writeFunc"})
	err = w.Start()
	if err != nil {
		return
	}

	w2 := worker.New(temporalClient, LongTaskQueue, worker.Options{Identity: "workerTwo"})
	// Register a workflow and activities for the worker
	w2.RegisterWorkflowWithOptions(LongRunning, workflow.RegisterOptions{Name: "LongRunningWF"})
	w2.RegisterActivityWithOptions(LongRunningAct, activity.RegisterOptions{Name: "LongRunningAct"})
	err = w2.Start()
	if err != nil {
		return
	}
	log.Fatal(srv.ListenAndServe())
}

func setHandlers(r *mux.Router) {
	r.HandleFunc("/workflows/{workflowID}/{instanceId}/start", startWorkflow).Methods("POST")
	r.HandleFunc("/workflows/{workflowID}/{instanceId}/terminate", terminateWorkFlow).Methods("POST", "DELETE")
	r.HandleFunc("/workflows/{workflowID}/{instanceId}/register", registerWorkFlow).Methods("POST")
	r.HandleFunc("/workflows/{workflowID}/{instanceId}/info", getWorkflowStats).Methods("GET")
}

func startWorkflow(rw http.ResponseWriter, r *http.Request) {

	rw.Write([]byte("starting workflow\n"))
	// Tokenize url
	urlVars := strings.Split(r.URL.Path, "/")
	// Set generic options
	options := client.StartWorkflowOptions{
		TaskQueue:           BankTaskQueue,
		WorkflowTaskTimeout: (time.Duration(100) * time.Second),
	}
	switch urlVars[2] {
	case "BankTransfer":
		we, err := temporalClient.ExecuteWorkflow(context.Background(), options, urlVars[2], "Bank1", "Bank2", 5)
		if err != nil {
			log.Fatalln("An Error occurred when executing the workflow: ", err)
		}
		rw.Write([]byte("\nwe.GetRunID(): "))
		rw.Write([]byte(we.GetRunID()))
		rw.Write([]byte("\nwe.GetID(): "))
		rw.Write([]byte(we.GetID()))
	case "LongRunning":
		// Set options of workflow to run
		options.TaskQueue = LongTaskQueue
		we, err := temporalClient.ExecuteWorkflow(context.Background(), options, urlVars[2])
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
	rw.Write([]byte("Terminating Following workflow:\n"))
	urlVars := strings.Split(r.URL.Path, "/")
	wf := WorkflowStruct{urlVars[2], urlVars[3]}
	err := temporalClient.CancelWorkflow(context.Background(), wf.WorkflowId, wf.WorkflowRunId)
	if err != nil {
		fmt.Println("ERROR during terminate: ", err.Error())
	}
	bn, err := json.Marshal(wf)
	if err != nil {
		fmt.Println("ERROR during marshall: ", err.Error())
	}
	rw.Write(bn)
}

func registerWorkFlow(rw http.ResponseWriter, r *http.Request) {

}

func getWorkflowStats(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("Getting Workflow Stats\n"))
	urlVars := strings.Split(r.URL.Path, "/")
	resp, err := temporalClient.DescribeWorkflowExecution(context.Background(), urlVars[2], urlVars[3])
	if err != nil {
		fmt.Println("ERROR during get: ", err.Error())
	}
	// Build the output struct
	timeSt := resp.WorkflowExecutionInfo.StartTime.String()
	taskQ := resp.WorkflowExecutionInfo.GetTaskQueue()
	status := resp.WorkflowExecutionInfo.Status
	outputStruct := StateResponse{
		WfInfo:    WorkflowStruct{urlVars[2], urlVars[3]},
		StartTime: timeSt,
		TaskQueue: taskQ,
		Status:    status,
	}
	bn, err := json.Marshal(outputStruct)
	if err != nil {
		fmt.Println("ERROR during marshall: ", err.Error())
	}
	rw.Write(bn)
}
