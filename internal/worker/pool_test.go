package worker

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewWorkerPool(t *testing.T) {
	t.Run("Creates a new worker pool with variables initialised", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		numWorkers := 5

		// Act
		wp := NewWorkerPool(ctx, numWorkers, nil)

		// Assert
		assert.Equal(t, numWorkers, len(wp.workers))
		assert.Equal(t, ctx, wp.ctx)
		assert.NotNil(t, wp.wg)
	})
}