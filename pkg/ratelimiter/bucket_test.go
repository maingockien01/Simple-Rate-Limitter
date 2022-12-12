package ratelimiter_test

import (
	"testing"
	"time"

	"github.com/maingockien01/proxy/pkg/ratelimiter"
	"github.com/stretchr/testify/require"
)

var tokenBucket *ratelimiter.TokenBucket

func setup() {
	tokenBucket = ratelimiter.NewTokenBucket(5, 2, "key", "/whoami")
}

func teardown() {

}

func TestIsAccepted(t *testing.T) {
	setup()
	defer teardown()

	accept1 := tokenBucket.IsRequestAllowed(1)

	require.True(t, accept1)

	accept2 := tokenBucket.IsRequestAllowed(1)

	require.True(t, accept2)

	accept3 := tokenBucket.IsRequestAllowed(1)

	require.False(t, accept3)
	time.Sleep(1 * time.Second)

	accept4 := tokenBucket.IsRequestAllowed(1)

	require.True(t, accept4)

}
