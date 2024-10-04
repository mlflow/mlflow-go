package main

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/mlflow/mlflow-go/pkg/config"
	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/utils"
)

type instanceMap[T any] struct {
	counter   int64
	mutex     sync.Mutex
	instances map[int64]T
}

func newInstanceMap[T any]() *instanceMap[T] {
	return &instanceMap[T]{
		instances: make(map[int64]T),
	}
}

//nolint:ireturn
func (s *instanceMap[T]) Get(id int64) (T, *contract.Error) {
	instance, ok := s.instances[id]
	if !ok {
		return instance, contract.NewError(
			protos.ErrorCode_RESOURCE_DOES_NOT_EXIST,
			"Instance not found",
		)
	}

	return instance, nil
}

func (s *instanceMap[T]) Create(
	creator func(ctx context.Context, cfg *config.Config) (T, error),
	configBytes []byte,
) int64 {
	cfg, err := config.NewConfigFromBytes(configBytes)
	if err != nil {
		logrus.Error("Failed to read config: ", err)

		return -1
	}

	logger := utils.NewLoggerFromConfig(cfg)

	logger.Debugf("Loaded config: %#v", cfg)

	instance, err := creator(
		utils.NewContextWithLogger(context.Background(), logger),
		cfg,
	)
	if err != nil {
		logger.Error("Failed to create instance: ", err)

		return -1
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.counter++
	s.instances[s.counter] = instance

	return s.counter
}

func (s *instanceMap[T]) Destroy(id int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.instances, id)
}
