package goparser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
)

// represents integer types
type iInt interface {
	uint8 | uint16 | uint32 | uint64 | int8 | int16 | int32 | int64
}

// represents float types
type iFloat interface {
	float32 | float64
}

// represents basic literal types
type iLit interface {
	string | iInt | iFloat | bool
}

// LitValue contains basic literal value
type LitValue[V iLit] struct {
	Doc   string
	Name  string
	Value V
}

// SliceLitValue contains a slice of basic literal values
type SliceLitValue[V iLit] struct {
	Doc   string
	Name  string
	Value []V
}

// MapLitValue contains a map with basic literal values as keys and values
type MapLitValue[K, V iLit] struct {
	Doc   string
	Name  string
	Value map[K]V
}

// LitVal represents a basic response type for walk callback function
type LitVal[K, V iLit] interface {
	LitValue[V] | SliceLitValue[V] | MapLitValue[K, V]
}

// GoParser contains an instance of ast.File
type GoParser struct {
	f *ast.File
}

// New returns a new instance of GoParser
func New(path string) (*GoParser, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return nil, fmt.Errorf("%q is a directory", path)
	}

	fileAst, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return &GoParser{f: fileAst}, nil
}

// GetBasicValues returns a list of values containing literal values by godoc label
//
//    // someLabel
//    var testVar = "3"
func GetBasicValues[V iLit](g *GoParser, docLabels ...string) []LitValue[V] {
	if len(docLabels) == 0 {
		return nil
	}

	docMap := make(map[string]struct{}, len(docLabels))
	for _, doc := range docLabels {
		docMap[doc] = struct{}{}
	}

	return getBasicValues[V](g.f, docMap)
}

func getBasicValues[V iLit](f *ast.File, docMap map[string]struct{}) []LitValue[V] {
	return walkDecls[int64, V, LitValue[V]](f, docMap, func(doc, name string, val ast.Expr) *LitValue[V] {
		var tVal V
		_, isBool := (interface{})(tVal).(bool)

		switch v := val.(type) {
		case *ast.Ident:
			if !isBool {
				return nil
			}

			b, ok := parseBool(v)
			if !ok {
				return nil
			}

			tVal = (interface{})(b).(V)
		case *ast.BasicLit:
			b := parseBasicLit[V](v)
			if b == nil {
				return nil
			}

			tVal = *b
		default:
			return nil
		}

		lVal := &LitValue[V]{
			Doc:   doc,
			Name:  name,
			Value: tVal,
		}

		return lVal
	})
}

// GetSliceValues returns a list of values containing slices of literal values by godoc label
//
//    // someLabel
//    var testVar = []string{"3"}
func GetSliceValues[V iLit](g *GoParser, docLabels ...string) []SliceLitValue[V] {
	if len(docLabels) == 0 {
		return nil
	}

	docMap := make(map[string]struct{}, len(docLabels))
	for _, doc := range docLabels {
		docMap[doc] = struct{}{}
	}

	return getSliceValues[V](g.f, docMap)
}

func getSliceValues[V iLit](f *ast.File, docMap map[string]struct{}) []SliceLitValue[V] {
	return walkDecls[int64, V, SliceLitValue[V]](f, docMap, func(doc, name string, val ast.Expr) *SliceLitValue[V] {
		cmpVal, ok := val.(*ast.CompositeLit)
		if !ok {
			return nil
		}

		sValues := make([]V, 0, len(cmpVal.Elts))
		for _, elt := range cmpVal.Elts {
			bVal, ok := elt.(*ast.BasicLit)
			if !ok {
				continue
			}

			pVal := parseBasicLit[V](bVal)
			if pVal == nil {
				return nil
			}

			if pVal == nil {
				continue
			}

			sValues = append(sValues, *pVal)
		}

		if len(sValues) > 0 {
			return &SliceLitValue[V]{
				Doc:   doc,
				Name:  name,
				Value: sValues,
			}
		}

		return nil
	})
}

// GetMapValues returns a list of values containing maps with literal types as keys and values by godoc label
//
//    // someLabel
//    var testVar = map[int]string{3: "3"}
func GetMapValues[K, V iLit](g *GoParser, docLabels ...string) []MapLitValue[K, V] {
	if len(docLabels) == 0 {
		return nil
	}

	docMap := make(map[string]struct{}, len(docLabels))
	for _, doc := range docLabels {
		docMap[doc] = struct{}{}
	}

	return getMapValues[K, V](g.f, docMap)
}

