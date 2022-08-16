package main

import (
	"fmt"

	gp "github.com/goiste/goparser"
)

func main() {
	p, err := gp.New("example_code.go")
	if err != nil {
		panic(err)
	}

	boolValues := gp.GetBasicValues[bool](p, "parser")
	for _, v := range boolValues {
		fmt.Printf("name: %s; value: %t\n", v.Name, v.Value) // name: boolValue; value: true
	}

	stringValues := gp.GetBasicValues[string](p, "parser", "parser:str")
	for _, v := range stringValues {
		fmt.Printf("name: %s; value: %q\n", v.Name, v.Value) // name: stringValue; value: "3"
	}

	floatValues := gp.GetBasicValues[float64](p, "parser")
	for _, v := range floatValues {
		fmt.Printf("name: %s; value: %.02f\n", v.Name, v.Value) // name: float64Value; value: 3.14
	}

	floatSliceValues := gp.GetSliceValues[float64](p, "parser")
	for _, v := range floatSliceValues {
		fmt.Printf("name: %s; values: %v\n", v.Name, v.Value) // name: float64SliceValue; values: [3.14 0.42]
	}

	stringSliceValues := gp.GetSliceValues[string](p, "parser", "parser:str")
	for _, v := range stringSliceValues {
		fmt.Printf("name: %s; values: %v\n", v.Name, v.Value) // name: stringSliceValue; values: [a b c]
	}

	stringMapValues := gp.GetMapValues[string, string](p, "parser", "parser:str")
	for _, v := range stringMapValues {
		fmt.Printf("name: %s; values: %+v\n", v.Name, v.Value) // name: stringToStringMapValue; values: map[a:1 b:2]
	}

	floatMapValues := gp.GetMapValues[int64, float64](p, "parser")
	for _, v := range floatMapValues {
		fmt.Printf("name: %s; values: %+v\n", v.Name, v.Value) // name: intToFloat64MapValue; values: map[3:3.14 17:42]
	}

	nms := gp.GetFuncNames(p, "LocalStruct", "Context", "string")
	fmt.Printf("%v\n", nms) // [usefulFunc1 usefulFunc2 usefulFunc3]

	nms = gp.GetFuncNames(p, "LocalStruct", "Context")
	fmt.Printf("%v\n", nms) // [usefulFunc1 usefulFunc2 usefulFunc3]

	nms = gp.GetFuncNames(p, "", "Context")
	fmt.Printf("%v\n", nms) // [usefulFunc4]
}
