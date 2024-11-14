package broker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jamesTait-jt/goflow/task"
	"github.com/redis/go-redis/v9"
)

type redisClient interface {
	LPush(ctx context.Context, key string, values ...any) *redis.IntCmd
	BRPop(ctx context.Context, timeout time.Duration, keys ...string) *redis.StringSliceCmd
}

type Serialiser[T task.TaskOrResult] interface {
	Serialise(toSerialise T) ([]byte, error)
}

type Deserialiser[T task.TaskOrResult] interface {
	Deserialise(toDeserialise []byte) (T, error)
}

type RedisBroker[T task.TaskOrResult] struct {
	client        redisClient
	redisQueueKey string
	outChan       chan T
	started       sync.Once
	wg            *sync.WaitGroup
	serialiser    Serialiser[T]
	deserialiser  Deserialiser[T]
	opts          redisBrokerOptions
}

func NewRedisBroker[T task.TaskOrResult](
	client redisClient,
	key string,
	serialiser Serialiser[T],
	deserialiser Deserialiser[T],
	opt ...RedisBrokerOption,
) *RedisBroker[T] {
	opts := defaultRedisBrokerOptions()

	for _, o := range opt {
		o.apply(&opts)
	}

	return &RedisBroker[T]{
		client:        client,
		redisQueueKey: key,
		serialiser:    serialiser,
		deserialiser:  deserialiser,
		opts:          opts,
		outChan:       make(chan T),
		wg:            &sync.WaitGroup{},
	}
}

func (rb *RedisBroker[T]) Submit(ctx context.Context, submission T) error {
	serialised, err := rb.serialiser.Serialise(submission)
	if err != nil {
		return err
	}

	_, err = rb.client.LPush(ctx, rb.redisQueueKey, serialised).Result()
	if err != nil {
		return err
	}

	return nil
}

func (rb *RedisBroker[T]) Dequeue(ctx context.Context) <-chan T {
	rb.started.Do(func() {
		rb.wg.Add(1)
		go rb.pollRedis(ctx)
	})

	return rb.outChan
}

func (rb *RedisBroker[T]) pollRedis(ctx context.Context) {
	defer rb.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		default:
			redisResult, err := rb.client.BRPop(ctx, rb.opts.pollTimeout, rb.redisQueueKey).Result()
			if err != nil {
				if err == redis.Nil {
					// BRPop timed out
					continue
				}

				rb.opts.logger.Warn(fmt.Sprintf("BRPop error: %v", err.Error()))

				continue
			}

			result, err := rb.deserialiser.Deserialise([]byte(redisResult[1]))
			if err != nil {
				rb.opts.logger.Warn(fmt.Sprintf("Failed to deserialize task: %v", err))

				continue
			}

			rb.outChan <- result
		}
	}
}

func (rb *RedisBroker[T]) AwaitShutdown() {
	rb.wg.Wait()
}