func getMapValues[K, V iLit](f *ast.File, docMap map[string]struct{}) []MapLitValue[K, V] {
	return walkDecls[K, V, MapLitValue[K, V]](f, docMap, func(doc, name string, val ast.Expr) *MapLitValue[K, V] {
		cmpVal, ok := val.(*ast.CompositeLit)
		if !ok {
			return nil
		}

		cValues := make(map[K]V, len(cmpVal.Elts))
		for _, elt := range cmpVal.Elts {
			cVal, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			keyVal := cVal.Key
			valVal := cVal.Value

			bKey, keyOk := keyVal.(*ast.BasicLit)
			bVal, valOk := valVal.(*ast.BasicLit)
			if !keyOk || !valOk {
				continue
			}

			k := parseBasicLit[K](bKey)
			v := parseBasicLit[V](bVal)
			if k == nil || v == nil {
				continue
			}

			cValues[*k] = *v
		}

		if len(cValues) > 0 {
			return &MapLitValue[K, V]{
				Doc:   doc,
				Name:  name,
				Value: cValues,
			}
		}

		return nil
	})
}

func walkDecls[K, V iLit, T LitVal[K, V]](f *ast.File, docMap map[string]struct{}, fn func(doc, name string, val ast.Expr) *T) []T {
	result := make([]T, 0)

	for _, d := range f.Decls {
		switch decl := d.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch s := spec.(type) {
				case *ast.ValueSpec:
					for _, n := range s.Names {
						if n.Obj == nil {
							continue
						}

						vSpec, ok := n.Obj.Decl.(*ast.ValueSpec)
						if !ok {
							continue
						}

						if vSpec.Doc == nil || len(vSpec.Doc.List) < 1 {
							continue
						}

						var foundDoc string
						for _, doc := range vSpec.Doc.List {
							docTxt := strings.TrimLeft(doc.Text, "/ ")
							if _, ok := docMap[docTxt]; ok {
								foundDoc = docTxt
								break
							}
						}

						if foundDoc == "" {
							continue
						}

						val := vSpec.Values[0]

						res := fn(foundDoc, n.Name, val)
						if res != nil {
							result = append(result, *res)
						}
					}
				}
			}
		}
	}

	return result
}

// GetFuncNames returns a list of function names by receiver type or param types
func GetFuncNames(g *GoParser, recType string, paramTypes ...string) []string {
	result := make([]string, 0)

outer:
	for _, d := range g.f.Decls {
		switch decl := d.(type) {
		case *ast.FuncDecl:
			rec := decl.Recv
			if rec == nil && recType != "" {
				continue
			}

			if rec != nil {
				r := rec.List[0]
				switch rType := r.Type.(type) {
				case *ast.Ident:
					if rType.Name != recType {
						continue
					}
				case *ast.StarExpr:
					id, ok := rType.X.(*ast.Ident)
					if !ok || id.Name != recType {
						continue
					}
				}
			}

			t := decl.Type
			if (t == nil || t.Params == nil) && len(paramTypes) > 0 {
				continue
			}

			if t != nil && t.Params != nil {
				paramsMap := make(map[string]struct{}, len(t.Params.List))

				for _, par := range t.Params.List {
					switch pType := par.Type.(type) {
					case *ast.Ident:
						paramsMap[pType.Name] = struct{}{}
					case *ast.StarExpr:
						switch sType := pType.X.(type) {
						case *ast.Ident:
							paramsMap[sType.Name] = struct{}{}
						case *ast.SelectorExpr:
							if sType.Sel == nil {
								continue
							}
							paramsMap[sType.Sel.Name] = struct{}{}
						default:
							continue
						}
					case *ast.SelectorExpr:
						if pType.Sel == nil {
							continue
						}
						paramsMap[pType.Sel.Name] = struct{}{}
					}
				}

				for _, par := range paramTypes {
					_, ok := paramsMap[par]
					_, okQt := paramsMap[fmt.Sprintf("%q", par)]
					if !ok && !okQt {
						continue outer
					}
				}
			}

			result = append(result, decl.Name.Name)
		}
	}

	return result
}

func parseBool(val *ast.Ident) (result bool, ok bool) {
	b, err := strconv.ParseBool(val.Name)
	if err != nil {
		return false, false
	}
	return b, true
}

