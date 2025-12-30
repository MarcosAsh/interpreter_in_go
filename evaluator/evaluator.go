package evaluator

import (
	"fmt"
	"pearl/ast"
	"pearl/object"
	"regexp"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	case *ast.Program:
		return evalProgram(node, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
		return val

	case *ast.ReturnStatement:
		if node.ReturnValue == nil {
			return &object.ReturnValue{Value: NULL}
		}
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.ForStatement:
		return evalForStatement(node, env)

	case *ast.WhileStatement:
		return evalWhileStatement(node, env)

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}

	case *ast.StringLiteral:
		return evalStringLiteral(node, env)

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.NullLiteral:
		return NULL

	case *ast.RegexLiteral:
		re, err := regexp.Compile(node.Pattern)
		if err != nil {
			return newError("invalid regex pattern: %s", err)
		}
		return &object.Regex{Pattern: node.Pattern, Regexp: re}

	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

	case *ast.MapLiteral:
		return evalMapLiteral(node, env)

	case *ast.RangeLiteral:
		return evalRangeLiteral(node, env)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		if node.Operator == "and" || node.Operator == "or" {
			return evalLogicalExpression(node, env)
		}

		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)

	case *ast.IfExpression:
		return evalIfExpression(node, env)

	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		fn := &object.Function{Parameters: params, Body: body, Env: env, Name: node.Name}
		if node.Name != "" {
			env.Set(node.Name, fn)
		}
		return fn

	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalCallArguments(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args, node.Arguments)

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)

	case *ast.PipeExpression:
		return evalPipeExpression(node, env)

	case *ast.AssignExpression:
		return evalAssignExpression(node, env)
	}

	return nil
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalStringLiteral(sl *ast.StringLiteral, env *object.Environment) object.Object {
	if len(sl.Parts) == 0 {
		return &object.String{Value: sl.Value}
	}

	var result string
	for _, part := range sl.Parts {
		if part.IsExpr {
			val := Eval(part.Expr, env)
			if isError(val) {
				return val
			}
			result += val.Inspect()
		} else {
			result += part.Text
		}
	}
	return &object.String{Value: result}
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object
	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func evalCallArguments(args []ast.CallArg, env *object.Environment) []object.Object {
	var result []object.Object
	for _, a := range args {
		evaluated := Eval(a.Value, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func evalMapLiteral(node *ast.MapLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.MapPair)

	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as map key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.MapPair{Key: key, Value: value}
	}

	return &object.Map{Pairs: pairs}
}

func evalRangeLiteral(node *ast.RangeLiteral, env *object.Environment) object.Object {
	start := Eval(node.Start, env)
	if isError(start) {
		return start
	}
	end := Eval(node.End, env)
	if isError(end) {
		return end
	}

	startInt, ok := start.(*object.Integer)
	if !ok {
		return newError("range start must be an integer, got %s", start.Type())
	}
	endInt, ok := end.(*object.Integer)
	if !ok {
		return newError("range end must be an integer, got %s", end.Type())
	}

	return &object.Range{Start: startInt.Value, End: endInt.Value}
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("undefined variable: %s", node.Value)
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!", "not":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	if isTruthy(right) {
		return FALSE
	}
	return TRUE
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	switch obj := right.(type) {
	case *object.Integer:
		return &object.Integer{Value: -obj.Value}
	case *object.Float:
		return &object.Float{Value: -obj.Value}
	default:
		return newError("unknown operator: -%s", right.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ || right.Type() == object.FLOAT_OBJ:
		return evalFloatInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.REGEX_OBJ:
		return evalRegexMatchExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError("division by zero")
		}
		return &object.Integer{Value: leftVal / rightVal}
	case "%":
		if rightVal == 0 {
			return newError("division by zero")
		}
		return &object.Integer{Value: leftVal % rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalFloatInfixExpression(operator string, left, right object.Object) object.Object {
	var leftVal, rightVal float64

	switch l := left.(type) {
	case *object.Float:
		leftVal = l.Value
	case *object.Integer:
		leftVal = float64(l.Value)
	}

	switch r := right.(type) {
	case *object.Float:
		rightVal = r.Value
	case *object.Integer:
		rightVal = float64(r.Value)
	}

	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError("division by zero")
		}
		return &object.Float{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch operator {
	case "++":
		return &object.String{Value: leftVal + rightVal}
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalRegexMatchExpression(operator string, left, right object.Object) object.Object {
	str := left.(*object.String).Value
	re := right.(*object.Regex).Regexp
	matched := re.MatchString(str)

	switch operator {
	case "~":
		return nativeBoolToBooleanObject(matched)
	case "!~":
		return nativeBoolToBooleanObject(!matched)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalLogicalExpression(node *ast.InfixExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	if node.Operator == "and" {
		if !isTruthy(left) {
			return left
		}
		return Eval(node.Right, env)
	}

	// or
	if isTruthy(left) {
		return left
	}
	return Eval(node.Right, env)
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

func evalForStatement(fs *ast.ForStatement, env *object.Environment) object.Object {
	iterable := Eval(fs.Iterable, env)
	if isError(iterable) {
		return iterable
	}

	var result object.Object = NULL

	switch obj := iterable.(type) {
	case *object.Array:
		for _, elem := range obj.Elements {
			innerEnv := object.NewEnclosedEnvironment(env)
			innerEnv.Set(fs.Variable.Value, elem)
			result = Eval(fs.Body, innerEnv)
			if isError(result) {
				return result
			}
			if _, ok := result.(*object.ReturnValue); ok {
				return result
			}
		}

	case *object.Range:
		for i := obj.Start; i < obj.End; i++ {
			innerEnv := object.NewEnclosedEnvironment(env)
			innerEnv.Set(fs.Variable.Value, &object.Integer{Value: i})
			result = Eval(fs.Body, innerEnv)
			if isError(result) {
				return result
			}
			if _, ok := result.(*object.ReturnValue); ok {
				return result
			}
		}

	case *object.String:
		for _, ch := range obj.Value {
			innerEnv := object.NewEnclosedEnvironment(env)
			innerEnv.Set(fs.Variable.Value, &object.String{Value: string(ch)})
			result = Eval(fs.Body, innerEnv)
			if isError(result) {
				return result
			}
			if _, ok := result.(*object.ReturnValue); ok {
				return result
			}
		}

	case *object.Map:
		for _, pair := range obj.Pairs {
			innerEnv := object.NewEnclosedEnvironment(env)
			innerEnv.Set(fs.Variable.Value, pair.Key)
			result = Eval(fs.Body, innerEnv)
			if isError(result) {
				return result
			}
			if _, ok := result.(*object.ReturnValue); ok {
				return result
			}
		}

	default:
		return newError("cannot iterate over %s", iterable.Type())
	}

	return result
}

func evalWhileStatement(ws *ast.WhileStatement, env *object.Environment) object.Object {
	var result object.Object = NULL

	for {
		condition := Eval(ws.Condition, env)
		if isError(condition) {
			return condition
		}
		if !isTruthy(condition) {
			break
		}

		result = Eval(ws.Body, env)
		if isError(result) {
			return result
		}
		if _, ok := result.(*object.ReturnValue); ok {
			return result
		}
	}

	return result
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.STRING_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalStringIndexExpression(left, index)
	case left.Type() == object.MAP_OBJ:
		return evalMapIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s[%s]", left.Type(), index.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arr := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arr.Elements) - 1)

	if idx < 0 {
		idx = int64(len(arr.Elements)) + idx
	}
	if idx < 0 || idx > max {
		return NULL
	}

	return arr.Elements[idx]
}

func evalStringIndexExpression(str, index object.Object) object.Object {
	s := str.(*object.String)
	idx := index.(*object.Integer).Value
	max := int64(len(s.Value) - 1)

	if idx < 0 {
		idx = int64(len(s.Value)) + idx
	}
	if idx < 0 || idx > max {
		return NULL
	}

	return &object.String{Value: string(s.Value[idx])}
}

func evalMapIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Map)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as map key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
}

func evalPipeExpression(pe *ast.PipeExpression, env *object.Environment) object.Object {
	left := Eval(pe.Left, env)
	if isError(left) {
		return left
	}

	switch right := pe.Right.(type) {
	case *ast.CallExpression:
		fn := Eval(right.Function, env)
		if isError(fn) {
			return fn
		}

		args := []object.Object{left}
		for _, a := range right.Arguments {
			arg := Eval(a.Value, env)
			if isError(arg) {
				return arg
			}
			args = append(args, arg)
		}

		return applyFunction(fn, args, nil)

	case *ast.Identifier:
		fn := evalIdentifier(right, env)
		if isError(fn) {
			return fn
		}
		return applyFunction(fn, []object.Object{left}, nil)

	default:
		return newError("right side of pipe must be a function call")
	}
}

func evalAssignExpression(ae *ast.AssignExpression, env *object.Environment) object.Object {
	val := Eval(ae.Value, env)
	if isError(val) {
		return val
	}

	switch target := ae.Name.(type) {
	case *ast.Identifier:
		if !env.Update(target.Value, val) {
			return newError("undefined variable: %s", target.Value)
		}
		return val

	case *ast.IndexExpression:
		left := Eval(target.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(target.Index, env)
		if isError(index) {
			return index
		}

		switch obj := left.(type) {
		case *object.Array:
			idx := index.(*object.Integer).Value
			if idx < 0 || idx >= int64(len(obj.Elements)) {
				return newError("array index out of bounds: %d", idx)
			}
			obj.Elements[idx] = val
			return val

		case *object.Map:
			key, ok := index.(object.Hashable)
			if !ok {
				return newError("unusable as map key: %s", index.Type())
			}
			obj.Pairs[key.HashKey()] = object.MapPair{Key: index, Value: val}
			return val

		default:
			return newError("cannot assign to index of %s", left.Type())
		}

	default:
		return newError("cannot assign to this expression")
	}
}

func applyFunction(fn object.Object, args []object.Object, callArgs []ast.CallArg) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args, callArgs)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		result := fn.Fn(args...)
		if result != nil {
			return result
		}
		return NULL

	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object, callArgs []ast.CallArg) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	namedArgs := make(map[string]object.Object)
	positionalArgs := []object.Object{}

	for i, arg := range args {
		if callArgs != nil && i < len(callArgs) && callArgs[i].Name != "" {
			namedArgs[callArgs[i].Name] = arg
		} else {
			positionalArgs = append(positionalArgs, arg)
		}
	}

	posIdx := 0
	for _, param := range fn.Parameters {
		name := param.Name.Value

		if val, ok := namedArgs[name]; ok {
			env.Set(name, val)
			continue
		}

		if posIdx < len(positionalArgs) {
			env.Set(name, positionalArgs[posIdx])
			posIdx++
			continue
		}

		if param.Default != nil {
			val := Eval(param.Default, fn.Env)
			env.Set(name, val)
			continue
		}

		env.Set(name, NULL)
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func isTruthy(obj object.Object) bool {
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

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
