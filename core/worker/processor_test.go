package worker

import (
	"fmt"
	"github.com/ananduee/raker/core/storage"
	"os"
	"testing"
)

var store, err = storage.OpenForUnitTest("processor__test")

func TestMain(m *testing.M) {
	if err != nil {
		fmt.Println("Failed to initialize storage", err)
		os.Exit(1)
	}
	exitValue := m.Run()
	store.CleanUpFoUnitTests()
	os.Exit(exitValue)
}

func TestNewTaskCanBeAddedAndExecuted(t *testing.T) {
	worker := New()
	job := &Job{}
	worker.AddPeriodicTask("TestNewTaskCanBeAddedAndExecuted", )
}