func parseBasicLit[V iLit](val *ast.BasicLit) *V {
	var zeroVal V
	switch (interface{})(zeroVal).(type) {
	case string:
		if val.Kind != token.STRING {
			return nil
		}

		strVal, err := strconv.Unquote(val.Value)
		if err != nil {
			return nil
		}

		zeroVal = (interface{})(strVal).(V)
	case int8:
		iv := parseIntLit[int8](val)
		if iv == nil {
			return nil
		}
		zeroVal = (interface{})(*iv).(V)
	case int16:
		iv := parseIntLit[int16](val)
		if iv == nil {
			return nil
		}
		zeroVal = (interface{})(*iv).(V)
	case int32:
		iv := parseIntLit[int32](val)
		if iv == nil {
			return nil
		}
		zeroVal = (interface{})(*iv).(V)
	case int64:
		iv := parseIntLit[int64](val)
		if iv == nil {
			return nil
		}
		zeroVal = (interface{})(*iv).(V)
	case uint8:
		iv := parseIntLit[uint8](val)
		if iv == nil {
			return nil
		}
		zeroVal = (interface{})(*iv).(V)
	case uint16:
		iv := parseIntLit[uint16](val)
		if iv == nil {
			return nil
		}
		zeroVal = (interface{})(*iv).(V)
	case uint32:
		iv := parseIntLit[uint32](val)
		if iv == nil {
			return nil
		}
		zeroVal = (interface{})(*iv).(V)
	case uint64:
		iv := parseIntLit[uint64](val)
		if iv == nil {
			return nil
		}
		zeroVal = (interface{})(*iv).(V)
	case float32:
		fv := parseFloatLit[float32](val)
		if fv == nil {
			return nil
		}
		zeroVal = (interface{})(*fv).(V)
	case float64:
		fv := parseFloatLit[float64](val)
		if fv == nil {
			return nil
		}
		zeroVal = (interface{})(*fv).(V)
	default:
		return nil
	}

	return &zeroVal
}

func parseIntLit[I iInt](val *ast.BasicLit) *I {
	if val.Kind != token.INT {
		return nil
	}

	var parsed I

	parseI := func() bool {
		v := parseInt[I](val.Value)
		if v == nil {
			return false
		}
		parsed = *v
		return true
	}

	parseU := func() bool {
		v := parseUint[I](val.Value)
		if v == nil {
			return false
		}
		parsed = *v
		return true
	}

	switch (interface{})(parsed).(type) {
	case int8:
		if !parseI() {
			return nil
		}
	case int16:
		if !parseI() {
			return nil
		}
	case int32:
		if !parseI() {
			return nil
		}
	case int64:
		if !parseI() {
			return nil
		}
	case uint8:
		if !parseU() {
			return nil
		}
	case uint16:
		if !parseU() {
			return nil
		}
	case uint32:
		if !parseU() {
			return nil
		}
	case uint64:
		if !parseU() {
			return nil
		}
	default:
		return nil
	}

	return &parsed
}

func parseUint[I iInt](s string) *I {
	var uintVal uint64
	var err error

	parse := func(bitSize int) {
		uintVal, err = strconv.ParseUint(s, 10, bitSize)
	}

	var zeroVal I
	switch (interface{})(zeroVal).(type) {
	case uint8:
		parse(8)
	case uint16:
		parse(16)
	case uint32:
		parse(32)
	case uint64:
		parse(64)
	default:
		return nil
	}

	if err != nil {
		return nil
	}

	zeroVal = (interface{})(uintVal).(I)

	return &zeroVal
}

func parseInt[I iInt](s string) *I {
	var intVal int64
	var err error

	parse := func(bitSize int) {
		intVal, err = strconv.ParseInt(s, 10, bitSize)
	}

	var zeroVal I
	switch (interface{})(zeroVal).(type) {
	case int8:
		parse(8)
	case int16:
		parse(16)
	case int32:
		parse(32)
	case int64:
		parse(64)
	default:
		return nil
	}

	if err != nil {
		return nil
	}

	zeroVal = (interface{})(intVal).(I)

	return &zeroVal
}

func parseFloatLit[F iFloat](val *ast.BasicLit) *F {
	if val.Kind != token.FLOAT {
		return nil
	}

	var parsed F

	parse := func() bool {
		v := parseFloat[F](val.Value)
		if v == nil {
			return false
		}
		parsed = *v
		return true
	}

	switch (interface{})(parsed).(type) {
	case float32:
		if !parse() {
			return nil
		}
	case float64:
		if !parse() {
			return nil
		}
	default:
		return nil
	}

	return &parsed
}

func parseFloat[F iFloat](s string) *F {
	var fVal float64
	var err error

	parse := func(bitSize int) {
		fVal, err = strconv.ParseFloat(s, bitSize)
	}

	var zeroVal F
	switch (interface{})(zeroVal).(type) {
	case float32:
		parse(32)
	case float64:
		parse(64)
	default:
		return nil
	}

	if err != nil {
		return nil
	}

	zeroVal = (interface{})(fVal).(F)

	return &zeroVal
}
