package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/activity"
)

func GetBankBalance(ctx context.Context, bankAccount string) (int, error) {
	fmt.Printf("Balance of %s is %d\n", bankAccount, 10)
	return 10, nil
}

func WriteFunc(ctx context.Context) error {
	fmt.Println("Calling Write Func")
	if err := os.WriteFile("file2.txt", []byte("Transfered money from bank1 to bank2"), 0666); err != nil {
		log.Fatal(err)
	}
	return nil
}

func LongRunningAct(ctx context.Context) error {
	counter := 0
	for i := 0; i <= 60; i++ {

		select {
		case <-time.After(1 * time.Second):
			fmt.Println("Running Long Running Activity", counter)
			os.WriteFile("file2.txt", []byte{byte(counter)}, 0666)
			counter++
			activity.RecordHeartbeat(ctx, "")
		case <-ctx.Done():
			fmt.Println("Activity Cancelled")
			return nil
		}
	}
	return nil
}
