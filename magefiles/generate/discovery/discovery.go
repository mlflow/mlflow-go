package discovery

import (
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/mlflow/mlflow-go/pkg/protos"
	"github.com/mlflow/mlflow-go/pkg/protos/artifacts"
)

type ServiceInfo struct {
	Name    string
	Methods []MethodInfo
}

type MethodInfo struct {
	Name        string
	PackageName string
	Input       string
	Output      string
	Endpoints   []Endpoint
}

type Endpoint struct {
	Method string
	Path   string
}

var routeParameterRegex = regexp.MustCompile(`<[^>]+:([^>]+)>`)

// Get the safe path to use in Fiber registration.
func (e Endpoint) GetFiberPath() string {
	// e.Path cannot be trusted, it could be something like /mlflow-artifacts/artifacts/<path:artifact_path>
	// Which would need to converted to /mlflow-artifacts/artifacts/:path
	path := routeParameterRegex.ReplaceAllStringFunc(e.Path, func(s string) string {
		parts := strings.Split(s, ":")

		return ":" + strings.Trim(parts[0], "< ")
	})

	// or it could be something like /mlflow/traces/{request_id}/tags
	// which would need to be converted to /mlflow/traces/:request_id/tags
	path = strings.ReplaceAll(path, "{", ":")
	path = strings.ReplaceAll(path, "}", "")

	return path
}

func GetServiceInfos() ([]ServiceInfo, error) {
	serviceInfos := make([]ServiceInfo, 0)

	services := []struct {
		Name        string
		PackageName string
		Descriptor  protoreflect.FileDescriptor
	}{
		{"MlflowService", "protos", protos.File_service_proto},
		{"ModelRegistryService", "protos", protos.File_model_registry_proto},
		{"MlflowArtifactsService", "artifacts", artifacts.File_mlflow_artifacts_proto},
	}

	for _, service := range services {
		serviceDescriptor := service.Descriptor.Services().ByName(protoreflect.Name(service.Name))

		if serviceDescriptor == nil {
			//nolint:goerr113
			return nil, fmt.Errorf("service %s not found", service.Name)
		}

		serviceInfo := ServiceInfo{Name: service.Name, Methods: make([]MethodInfo, 0)}

		methods := serviceDescriptor.Methods()
		for mIdx := range methods.Len() {
			method := methods.Get(mIdx)
			options := method.Options()
			extension := proto.GetExtension(options, protos.E_Rpc)

			endpoints := make([]Endpoint, 0)
			rpcOptions, ok := extension.(*protos.DatabricksRpcOptions)

			if ok {
				for _, endpoint := range rpcOptions.GetEndpoints() {
					endpoints = append(endpoints, Endpoint{Method: endpoint.GetMethod(), Path: endpoint.GetPath()})
				}
			}

			output := fmt.Sprintf("%s_%s", string(method.Output().Parent().Name()), string(method.Output().Name()))
			methodInfo := MethodInfo{
				string(method.Name()), service.PackageName, string(method.Input().Name()), output, endpoints,
			}
			serviceInfo.Methods = append(serviceInfo.Methods, methodInfo)
		}

		serviceInfos = append(serviceInfos, serviceInfo)
	}

	return serviceInfos, nil
}
