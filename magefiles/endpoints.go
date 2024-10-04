//go:build mage

//nolint:wrapcheck
package main

import (
	"os"

	"github.com/olekukonko/tablewriter"

	"github.com/mlflow/mlflow-go/magefiles/generate"
	"github.com/mlflow/mlflow-go/magefiles/generate/discovery"
)

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}

	return false
}

// Print an overview of implementated API endpoints.
func Endpoints() error {
	services, err := discovery.GetServiceInfos()
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Service", "Endpoint", "Implemented"})
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER})
	table.SetRowLine(true)

	for _, service := range services {
		servinceInfo, ok := generate.ServiceInfoMap[service.Name]
		if !ok {
			continue
		}

		for _, method := range service.Methods {
			implemented := "No"
			if contains(servinceInfo.ImplementedEndpoints, method.Name) {
				implemented = "Yes"
			}

			table.Append([]string{service.Name, method.Name, implemented})
		}
	}

	table.Render()

	return nil
}
