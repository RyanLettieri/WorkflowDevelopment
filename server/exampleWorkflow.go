package main

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

func BankTransfer(ctx workflow.Context, sourceAcc string, destinationAcc string, ammount float32) error {
	fmt.Printf("RRL STARTING TRANSFER of %f from %s to %s\n", ammount, sourceAcc, destinationAcc)

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
		fmt.Println("RRL ERROR During 1st call")
		fmt.Println("RRL ERROR: ", err.Error())
	}

	err = workflow.ExecuteActivity(ctx, GetBankBalance, destinationAcc).Get(ctx, &balanceB)
	if err != nil {
		fmt.Println("RRL ERROR During 2nd call")
		fmt.Println("RRL ERROR: ", err.Error())
	}

	err = workflow.ExecuteActivity(ctx, WriteFunc).Get(ctx, nil)
	if err != nil {
		fmt.Println("RRL ERROR During 3rd call")
		fmt.Println("RRL ERROR: ", err.Error())
	}
	fmt.Println("END OF WORKFLOW")
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

	err := workflow.ExecuteActivity(ctx, longRunningAct).Get(ctx, nil)
	if err != nil {
		fmt.Println("RRL ERROR During long run call")
		fmt.Println("RRL ERROR: ", err.Error())
	}

	future := workflow.ExecuteActivity(ctx, longRunningAct)
	// time.Sleep(time.Duration(1) * time.Second)
	if future.IsReady() {
		err := future.Get(ctx, nil)
		if err != nil {
			fmt.Println("RRL ERROR During long run call")
			fmt.Println("RRL ERROR: ", err.Error())
		} else {
			fmt.Println("RRL SUCCESS")
		}

	}
	fmt.Println("END OF WORKFLOW")

	return nil
}
