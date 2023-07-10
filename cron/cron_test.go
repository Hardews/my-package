package cron

import (
	"fmt"
	"testing"
)

func TestEngine_AddFunc(t *testing.T) {
	e := NewEngine()

	e.AddFunc("test1", "*/5 * * * * *", func(a ...interface{}) {
		fmt.Println("test1 successful")
	})

	e.AddFunc("test2", "5,15,25 * * * * *", func(a ...interface{}) {
		fmt.Println("test2 successful")
	})

	e.AddFunc("test3", "10 * * * * *", func(a ...interface{}) {
		fmt.Println("test3 successful")
	})

	e.Start()

	select {}
}
