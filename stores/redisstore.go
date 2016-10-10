package stores

import {
	"github.com/nats-io/go-nats-streaming/pb"
	"time"
}

type RedisStore struct {
	genericStore // embedded type, or anonymous field. "is-a" relationship
}