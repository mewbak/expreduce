package expreduce

import (
	"math/big"
	"strings"

	"github.com/corywalker/expreduce/pkg/expreduceapi"
)

func ExArrayContainsFloat(a []expreduceapi.Ex) bool {
	res := false
	for _, e := range a {
		_, isfloat := e.(*Flt)
		res = res || isfloat
	}
	return res
}

func RationalAssertion(num expreduceapi.Ex, den expreduceapi.Ex) (r *Rational, isR bool) {
	numInt, numIsInt := num.(*Integer)
	denPow, denIsPow := HeadAssertion(den, "System`Power")
	if !numIsInt || !denIsPow {
		return nil, false
	}
	powInt, powIsInt := denPow.GetParts()[2].(*Integer)
	if !powIsInt {
		return nil, false
	}
	if powInt.Val.Cmp(big.NewInt(-1)) != 0 {
		return nil, false
	}
	denInt, denIsInt := denPow.GetParts()[1].(*Integer)
	if !denIsInt {
		return nil, false
	}
	return NewRational(numInt.Val, denInt.Val), true
}

type FoldFn int

const (
	FoldFnAdd FoldFn = iota
	FoldFnMul
)

func typedRealPart(fn FoldFn, i *Integer, r *Rational, f *Flt, c *Complex) expreduceapi.Ex {
	if c != nil {
		toReturn := c
		if f != nil {
			if fn == FoldFnAdd {
				toReturn.AddF(f)
			} else if fn == FoldFnMul {
				toReturn.MulF(f)
			}
		}
		if r != nil {
			if fn == FoldFnAdd {
				toReturn.AddR(r)
			} else if fn == FoldFnMul {
				toReturn.MulR(r)
			}
		}
		if i != nil {
			if fn == FoldFnAdd {
				toReturn.AddI(i)
			} else if fn == FoldFnMul {
				toReturn.MulI(i)
			}
		}
		return toReturn
	}
	if f != nil {
		toReturn := f
		if r != nil {
			if fn == FoldFnAdd {
				toReturn.AddR(r)
			} else if fn == FoldFnMul {
				toReturn.MulR(r)
			}
		}
		if i != nil {
			if fn == FoldFnAdd {
				toReturn.AddI(i)
			} else if fn == FoldFnMul {
				toReturn.MulI(i)
			}
		}
		return toReturn
	}
	if r != nil {
		toReturn := r
		if i != nil {
			if fn == FoldFnAdd {
				toReturn.AddI(i)
			} else if fn == FoldFnMul {
				toReturn.MulI(i)
			}
		}
		return toReturn
	}
	if i != nil {
		return i
	}
	return nil
}

func computeNumericPart(fn FoldFn, e expreduceapi.ExpressionInterface) (expreduceapi.Ex, int) {
	var foldedInt *Integer
	var foldedRat *Rational
	var foldedFlt *Flt
	var foldedComp *Complex
	for i := 1; i < len(e.GetParts()); i++ {
		// TODO: implement short circuiting if we encounter a zero while
		// multiplying.
		asInt, isInt := e.GetParts()[i].(*Integer)
		if isInt {
			if foldedInt == nil {
				// Try deepcopy if problems. I think this does not cause
				// problems now because we will only modify the value if we end
				// up creating an entirely new expression.
				foldedInt = asInt.DeepCopy().(*Integer)
				continue
			}
			if fn == FoldFnAdd {
				foldedInt.AddI(asInt)
			} else if fn == FoldFnMul {
				foldedInt.MulI(asInt)
			}
			continue
		}
		asRat, isRat := e.GetParts()[i].(*Rational)
		if isRat {
			if foldedRat == nil {
				foldedRat = asRat.DeepCopy().(*Rational)
				continue
			}
			if fn == FoldFnAdd {
				foldedRat.AddR(asRat)
			} else if fn == FoldFnMul {
				foldedRat.MulR(asRat)
			}
			continue
		}
		asFlt, isFlt := e.GetParts()[i].(*Flt)
		if isFlt {
			if foldedFlt == nil {
				foldedFlt = asFlt.DeepCopy().(*Flt)
				continue
			}
			if fn == FoldFnAdd {
				foldedFlt.AddF(asFlt)
			} else if fn == FoldFnMul {
				foldedFlt.MulF(asFlt)
			}
			continue
		}
		asComp, isComp := e.GetParts()[i].(*Complex)
		if isComp {
			if foldedComp == nil {
				foldedComp = asComp.DeepCopy().(*Complex)
				continue
			}
			if fn == FoldFnAdd {
				foldedComp.AddC(asComp)
			} else if fn == FoldFnMul {
				foldedComp.MulC(asComp)
			}
			continue
		}
		return typedRealPart(fn, foldedInt, foldedRat, foldedFlt, foldedComp), i
	}
	return typedRealPart(fn, foldedInt, foldedRat, foldedFlt, foldedComp), -1
}

