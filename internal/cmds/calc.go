package cmds

import (
	"fmt"
	"math"

	"github.com/Knetic/govaluate"
	"github.com/awfufu/qbot"
)

const calcHelpMsg string = `Calculates the result of a mathematical expression.
Usage: /calc -l|<expression>
Options:
	-l	List supported functions
Example: /calc 1+1`

var calcCommand *Command = &Command{
	Name:       "calc",
	HelpMsg:    calcHelpMsg,
	Permission: getCmdPermLevel("calc"),
	NeedRawMsg: false,
	MinArgs:    2,
	Exec:       calcExec,
}

func unaryMathOp(op func(float64) float64) func(args ...any) (any, error) {
	return func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 argument")
		}
		val, ok := args[0].(float64)
		if !ok {
			return nil, fmt.Errorf("argument must be a number")
		}
		return op(val), nil
	}
}

func binaryMathOp(op func(float64, float64) float64) func(args ...any) (any, error) {
	return func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 arguments")
		}
		val1, ok1 := args[0].(float64)
		val2, ok2 := args[1].(float64)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("arguments must be numbers")
		}
		return op(val1, val2), nil
	}
}

var calcFunctions = map[string]govaluate.ExpressionFunction{
	"sin":   unaryMathOp(math.Sin),
	"cos":   unaryMathOp(math.Cos),
	"tan":   unaryMathOp(math.Tan),
	"asin":  unaryMathOp(math.Asin),
	"acos":  unaryMathOp(math.Acos),
	"atan":  unaryMathOp(math.Atan),
	"sqrt":  unaryMathOp(math.Sqrt),
	"cbrt":  unaryMathOp(math.Cbrt),
	"abs":   unaryMathOp(math.Abs),
	"ceil":  unaryMathOp(math.Ceil),
	"floor": unaryMathOp(math.Floor),
	"round": unaryMathOp(math.Round),
	"ln":    unaryMathOp(math.Log),
	"log2":  unaryMathOp(math.Log2),
	"log10": unaryMathOp(math.Log10),
	"exp":   unaryMathOp(math.Exp),
	"pow":   binaryMathOp(math.Pow),
	"max":   binaryMathOp(math.Max),
	"min":   binaryMathOp(math.Min),
}

func calcExec(b *qbot.Sender, msg *qbot.Message) {
	exprString := ""
	for _, item := range msg.Array[1:] {
		if item.Type() == qbot.TextType {
			exprString += item.Text()
		} else {
			b.SendGroupMsg(msg.GroupID, "invalid expression")
			return
		}
	}

	if exprString == "-l" {
		var functionsList string
		for k := range calcFunctions {
			functionsList += k + ", "
		}
		functionsList = functionsList[:len(functionsList)-2]
		b.SendGroupMsg(msg.GroupID, "supported functions: "+functionsList)
		return
	}

	expression, err := govaluate.NewEvaluableExpressionWithFunctions(exprString, calcFunctions)
	if err != nil {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("%v", err))
		return
	}

	result, err := expression.Evaluate(map[string]any{
		"pi": math.Pi,
		"PI": math.Pi,
		"π":  math.Pi,
		"e":  math.E,
		"E":  math.E,
	})

	if err != nil {
		b.SendGroupMsg(msg.GroupID, fmt.Sprintf("%v", err))
		return
	}

	b.SendGroupMsg(msg.GroupID, fmt.Sprintf("%v", result))
}
