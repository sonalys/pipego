package retry_test

import (
	"testing"
	"time"

	"github.com/sonalys/pipego/retry"
	"github.com/stretchr/testify/require"
)

func Test_ConstantRetry(t *testing.T) {
	cr := retry.Constant(time.Second)
	for i := 1; i <= 10; i++ {
		require.Equal(t, cr.Retry(i), time.Second)
	}
}

func Test_LinearRetry(t *testing.T) {
	lr := retry.Linear(time.Second)
	for i := 1; i <= 10; i++ {
		require.Equal(t, lr.Retry(i), time.Duration(i)*time.Second)
	}
}

func Test_ExpRetry(t *testing.T) {
	er := retry.Exp(time.Second, 10*time.Second, 2)
	expSlice := []time.Duration{
		1 * time.Second,  // 0 ^ 2 + 1 = 1s
		2 * time.Second,  // 1 ^ 2 + 1 = 2s
		5 * time.Second,  // 2 ˆ 2 + 1 = 5s
		10 * time.Second, // 3 ^ 2 + 1 = 10s
		10 * time.Second, // maxDelay
	}
	for i := range expSlice {
		require.Equal(t, expSlice[i], er.Retry(i))
	}
}
