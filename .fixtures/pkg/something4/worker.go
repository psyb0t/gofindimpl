package something4

import "fmt"

type BackgroundWorker struct {
	taskQueue []string
}

func (b *BackgroundWorker) Process() error {
	fmt.Println("BackgroundWorker processing tasks")
	return nil
}

func (b *BackgroundWorker) AddTask(task string) {
	b.taskQueue = append(b.taskQueue, task)
}

func (b *BackgroundWorker) GetTaskCount() int {
	return len(b.taskQueue)
}