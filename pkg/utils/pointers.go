package utils

import (
	"strconv"
)

func PtrTo[T any](v T) *T {
	return &v
}

func ConvertInt32PointerToStringPointer(iPtr *int32) *string {
	if iPtr == nil {
		return nil
	}

	iValue := *iPtr
	sValue := strconv.Itoa(int(iValue))

	return &sValue
}

func ConvertStringPointerToInt32Pointer(s *string) int32 {
	if s == nil {
		return 0
	}

	iValue, err := strconv.ParseInt(*s, 10, 32)
	if err != nil {
		return 0
	}

	return int32(iValue)
}

func DumpStringPointer(s *string) string {
	if s == nil {
		return "<nil>"
	}

	return *s
}