// Define a special NumberQ for our purposes since this logic does not support
// complex numbers yet. TODO(corywalker): fix this.
func numberQForTermCollection(e expreduceapi.Ex) bool {
	// _, ok := e.(*Complex)
	// if ok {
	// 	return false
	// }
	return numberQ(e)
}

func splitTerm(e expreduceapi.Ex) (expreduceapi.Ex, expreduceapi.Ex, bool) {
	asSym, isSym := e.(*Symbol)
	if isSym {
		return NewInteger(big.NewInt(1)), NewExpression([]expreduceapi.Ex{
			NewSymbol("System`Times"),
			asSym,
		}), true
	}
	asTimes, isTimes := HeadAssertion(e, "System`Times")
	if isTimes {
		if len(asTimes.GetParts()) < 2 {
			return nil, nil, false
		}
		if numberQForTermCollection(asTimes.GetParts()[1]) {
			if len(asTimes.GetParts()) > 2 {
				return asTimes.GetParts()[1], NewExpression(append([]expreduceapi.Ex{NewSymbol("System`Times")}, asTimes.GetParts()[2:]...)), true
			}
		} else {
			return NewInteger(big.NewInt(1)), NewExpression(append([]expreduceapi.Ex{NewSymbol("System`Times")}, asTimes.GetParts()[1:]...)), true
		}
	}
	asExpr, isExpr := e.(expreduceapi.ExpressionInterface)
	if isExpr {
		return NewInteger(big.NewInt(1)), NewExpression([]expreduceapi.Ex{
			NewSymbol("System`Times"),
			asExpr,
		}), true
	}
	return nil, nil, false
}

func collectedToTerm(coeffs []expreduceapi.Ex, vars expreduceapi.Ex, fullPart expreduceapi.Ex) expreduceapi.Ex {
	// Preserve the original expression if there is no need to change it.
	// We can keep all the cached values like the hash.
	if len(coeffs) == 1 {
		return fullPart
	}

	finalC, _ := computeNumericPart(FoldFnAdd, NewExpression(append([]expreduceapi.Ex{
		NewSymbol("System`Plus")}, coeffs...)))

	toAdd := NewExpression([]expreduceapi.Ex{NewSymbol("System`Times")})
	cAsInt, cIsInt := finalC.(*Integer)
	if !(cIsInt && cAsInt.Val.Cmp(big.NewInt(1)) == 0) {
		toAdd.GetParts() = append(toAdd.GetParts(), finalC)
	}
	vAsExpr, vIsExpr := HeadAssertion(vars, "System`Times")
	if vIsExpr && len(vAsExpr.GetParts()) == 2 {
		vars = vAsExpr.GetParts()[1]
	}
	toAdd.GetParts() = append(toAdd.GetParts(), vars)
	if len(toAdd.GetParts()) == 2 {
		return toAdd.GetParts()[1]
	}
	return toAdd
}

func collectTerms(e expreduceapi.ExpressionInterface) expreduceapi.ExpressionInterface {
	collected := NewExpression([]expreduceapi.Ex{NewSymbol("System`Plus")})
	var lastVars expreduceapi.Ex
	var lastFullPart expreduceapi.Ex
	lastCoeffs := []expreduceapi.Ex{}
	for _, part := range e.GetParts()[1:] {
		coeff, vars, isTerm := splitTerm(part)
		if isTerm {
			if lastVars == nil {
				lastCoeffs = []expreduceapi.Ex{coeff}
				lastVars = vars
				lastFullPart = part
			} else {
				if hashEx(vars) == hashEx(lastVars) {
					lastCoeffs = append(lastCoeffs, coeff)
				} else {
					collected.GetParts() = append(collected.GetParts(), collectedToTerm(lastCoeffs, lastVars, lastFullPart))

					lastCoeffs = []expreduceapi.Ex{coeff}
					lastVars = vars
					lastFullPart = part
				}
			}
		} else {
			collected.GetParts() = append(collected.GetParts(), part)
		}
	}
	if lastVars != nil {
		collected.GetParts() = append(collected.GetParts(), collectedToTerm(lastCoeffs, lastVars, lastFullPart))
	}
	return collected
}

