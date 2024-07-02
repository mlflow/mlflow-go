package main

import (
	"encoding/json"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/mlflow/mlflow-go/pkg/config"
	"github.com/mlflow/mlflow-go/pkg/contract"
	"github.com/mlflow/mlflow-go/pkg/protos"
)

type serviceMap[T any] struct {
	counter  int64
	mutex    sync.Mutex
	services map[int64]T
}

func newServiceMap[T any]() *serviceMap[T] {
	return &serviceMap[T]{
		services: make(map[int64]T),
	}
}

//nolint:ireturn
func (s *serviceMap[T]) Get(id int64) (T, *contract.Error) {
	service, ok := s.services[id]
	if !ok {
		return service, contract.NewError(
			protos.ErrorCode_RESOURCE_DOES_NOT_EXIST,
			"Service not found",
		)
	}

	return service, nil
}

func (s *serviceMap[T]) Create(
	creator func(logger *logrus.Logger, config *config.Config) (*T, error),
	configBytes []byte,
) int64 {
	var config *config.Config
	if err := json.Unmarshal(configBytes, &config); err != nil {
		logrus.Error(err)

		return -1
	}

	logger := logrus.New()

	logLevel, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		logrus.Error(err)

		return -1
	}

	logger.SetLevel(logLevel)

	logger.Debugf("Loaded config: %#v", config)

	service, err := creator(logger, config)
	if err != nil {
		logger.Error("Failed to create service: ", err)

		return -1
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.counter++
	s.services[s.counter] = *service

	return s.counter
}

func (s *serviceMap[T]) Destroy(id int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.services, id)
}
