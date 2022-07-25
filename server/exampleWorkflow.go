package main

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

func BankTransfer(ctx workflow.Context, sourceAcc string, destinationAcc string, ammount float32) error {
	fmt.Printf("STARTING TRANSFER of %f from %s to %s\n", ammount, sourceAcc, destinationAcc)

	options := workflow.ActivityOptions{
		TaskQueue:              BankTaskQueue,
		ScheduleToCloseTimeout: time.Second * 60,
		ScheduleToStartTimeout: time.Second * 60,
		StartToCloseTimeout:    time.Second * 60,
		HeartbeatTimeout:       time.Second * 41,
		WaitForCancellation:    false,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var balanceA int
	var balanceB int

	err := workflow.ExecuteActivity(ctx, GetBankBalance, sourceAcc).Get(ctx, &balanceA)
	if err != nil {
		fmt.Println("ERROR during balance call for source bank: ", err.Error())
	}

	err = workflow.ExecuteActivity(ctx, GetBankBalance, destinationAcc).Get(ctx, &balanceB)
	if err != nil {
		fmt.Println("ERROR during balance call for destination bank: ", err.Error())
	}

	err = workflow.ExecuteActivity(ctx, WriteFunc).Get(ctx, nil)
	if err != nil {
		fmt.Println("ERROR during write call: ", err.Error())
	}
	return nil
}

func longRunning(ctx workflow.Context) error {
	fmt.Println("RRL STARTING LONG RUN")
	options := workflow.ActivityOptions{
		TaskQueue:              LongTaskQueue,
		ScheduleToCloseTimeout: time.Second * 600,
		ScheduleToStartTimeout: time.Second * 600,
		StartToCloseTimeout:    time.Second * 600,
		HeartbeatTimeout:       time.Second * 600,
		WaitForCancellation:    false,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	// The reason there are 2 instances of executeActivity here:
	// The first call seems to work as intended, but the second call is what is ion the recommended guide on the temporal page. Howver, it doesn't seem to work
	err := workflow.ExecuteActivity(ctx, longRunningAct).Get(ctx, nil)
	if err != nil {
		fmt.Println("RRL ERROR During long run call")
		fmt.Println("RRL ERROR: ", err.Error())
	}

	future := workflow.ExecuteActivity(ctx, longRunningAct)
	if future.IsReady() {
		err := future.Get(ctx, nil)
		if err != nil {
			fmt.Println("ERROR during long run activity: ", err.Error())
		}
	}

	return nil
}
