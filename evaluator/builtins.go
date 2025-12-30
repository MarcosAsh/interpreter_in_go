package evaluator

import (
	"fmt"
	"pearl/ast"
	"pearl/object"
	"regexp"
	"sort"
	"strings"
)

// EvalFn is set by init() to break the cycle
var EvalFn func(node ast.Node, env *object.Environment) object.Object

func init() {
	EvalFn = Eval
}

// helper functions for builtins
func unwrapReturn(obj object.Object) object.Object {
	if rv, ok := obj.(*object.ReturnValue); ok {
		return rv.Value
	}
	return obj
}

func isTruthyBuiltin(obj object.Object) bool {
	if obj == nil {
		return false
	}
	switch obj := obj.(type) {
	case *object.Null:
		return false
	case *object.Boolean:
		return obj.Value
	case *object.Integer:
		return obj.Value != 0
	case *object.String:
		return obj.Value != ""
	case *object.Array:
		return len(obj.Elements) > 0
	default:
		return true
	}
}

var builtins = map[string]*object.Builtin{
	"print": {
		Name: "print",
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Print(arg.Inspect())
			}
			fmt.Println()
			return NULL
		},
	},

	"type": {
		Name: "type",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("type() takes 1 argument, got %d", len(args))
			}
			return &object.String{Value: string(args[0].Type())}
		},
	},

	"len": {
		Name: "len",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("len() takes 1 argument, got %d", len(args))
			}
			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			case *object.Map:
				return &object.Integer{Value: int64(len(arg.Pairs))}
			default:
				return newError("len() not supported for %s", args[0].Type())
			}
		},
	},

	"upper": {
		Name: "upper",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("upper() takes 1 argument")
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("upper() requires a string")
			}
			return &object.String{Value: strings.ToUpper(s.Value)}
		},
	},

	"lower": {
		Name: "lower",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("lower() takes 1 argument")
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("lower() requires a string")
			}
			return &object.String{Value: strings.ToLower(s.Value)}
		},
	},

	"trim": {
		Name: "trim",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("trim() takes 1 argument")
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("trim() requires a string")
			}
			return &object.String{Value: strings.TrimSpace(s.Value)}
		},
	},

	"ltrim": {
		Name: "ltrim",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("ltrim() takes 1 argument")
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("ltrim() requires a string")
			}
			return &object.String{Value: strings.TrimLeft(s.Value, " \t\n\r")}
		},
	},

	"rtrim": {
		Name: "rtrim",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("rtrim() takes 1 argument")
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("rtrim() requires a string")
			}
			return &object.String{Value: strings.TrimRight(s.Value, " \t\n\r")}
		},
	},

	"split": {
		Name: "split",
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("split() takes 1-2 arguments")
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("split() requires a string")
			}
			sep := " "
			if len(args) == 2 {
				sepStr, ok := args[1].(*object.String)
				if !ok {
					return newError("split() separator must be a string")
				}
				sep = sepStr.Value
			}
			parts := strings.Split(s.Value, sep)
			elements := make([]object.Object, len(parts))
			for i, p := range parts {
				elements[i] = &object.String{Value: p}
			}
			return &object.Array{Elements: elements}
		},
	},

	"join": {
		Name: "join",
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("join() takes 1-2 arguments")
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("join() requires an array")
			}
			sep := ""
			if len(args) == 2 {
				sepStr, ok := args[1].(*object.String)
				if !ok {
					return newError("join() separator must be a string")
				}
				sep = sepStr.Value
			}
			parts := make([]string, len(arr.Elements))
			for i, el := range arr.Elements {
				parts[i] = el.Inspect()
			}
			return &object.String{Value: strings.Join(parts, sep)}
		},
	},

	"replace": {
		Name: "replace",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("replace() takes 3 arguments: string, old, new")
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("replace() first arg must be a string")
			}
			switch old := args[1].(type) {
			case *object.String:
				newStr, ok := args[2].(*object.String)
				if !ok {
					return newError("replace() new must be a string")
				}
				return &object.String{Value: strings.Replace(s.Value, old.Value, newStr.Value, 1)}
			case *object.Regex:
				newStr, ok := args[2].(*object.String)
				if !ok {
					return newError("replace() new must be a string")
				}
				return &object.String{Value: old.Regexp.ReplaceAllString(s.Value, newStr.Value)}
			default:
				return newError("replace() old must be a string or regex")
			}
		},
	},

	"replace_all": {
		Name: "replace_all",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("replace_all() takes 3 arguments")
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("replace_all() first arg must be a string")
			}
			old, ok := args[1].(*object.String)
			if !ok {
				return newError("replace_all() old must be a string")
			}
			newStr, ok := args[2].(*object.String)
			if !ok {
				return newError("replace_all() new must be a string")
			}
			return &object.String{Value: strings.ReplaceAll(s.Value, old.Value, newStr.Value)}
		},
	},

	"contains": {
		Name: "contains",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("contains() takes 2 arguments")
			}
			switch container := args[0].(type) {
			case *object.String:
				needle, ok := args[1].(*object.String)
				if !ok {
					return newError("contains() needle must be a string for string search")
				}
				return nativeBoolToBooleanObject(strings.Contains(container.Value, needle.Value))
			case *object.Array:
				for _, el := range container.Elements {
					if el.Inspect() == args[1].Inspect() {
						return TRUE
					}
				}
				return FALSE
			default:
				return newError("contains() requires string or array")
			}
		},
	},

	"starts_with": {
		Name: "starts_with",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("starts_with() takes 2 arguments")
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("starts_with() requires a string")
			}
			prefix, ok := args[1].(*object.String)
			if !ok {
				return newError("starts_with() prefix must be a string")
			}
			return nativeBoolToBooleanObject(strings.HasPrefix(s.Value, prefix.Value))
		},
	},

	"ends_with": {
		Name: "ends_with",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("ends_with() takes 2 arguments")
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("ends_with() requires a string")
			}
			suffix, ok := args[1].(*object.String)
			if !ok {
				return newError("ends_with() suffix must be a string")
			}
			return nativeBoolToBooleanObject(strings.HasSuffix(s.Value, suffix.Value))
		},
	},

	"substr": {
		Name: "substr",
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 2 || len(args) > 3 {
				return newError("substr() takes 2-3 arguments")
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("substr() requires a string")
			}
			start, ok := args[1].(*object.Integer)
			if !ok {
				return newError("substr() start must be an integer")
			}
			startIdx := int(start.Value)
			if startIdx < 0 {
				startIdx = len(s.Value) + startIdx
			}
			if startIdx < 0 {
				startIdx = 0
			}
			if startIdx >= len(s.Value) {
				return &object.String{Value: ""}
			}
			if len(args) == 2 {
				return &object.String{Value: s.Value[startIdx:]}
			}
			length, ok := args[2].(*object.Integer)
			if !ok {
				return newError("substr() length must be an integer")
			}
			endIdx := startIdx + int(length.Value)
			if endIdx > len(s.Value) {
				endIdx = len(s.Value)
			}
			return &object.String{Value: s.Value[startIdx:endIdx]}
		},
	},

	"repeat": {
		Name: "repeat",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("repeat() takes 2 arguments")
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("repeat() requires a string")
			}
			n, ok := args[1].(*object.Integer)
			if !ok {
				return newError("repeat() count must be an integer")
			}
			return &object.String{Value: strings.Repeat(s.Value, int(n.Value))}
		},
	},

	"reverse": {
		Name: "reverse",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("reverse() takes 1 argument")
			}
			switch arg := args[0].(type) {
			case *object.String:
				runes := []rune(arg.Value)
				for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
					runes[i], runes[j] = runes[j], runes[i]
				}
				return &object.String{Value: string(runes)}
			case *object.Array:
				newElements := make([]object.Object, len(arg.Elements))
				for i, j := 0, len(arg.Elements)-1; j >= 0; i, j = i+1, j-1 {
					newElements[i] = arg.Elements[j]
				}
				return &object.Array{Elements: newElements}
			default:
				return newError("reverse() requires string or array")
			}
		},
	},

	"lines": {
		Name: "lines",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("lines() takes 1 argument")
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("lines() requires a string")
			}
			parts := strings.Split(s.Value, "\n")
			elements := make([]object.Object, len(parts))
			for i, p := range parts {
				elements[i] = &object.String{Value: p}
			}
			return &object.Array{Elements: elements}
		},
	},

	"chars": {
		Name: "chars",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("chars() takes 1 argument")
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("chars() requires a string")
			}
			runes := []rune(s.Value)
			elements := make([]object.Object, len(runes))
			for i, r := range runes {
				elements[i] = &object.String{Value: string(r)}
			}
			return &object.Array{Elements: elements}
		},
	},

	"match": {
		Name: "match",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("match() takes 2 arguments: string, regex")
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("match() first arg must be a string")
			}
			re, ok := args[1].(*object.Regex)
			if !ok {
				return newError("match() second arg must be a regex")
			}
			matches := re.Regexp.FindStringSubmatch(s.Value)
			if matches == nil {
				return NULL
			}
			elements := make([]object.Object, len(matches))
			for i, m := range matches {
				elements[i] = &object.String{Value: m}
			}
			return &object.Array{Elements: elements}
		},
	},

	"match_all": {
		Name: "match_all",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("match_all() takes 2 arguments")
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("match_all() first arg must be a string")
			}
			re, ok := args[1].(*object.Regex)
			if !ok {
				return newError("match_all() second arg must be a regex")
			}
			allMatches := re.Regexp.FindAllStringSubmatch(s.Value, -1)
			results := make([]object.Object, len(allMatches))
			for i, matches := range allMatches {
				elements := make([]object.Object, len(matches))
				for j, m := range matches {
					elements[j] = &object.String{Value: m}
				}
				results[i] = &object.Array{Elements: elements}
			}
			return &object.Array{Elements: results}
		},
	},

	"regex": {
		Name: "regex",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("regex() takes 1 argument")
			}
			s, ok := args[0].(*object.String)
			if !ok {
				return newError("regex() requires a string pattern")
			}
			re, err := regexp.Compile(s.Value)
			if err != nil {
				return newError("invalid regex: %s", err)
			}
			return &object.Regex{Pattern: s.Value, Regexp: re}
		},
	},

	"push": {
		Name: "push",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("push() takes 2 arguments")
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("push() requires an array")
			}
			arr.Elements = append(arr.Elements, args[1])
			return arr
		},
	},

	"pop": {
		Name: "pop",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("pop() takes 1 argument")
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("pop() requires an array")
			}
			if len(arr.Elements) == 0 {
				return NULL
			}
			last := arr.Elements[len(arr.Elements)-1]
			arr.Elements = arr.Elements[:len(arr.Elements)-1]
			return last
		},
	},

	"shift": {
		Name: "shift",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("shift() takes 1 argument")
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("shift() requires an array")
			}
			if len(arr.Elements) == 0 {
				return NULL
			}
			first := arr.Elements[0]
			arr.Elements = arr.Elements[1:]
			return first
		},
	},

	"unshift": {
		Name: "unshift",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("unshift() takes 2 arguments")
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("unshift() requires an array")
			}
			arr.Elements = append([]object.Object{args[1]}, arr.Elements...)
			return arr
		},
	},

	"slice": {
		Name: "slice",
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 2 || len(args) > 3 {
				return newError("slice() takes 2-3 arguments")
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("slice() requires an array")
			}
			start, ok := args[1].(*object.Integer)
			if !ok {
				return newError("slice() start must be an integer")
			}
			startIdx := int(start.Value)
			if startIdx < 0 {
				startIdx = len(arr.Elements) + startIdx
			}
			endIdx := len(arr.Elements)
			if len(args) == 3 {
				end, ok := args[2].(*object.Integer)
				if !ok {
					return newError("slice() end must be an integer")
				}
				endIdx = int(end.Value)
				if endIdx < 0 {
					endIdx = len(arr.Elements) + endIdx
				}
			}
			if startIdx < 0 {
				startIdx = 0
			}
			if endIdx > len(arr.Elements) {
				endIdx = len(arr.Elements)
			}
			if startIdx >= endIdx {
				return &object.Array{Elements: []object.Object{}}
			}
			return &object.Array{Elements: arr.Elements[startIdx:endIdx]}
		},
	},

	"sort": {
		Name: "sort",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("sort() takes 1 argument")
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("sort() requires an array")
			}
			newElements := make([]object.Object, len(arr.Elements))
			copy(newElements, arr.Elements)
			sort.Slice(newElements, func(i, j int) bool {
				return newElements[i].Inspect() < newElements[j].Inspect()
			})
			return &object.Array{Elements: newElements}
		},
	},

	"unique": {
		Name: "unique",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("unique() takes 1 argument")
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("unique() requires an array")
			}
			seen := make(map[string]bool)
			var result []object.Object
			for _, el := range arr.Elements {
				key := el.Inspect()
				if !seen[key] {
					seen[key] = true
					result = append(result, el)
				}
			}
			return &object.Array{Elements: result}
		},
	},

	"flatten": {
		Name: "flatten",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("flatten() takes 1 argument")
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("flatten() requires an array")
			}
			var result []object.Object
			var flattenRecursive func([]object.Object)
			flattenRecursive = func(elements []object.Object) {
				for _, el := range elements {
					if inner, ok := el.(*object.Array); ok {
						flattenRecursive(inner.Elements)
					} else {
						result = append(result, el)
					}
				}
			}
			flattenRecursive(arr.Elements)
			return &object.Array{Elements: result}
		},
	},

	"map": {
		Name: "map",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("map() takes 2 arguments: array, function")
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("map() first arg must be an array")
			}
			fn, ok := args[1].(*object.Function)
			if !ok {
				return newError("map() second arg must be a function")
			}
			results := make([]object.Object, len(arr.Elements))
			for i, el := range arr.Elements {
				env := object.NewEnclosedEnvironment(fn.Env)
				if len(fn.Parameters) > 0 {
					env.Set(fn.Parameters[0].Name.Value, el)
				}
				if len(fn.Parameters) > 1 {
					env.Set(fn.Parameters[1].Name.Value, &object.Integer{Value: int64(i)})
				}
				result := EvalFn(fn.Body, env)
				if result != nil && result.Type() == object.ERROR_OBJ {
					return result
				}
				results[i] = unwrapReturn(result)
			}
			return &object.Array{Elements: results}
		},
	},

	"filter": {
		Name: "filter",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("filter() takes 2 arguments")
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("filter() first arg must be an array")
			}
			fn, ok := args[1].(*object.Function)
			if !ok {
				return newError("filter() second arg must be a function")
			}
			var results []object.Object
			for i, el := range arr.Elements {
				env := object.NewEnclosedEnvironment(fn.Env)
				if len(fn.Parameters) > 0 {
					env.Set(fn.Parameters[0].Name.Value, el)
				}
				if len(fn.Parameters) > 1 {
					env.Set(fn.Parameters[1].Name.Value, &object.Integer{Value: int64(i)})
				}
				result := EvalFn(fn.Body, env)
				if result != nil && result.Type() == object.ERROR_OBJ {
					return result
				}
				if isTruthyBuiltin(unwrapReturn(result)) {
					results = append(results, el)
				}
			}
			return &object.Array{Elements: results}
		},
	},

	"reduce": {
		Name: "reduce",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("reduce() takes 3 arguments: array, function, initial")
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("reduce() first arg must be an array")
			}
			fn, ok := args[1].(*object.Function)
			if !ok {
				return newError("reduce() second arg must be a function")
			}
			acc := args[2]
			for _, el := range arr.Elements {
				env := object.NewEnclosedEnvironment(fn.Env)
				if len(fn.Parameters) > 0 {
					env.Set(fn.Parameters[0].Name.Value, acc)
				}
				if len(fn.Parameters) > 1 {
					env.Set(fn.Parameters[1].Name.Value, el)
				}
				result := EvalFn(fn.Body, env)
				if result != nil && result.Type() == object.ERROR_OBJ {
					return result
				}
				acc = unwrapReturn(result)
			}
			return acc
		},
	},

	"keys": {
		Name: "keys",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("keys() takes 1 argument")
			}
			m, ok := args[0].(*object.Map)
			if !ok {
				return newError("keys() requires a map")
			}
			var keys []object.Object
			for _, pair := range m.Pairs {
				keys = append(keys, pair.Key)
			}
			return &object.Array{Elements: keys}
		},
	},

	"values": {
		Name: "values",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("values() takes 1 argument")
			}
			m, ok := args[0].(*object.Map)
			if !ok {
				return newError("values() requires a map")
			}
			var values []object.Object
			for _, pair := range m.Pairs {
				values = append(values, pair.Value)
			}
			return &object.Array{Elements: values}
		},
	},

	"int": {
		Name: "int",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("int() takes 1 argument")
			}
			switch arg := args[0].(type) {
			case *object.Integer:
				return arg
			case *object.Float:
				return &object.Integer{Value: int64(arg.Value)}
			case *object.String:
				var i int64
				_, err := fmt.Sscanf(arg.Value, "%d", &i)
				if err != nil {
					return newError("cannot convert %q to int", arg.Value)
				}
				return &object.Integer{Value: i}
			case *object.Boolean:
				if arg.Value {
					return &object.Integer{Value: 1}
				}
				return &object.Integer{Value: 0}
			default:
				return newError("cannot convert %s to int", args[0].Type())
			}
		},
	},

	"float": {
		Name: "float",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("float() takes 1 argument")
			}
			switch arg := args[0].(type) {
			case *object.Float:
				return arg
			case *object.Integer:
				return &object.Float{Value: float64(arg.Value)}
			case *object.String:
				var f float64
				_, err := fmt.Sscanf(arg.Value, "%f", &f)
				if err != nil {
					return newError("cannot convert %q to float", arg.Value)
				}
				return &object.Float{Value: f}
			default:
				return newError("cannot convert %s to float", args[0].Type())
			}
		},
	},

	"str": {
		Name: "str",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("str() takes 1 argument")
			}
			return &object.String{Value: args[0].Inspect()}
		},
	},

	"find": {
		Name: "find",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("find() takes 2 arguments")
			}
			switch container := args[0].(type) {
			case *object.String:
				needle, ok := args[1].(*object.String)
				if !ok {
					return newError("find() needle must be a string")
				}
				idx := strings.Index(container.Value, needle.Value)
				return &object.Integer{Value: int64(idx)}
			case *object.Array:
				for i, el := range container.Elements {
					if el.Inspect() == args[1].Inspect() {
						return &object.Integer{Value: int64(i)}
					}
				}
				return &object.Integer{Value: -1}
			default:
				return newError("find() requires string or array")
			}
		},
	},

	"range": {
		Name: "range",
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("range() takes 1-2 arguments")
			}
			var start, end int64
			if len(args) == 1 {
				n, ok := args[0].(*object.Integer)
				if !ok {
					return newError("range() requires integers")
				}
				start = 0
				end = n.Value
			} else {
				s, ok := args[0].(*object.Integer)
				if !ok {
					return newError("range() requires integers")
				}
				e, ok := args[1].(*object.Integer)
				if !ok {
					return newError("range() requires integers")
				}
				start = s.Value
				end = e.Value
			}
			return &object.Range{Start: start, End: end}
		},
	},
}
