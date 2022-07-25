package main

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

func BankTransfer(ctx workflow.Context, sourceAcc string, destinationAcc string, ammount float32) error {
	fmt.Printf("Starting Transfer of %f from %s to %s\n", ammount, sourceAcc, destinationAcc)
	options := workflow.ActivityOptions{
		TaskQueue:              BankTaskQueue,
		ScheduleToCloseTimeout: time.Second * 60,
		ScheduleToStartTimeout: time.Second * 60,
		StartToCloseTimeout:    time.Second * 60,
		HeartbeatTimeout:       time.Second * 60,
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

func LongRunning(ctx workflow.Context) error {
	fmt.Println("Starting Long Running Activity")

	options := workflow.ActivityOptions{
		TaskQueue:              LongTaskQueue,
		ScheduleToCloseTimeout: time.Second * 600,
		ScheduleToStartTimeout: time.Second * 600,
		StartToCloseTimeout:    time.Second * 600,
		HeartbeatTimeout:       time.Second * 5,
		WaitForCancellation:    true,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	err := workflow.ExecuteActivity(ctx, LongRunningAct).Get(ctx, nil)
	if err != nil {
		fmt.Println("ERROR During long run call: ", err.Error())
	}

	fmt.Println("Long Running Activity Finished")
	return nil
}