func getArithmeticDefinitions() (defs []Definition) {
	defs = append(defs, Definition{
		Name:    "Plus",
		Default: "0",
		toString: func(this expreduceapi.ExpressionInterface, params expreduceapi.ToStringParams) (bool, string) {
			return ToStringInfix(this.GetParts()[1:], " + ", "System`Plus", params)
		},
		legacyEvalFn: func(this expreduceapi.ExpressionInterface, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
			// Calls without argument receive identity values
			if len(this.GetParts()) == 1 {
				return NewInteger(big.NewInt(0))
			}

			res := this
			realPart, symStart := computeNumericPart(FoldFnAdd, this)
			if realPart != nil {
				if symStart == -1 {
					return realPart
				}
				res = NewExpression([]expreduceapi.Ex{NewSymbol("System`Plus")})
				rAsInt, rIsInt := realPart.(*Integer)
				if !(rIsInt && rAsInt.Val.Cmp(big.NewInt(0)) == 0) {
					res.GetParts() = append(res.GetParts(), realPart)
				}
				res.GetParts() = append(res.GetParts(), this.GetParts()[symStart:]...)
			}

			collected := collectTerms(res)
			if hashEx(collected) != hashEx(res) {
				res = collected
			}

			// If one expression remains, replace this Plus with the expression
			if len(res.GetParts()) == 2 {
				return res.GetParts()[1]
			}

			// Not exactly right because of "1. + foo[1]", but close enough.
			if _, rIsReal := realPart.(*Flt); rIsReal {
				return exprToN(es, res)
			}
			return res
		},
	})
	defs = append(defs, Definition{
		Name: "Sum",
		legacyEvalFn: func(this expreduceapi.ExpressionInterface, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
			return evalIterationFunc(this, es, NewInteger(big.NewInt(0)), "System`Plus")
		},
	})
	defs = append(defs, Definition{
		Name:    "Times",
		Default: "1",
		toString: func(this expreduceapi.ExpressionInterface, params expreduceapi.ToStringParams) (bool, string) {
			delim := "*"
			if params.form == "TeXForm" {
				delim = " "
			}
			ok, res := ToStringInfix(this.GetParts()[1:], delim, "System`Times", params)
			if ok && strings.HasPrefix(res, "(-1)"+delim) {
				return ok, "-" + res[5:]
			}
			return ok, res
		},
		legacyEvalFn: func(this expreduceapi.ExpressionInterface, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
			// Calls without argument receive identity values
			if len(this.GetParts()) == 1 {
				return NewInteger(big.NewInt(1))
			}

			res := this
			realPart, symStart := computeNumericPart(FoldFnMul, this)
			if realPart != nil {
				if symStart == -1 {
					return realPart
				}
				res = NewExpression([]expreduceapi.Ex{NewSymbol("System`Times")})
				rAsInt, rIsInt := realPart.(*Integer)
				if rIsInt && rAsInt.Val.Cmp(big.NewInt(0)) == 0 {
					containsInfinity := MemberQ(this.GetParts()[symStart:], NewExpression([]expreduceapi.Ex{
						NewSymbol("System`Alternatives"),
						NewSymbol("System`Infinity"),
						NewSymbol("System`ComplexInfinity"),
						NewSymbol("System`Indeterminate"),
					}), es)
					if containsInfinity {
						return NewSymbol("System`Indeterminate")
					}
					return NewInteger(big.NewInt(0))
				}
				if !(rIsInt && rAsInt.Val.Cmp(big.NewInt(1)) == 0) {
					res.GetParts() = append(res.GetParts(), realPart)
				}
				res.GetParts() = append(res.GetParts(), this.GetParts()[symStart:]...)
			}

			// If one expression remains, replace this Times with the expression
			if len(res.GetParts()) == 2 {
				return res.GetParts()[1]
			}

			// Automatically Expand negations (*-1), not (*-1.) of a Plus expression
			// Perhaps better implemented as a rule.
			if len(res.GetParts()) == 3 {
				leftint, leftintok := res.GetParts()[1].(*Integer)
				rightplus, rightplusok := HeadAssertion(res.GetParts()[2], "System`Plus")
				if leftintok && rightplusok {
					if leftint.Val.Cmp(big.NewInt(-1)) == 0 {
						toreturn := NewExpression([]expreduceapi.Ex{NewSymbol("System`Plus")})
						addends := rightplus.GetParts()[1:len(rightplus.GetParts())]
						for i := range addends {
							toAppend := NewExpression([]expreduceapi.Ex{
								NewSymbol("System`Times"),
								addends[i],
								NewInteger(big.NewInt(-1)),
							})

							toreturn.GetParts() = append(toreturn.GetParts(), toAppend)
						}
						return toreturn.Eval(es)
					}
				}
			}

			// Not exactly right because of "1. + foo[1]", but close enough.
			if _, rIsReal := realPart.(*Flt); rIsReal {
				return exprToN(es, res)
			}
			return res
		},
	})
	defs = append(defs, Definition{
		Name: "Product",
		legacyEvalFn: func(this expreduceapi.ExpressionInterface, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
			return evalIterationFunc(this, es, NewInteger(big.NewInt(1)), "System`Times")
		},
	})
	defs = append(defs, Definition{Name: "Abs"})
	defs = append(defs, Definition{Name: "Divide"})
	defs = append(defs, Definition{Name: "Increment"})
	defs = append(defs, Definition{Name: "Decrement"})
	defs = append(defs, Definition{Name: "PreIncrement"})
	defs = append(defs, Definition{Name: "PreDecrement"})
	defs = append(defs, Definition{Name: "AddTo"})
	defs = append(defs, Definition{Name: "SubtractFrom"})
	return
}
