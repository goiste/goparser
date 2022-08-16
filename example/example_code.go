package main

import (
	"context"
)

const floatConst = 0.0

var (
	// parser
	boolValue = true

	// parser:str
	stringValue = "3"

	// parser
	intValue = 3

	notParsedInValue = 17

	// parser
	float64Value = 3.14

	// parser
	float32Value = float32(3.14) // not implemented yet

	// parser
	notParsedFloatValue = floatConst // not implemented yet

	// parser
	float64SliceValue = []float64{3.14, 0.42}

	// parser:str
	stringSliceValue = []string{"a", "b", "c"}

	// parser:str
	stringToStringMapValue = map[string]string{"a": "1", "b": "2"}

	// parser
	pointerMapValue = &map[string]string{} // not implemented yet

	// parser
	intToFloat64MapValue = map[int]float64{3: 3.14, 17: 42.0}

	_, _, _, _, _, _, _, _, _, _, _ = stringValue, intValue, notParsedInValue, float64Value, float32Value,
		notParsedFloatValue, float64SliceValue, stringSliceValue, stringToStringMapValue, pointerMapValue,
		intToFloat64MapValue
)

type LocalStruct struct{}

func (s *LocalStruct) usefulFunc1(ctx *context.Context, str string) bool {
	_, _ = ctx, str
	if boolValue {
		boolValue = !boolValue
	}
	return boolValue
}

func (s LocalStruct) usefulFunc2(ctx *context.Context, str string) {
	_, _ = ctx, str
}

func (s LocalStruct) usefulFunc3(ctx context.Context, str string) {
	usefulFunc4(ctx, &str)
}

func usefulFunc4(ctx context.Context, str *string) {
	_, _ = ctx, str
	_ = LocalStruct{}
}
