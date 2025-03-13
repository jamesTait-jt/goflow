package main

import (
	"fmt"
	"time"

	"github.com/jamesTait-jt/goflow"
	"github.com/jamesTait-jt/goflow/broker"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/jamesTait-jt/goflow/pkg/serialise"
	"github.com/jamesTait-jt/goflow/pkg/store"
	"github.com/jamesTait-jt/goflow/task"
	"github.com/redis/go-redis/v9"
)

func main() {
	fmt.Println("starting")

	redisClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	fmt.Println("conntected to redis")

	resultsStore := store.NewInMemoryKVStore[string, task.Result]()

	taskSerialiser := serialise.NewGobSerialiser[task.Task]()
	resultSerialiser := serialise.NewGobSerialiser[task.Result]()
	logger := log.NewConsoleLogger()

	gf := goflow.New(
		broker.NewRedisBroker(redisClient, "tasks", taskSerialiser, broker.WithLogger(logger)),
		broker.NewRedisBroker(redisClient, "results", resultSerialiser, broker.WithLogger(logger)),
		goflow.WithResultsStore(resultsStore),
	)

	gf.Start()

	maxItrs := 10
	results := make(chan task.Result, maxItrs)

	ids := make(chan string, maxItrs)
	go func() {
		for id := range ids {
			for {
				result, ok, _ := gf.GetResult(id)
				fmt.Println(id)
				if ok {
					fmt.Println("GOT RESULT")
					results <- result
					break
				}

				time.Sleep(time.Second)
			}
		}

		close(results)
	}()

	for i := 0; i < maxItrs; i++ {
		id, _ := gf.Push("doubler", fmt.Sprintf(`{"N": %d}`, i))
		fmt.Println("GOT ID BACK: ", id)
		ids <- id
	}

	close(ids)

	i := 0
	for r := range results {
		i++
		fmt.Println(i, r)
	}
}
