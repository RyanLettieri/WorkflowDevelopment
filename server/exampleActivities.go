package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

func GetBankBalance(ctx context.Context, bankAccount string) (int, error) {
	fmt.Printf("Balance of %s is %d\n", bankAccount, 10)
	return 10, nil
}

func WriteFunc(ctx context.Context) error {
	fmt.Println("Calling Write Func")
	if err := os.WriteFile("file2.txt", []byte("Transfered money from bank1 to bank2 RRL2"), 0666); err != nil {
		log.Fatal(err)
	}
	return nil
}

func longRunningAct(ctx context.Context) error {
	var counter int
	for i := 0; i < 70; i++ {
		fmt.Println("Running Long Running Activity", counter)
		counter++
		time.Sleep(time.Duration(5) * time.Second)
	}
	return nil
}
