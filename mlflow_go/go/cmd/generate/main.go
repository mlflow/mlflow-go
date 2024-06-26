package main

import (
	"fmt"
	"os"
)

const ExpectedNumberOfArguments = 2

//nolint:forbidigo
func main() {
	if len(os.Args) != ExpectedNumberOfArguments {
		fmt.Println("Usage: program <path to mlflow/go folder>")
		os.Exit(1)
	}

	if err := SyncProtos(); err != nil {
		fmt.Printf("Error downloading protos: %s\n", err)
		os.Exit(1)
	}

	pkgFolder := os.Args[1]
	if _, err := os.Stat(pkgFolder); os.IsNotExist(err) {
		fmt.Printf("The provided path does not exist: %s\n", pkgFolder)
		os.Exit(1)
	}

	if err := addQueryAnnotations(pkgFolder); err != nil {
		fmt.Printf("Error adding query annotations: %s\n", err)
		os.Exit(1)
	}

	if err := generateSourceCode(pkgFolder); err != nil {
		fmt.Printf("Error generating source code: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully added query annotations and generated services!")
}
