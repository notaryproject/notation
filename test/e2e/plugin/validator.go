package main

import (
	"fmt"
	"reflect"
)

func fieldSet(name ...string) map[string]struct{} {
	fs := make(map[string]struct{})
	for _, n := range name {
		fs[n] = struct{}{}
	}
	return fs
}

func validateRequiredField(data any, fieldSet map[string]struct{}) error {
	valueOfData := reflect.ValueOf(data)
	typeOfData := valueOfData.Type()
	for i := 0; i < valueOfData.NumField(); i++ {
		filedName := typeOfData.Field(i).Name
		if _, ok := fieldSet[filedName]; ok {
			// check the required filed
			if valueOfData.Field(i).IsZero() {
				return fmt.Errorf("%s is not set", filedName)
			}
		}
	}
	return nil
}
