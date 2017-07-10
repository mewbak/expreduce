package expreduce

import (
	"github.com/corywalker/mathbigext"
	"math/big"
)

func GetPowerDefinitions() (defs []Definition) {
	defs = append(defs, Definition{
		Name:       "Power",
		Usage:      "`base^exp` finds `base` raised to the power of `exp`.",
		Attributes: []string{"Listable", "NumericFunction", "OneIdentity"},
		Default:	"1",
		Rules: []Rule{
			// Simplify nested exponents
			{"Power[Power[a_,b_Integer],c_Integer]", "a^(b*c)"},
			{"Power[Power[a_,b_Real],c_Integer]", "a^(b*c)"},
			{"Power[Power[a_,b_Symbol],c_Integer]", "a^(b*c)"},

			{"Power[Infinity, -1]", "0"},

			// Power definitions
			{"(Except[_Symbol, first_] * inner___)^Except[_Symbol, pow_]", "first^pow * Times[inner]^pow"},
			{"(first_ * inner___)^Except[_Symbol, pow_]", "first^pow * Times[inner]^pow"},

			// Rational simplifications
			{"Power[Rational[a_,b_], -1]", "Rational[b,a]"},
			{"Power[Rational[a_,b_], e_?Positive]", "Rational[a^e,b^e]"},
		},
		toString: func(this *Expression, form string) (bool, string) {
			return ToStringInfixAdvanced(this.Parts[1:], "^", false, "", "", form)
		},
		legacyEvalFn: func(this *Expression, es *EvalState) Ex {
			if len(this.Parts) != 3 {
				return this
			}

			// TODO: Handle cases like float raised to the float and things raised to
			// zero and 1

			baseInt, baseIsInt := this.Parts[1].(*Integer)
			powerInt, powerIsInt := this.Parts[2].(*Integer)
			baseFlt, baseIsFlt := this.Parts[1].(*Flt)
			powerFlt, powerIsFlt := this.Parts[2].(*Flt)
			// Anything raised to the 1st power is itself
			if powerIsFlt {
				if powerFlt.Val.Cmp(big.NewFloat(1)) == 0 {
					return this.Parts[1]
				}
			} else if powerIsInt {
				if powerInt.Val.Cmp(big.NewInt(1)) == 0 {
					return this.Parts[1]
				}
			}
			// Anything raised to the 0th power is 1, with a small exception
			isZerothPower := false
			if powerIsFlt {
				if powerFlt.Val.Cmp(big.NewFloat(0)) == 0 {
					isZerothPower = true
				}
			} else if powerIsInt {
				if powerInt.Val.Cmp(big.NewInt(0)) == 0 {
					isZerothPower = true
				}
			}
			isZeroBase := false
			if baseIsFlt {
				if baseFlt.Val.Cmp(big.NewFloat(0)) == 0 {
					isZeroBase = true
				}
			} else if baseIsInt {
				if baseInt.Val.Cmp(big.NewInt(0)) == 0 {
					isZeroBase = true
				}
			}
			if isZerothPower {
				if isZeroBase {
					return &Symbol{"Indeterminate"}
				}
				return &Integer{big.NewInt(1)}
			}

			//es.Debugf("Power eval. baseIsInt=%v, powerIsInt=%v", baseIsInt, powerIsInt)
			// Fully integer Power expression
			if baseIsInt && powerIsInt {
				cmpres := powerInt.Val.Cmp(big.NewInt(0))
				//es.Debugf("Cmpres: %v", cmpres)
				if cmpres == 1 {
					res := big.NewInt(0)
					res.Exp(baseInt.Val, powerInt.Val, nil)
					return &Integer{res}
				} else if cmpres == -1 {
					newbase := big.NewInt(0)
					absPower := big.NewInt(0)
					absPower.Abs(powerInt.Val)
					newbase.Exp(baseInt.Val, absPower, nil)
					if newbase.Cmp(big.NewInt(1)) == 0 {
						return &Integer{big.NewInt(1)}
					}
					if newbase.Cmp(big.NewInt(-1)) == 0 {
						return &Integer{big.NewInt(-1)}
					}
					//return NewExpression([]Ex{&Symbol{"Power"}, &Integer{newbase}, &Integer{big.NewInt(-1)}})
					return NewExpression([]Ex{&Symbol{"Rational"}, &Integer{big.NewInt(1)}, &Integer{newbase}})
				} else {
					return NewExpression([]Ex{&Symbol{"Error"}, &String{"Unexpected zero power in Power evaluation."}})
				}
			}

			if baseIsFlt && powerIsInt {
				return &Flt{mathbigext.Pow(baseFlt.Val, big.NewFloat(0).SetInt(powerInt.Val))}
			}
			if baseIsInt && powerIsFlt {
				return &Flt{mathbigext.Pow(big.NewFloat(0).SetInt(baseInt.Val), powerFlt.Val)}
			}
			if baseIsFlt && powerIsFlt {
				return &Flt{mathbigext.Pow(baseFlt.Val, powerFlt.Val)}
			}

			powerRat, powerIsRat := this.Parts[2].(*Rational)
			if powerIsRat {
				if powerRat.Num.Cmp(big.NewInt(1)) == 0 && powerRat.Den.Cmp(big.NewInt(2)) == 0 {
					return NewExpression([]Ex{
						&Symbol{"Sqrt"},
						this.Parts[1],
					})

				}
			}

			return this
		},
		SimpleExamples: []TestInstruction{
			&TestComment{"Exponents of integers are computed exactly:"},
			&StringTest{"-1/125", "(-5)^-3"},
			&TestComment{"Floating point exponents are handled with floating point precision:"},
			&StringTest{"1.99506e+3010", ".5^-10000."},
			&TestComment{"Automatically apply some basic simplification rules:"},
			&SameTest{"m^4.", "(m^2.)^2"},
		},
		FurtherExamples: []TestInstruction{
			&TestComment{"Expreduce handles problematic exponents accordingly:"},
			&StringTest{"Indeterminate", "0^0"},
			&SameTest{"ComplexInfinity", "0^(-1)"},
		},
		Tests: []TestInstruction{
			// Test raising expressions to the first power
			&StringTest{"(1 + x)", "(x+1)^1"},
			&StringTest{"0", "0^1"},
			&StringTest{"0.", "0.^1"},
			&StringTest{"-5", "-5^1"},
			&StringTest{"-5.5", "-5.5^1"},
			&StringTest{"(1 + x)", "(x+1)^1."},
			&StringTest{"0", "0^1."},
			&StringTest{"0.", "0.^1."},
			&StringTest{"-5", "(-5)^1."},
			&StringTest{"-5.5", "-5.5^1."},

			// Test raising expressions to the zero power
			&StringTest{"1", "(x+1)^0"},
			&StringTest{"Indeterminate", "0^0"},
			&StringTest{"Indeterminate", "0.^0"},
			&StringTest{"-1", "-5^0"},
			&StringTest{"1", "(-5)^0"},
			&StringTest{"1", "(-5.5)^0"},
			&StringTest{"1", "(x+1)^0."},
			&StringTest{"Indeterminate", "0^0."},
			&StringTest{"Indeterminate", "0.^0."},
			&StringTest{"-1", "-5^0."},
			&StringTest{"1", "(-5.5)^0."},
			&StringTest{"-1", "-5^0"},
			&StringTest{"1", "99^0"},

			&StringTest{"125", "5^3"},
			&StringTest{"1/125", "5^-3"},
			&StringTest{"-125", "(-5)^3"},
			&StringTest{"-1/125", "(-5)^-3"},

			&StringTest{"2.97538e+1589", "39^999."},
			&StringTest{"3.36092e-1590", "39^-999."},
			&StringTest{"1.99506e+3010", ".5^-10000."},
			&StringTest{"1.99506e+3010", ".5^-10000"},

			&StringTest{"1", "1^1"},
			&StringTest{"1", "1^2"},
			&StringTest{"1", "1^0"},
			&StringTest{"1", "1^-1"},
			&StringTest{"1", "1^-2"},
			&StringTest{"1", "1^99999992"},
			&StringTest{"1.", "1^2."},
			&StringTest{"1.", "1^99999992."},
			&StringTest{"1.", "1.^30"},
			&StringTest{"4.", "(1.*2*1.)^2"},
			&StringTest{"-1", "(-1)^1"},
			&StringTest{"1", "(-1)^2"},
			&StringTest{"1", "(-1)^0"},
			&StringTest{"1", "(-1)^0"},
			&StringTest{"-1", "(-1)^-1"},
			&StringTest{"1", "(-1)^-2"},
			&StringTest{"1", "(-1)^99999992"},
			&StringTest{"1.", "(-1.)^30"},
			&StringTest{"4.", "(1.*2*-1.)^2"},
			&StringTest{"-0.5", "(1.*2*-1.)^(-1)"},

			&SameTest{"Rational", "Power[2, -1] // Head"},
			&SameTest{"Integer", "Power[1, -1] // Head"},
			&SameTest{"Integer", "Power[2, 2] // Head"},
			&SameTest{"Rational", "Power[-2, -1] // Head"},
			&SameTest{"Rational", "Power[2, -2] // Head"},

			// Exponent simplifications
			&SameTest{"m^4", "m^2*m^2"},
			&SameTest{"m^4", "(m^2)^2"},
			&SameTest{"(m^2)^2.", "(m^2)^2."},
			&SameTest{"(m^2.)^2.", "(m^2.)^2."},
			&SameTest{"m^4.", "(m^2.)^2"},

			&SameTest{"ComplexInfinity", "0^(-1)"},

			&SameTest{"{1}", "ReplaceAll[a, a^p_. -> {p}]"},
		},
		KnownFailures: []TestInstruction{
			// Fix these when I have Abs functionality
			&StringTest{"2.975379863266329e+1589", "39^999."},
			&StringTest{"3.360915398890324e-1590", "39^-999."},
			&StringTest{"1.9950631168791027e+3010", ".5^-10000."},
			&StringTest{"1.9950631168791027e+3010", ".5^-10000"},
		},
	})
	defs = append(defs, Definition{
		Name: "PowerExpand",
		// This function is not implemented to a satisfiable extent. Do not
		// document it yet.
		OmitDocumentation: true,
		SimpleExamples: []TestInstruction{
			&TestComment{"`PowerExpand` can expand nested log expressions:"},
			&SameTest{"Log[a] + e (Log[b] + d Log[c])", "PowerExpand[Log[a (b c^d)^e]]"},
		},
		Rules: []Rule{
			{"PowerExpand[exp_]", "exp //. {Log[x_ y_]:>Log[x]+Log[y],Log[x_^k_]:>k Log[x]}"},
		},
	})
	defs = append(defs, Definition{
		Name:  "Expand",
		Usage: "`Expand[expr]` attempts to expand `expr`.",
		SimpleExamples: []TestInstruction{
			&SameTest{"a^3 + 3 a^2 * b + 3 a b^2 + b^3 + 3 a^2 * c + 6 a b c + 3 b^2 * c + 3 a c^2 + 3 b c^2 + c^3", "Expand[(a + b + c)^3]"},
			&SameTest{"a c + b c + a d + b d + a e + b e", "(a + b) * (c + d + e) // Expand"},
			&SameTest{"a d^2 + b d^2 + c d^2 + 2 a d e + 2 b d e + 2 c d e + a e^2 + b e^2 + c e^2", "(a + b + c)*(d + e)^2 // Expand"},
			&SameTest{"a^(2 b) + 2 a^b * c^d + c^(2 d)", "Expand[(a^b + c^d)^2]"},
			&SameTest{"a/d + b/d + c/d", "Expand[(a + b + c)/d]"},
			&SameTest{"1/d + (2 a)/d + a^2/d + b/d + c/d", "Expand[((a + 1)^2 + b + c)/d]"},
			&SameTest{"2 + 2 a", "2*(a + 1) // Expand"},
		},
		Tests: []TestInstruction{
		},
		KnownDangerous: []TestInstruction{
			// The following tests should not take 10 seconds:
			&SameTest{"Null", "((60 * c * a^2 * b^2) + (30 * c * a^2 * b^2) + (30 * c * a^2 * b^2) + a^5 + b^5 + c^5 + (5 * a * b^4) + (5 * a * c^4) + (5 * b * a^4) + (5 * b * c^4) + (5 * c * a^4) + (5 * c * b^4) + (10 * a^2 * b^3) + (10 * a^2 * c^3) + (10 * a^3 * b^2) + (10 * a^3 * c^2) + (10 * b^2 * c^3) + (10 * b^3 * c^2) + (20 * a * b * c^3) + (20 * a * c * b^3) + (20 * b * c * a^3));"},
		},
	})
	defs = append(defs, Definition{
		Name:  "PolynomialQ",
		Usage:  "`PolynomialQ[e, var]` returns True if `e` is a polynomial in `var`.",
		Tests: []TestInstruction{
			&SameTest{"True", "PolynomialQ[2x^2-3x+2, x]"},
			&SameTest{"True", "PolynomialQ[2x^2, x]"},
			&SameTest{"False", "PolynomialQ[2x^2.5, x]"},
			&SameTest{"False", "PolynomialQ[2x^-2, x]"},
			&SameTest{"True", "PolynomialQ[2x^0, x]"},
			&SameTest{"False", "PolynomialQ[2x^y, x]"},
			&SameTest{"True", "PolynomialQ[-3x, x]"},
			&SameTest{"True", "PolynomialQ[2, x]"},
			&SameTest{"True", "PolynomialQ[2y^2, x]"},
			&SameTest{"True", "PolynomialQ[-3y, x]"},
			&SameTest{"False", "PolynomialQ[2x^2-3x+2+Cos[x], x]"},
			&SameTest{"False", "PolynomialQ[Cos[x], x]"},
			&SameTest{"False", "PolynomialQ[2x^2-3x+Cos[x], x]"},
			&SameTest{"False", "PolynomialQ[2x^2-x*Cos[x], x]"},
			&SameTest{"True", "PolynomialQ[2x^2-x*Cos[y], x]"},
			&SameTest{"True", "PolynomialQ[2.5x^2-3x+2.5, 2.5]"},
			&SameTest{"True", "PolynomialQ[2x^2-3x+2, \"hello\"]"},
			&SameTest{"True", "PolynomialQ[2x^2-3x+2, y]"},
			&SameTest{"True", "PolynomialQ[x, y]"},
			&SameTest{"True", "PolynomialQ[y, y]"},
			&SameTest{"True", "PolynomialQ[2*x, y]"},
			&SameTest{"False", "PolynomialQ[2*x^Sin[y], x]"},
			&SameTest{"True", "PolynomialQ[Sin[y]*x^2, x]"},
			&SameTest{"False", "PolynomialQ[Sin[y]*x^2.5, x]"},
			&SameTest{"False", "PolynomialQ[Sin[y]*x^y, x]"},
			&SameTest{"True", "PolynomialQ[2*y, y]"},
			&SameTest{"True", "PolynomialQ[y*x, y]"},
			&SameTest{"True", "PolynomialQ[y*x, z]"},
			&SameTest{"True", "PolynomialQ[y*Sin[x], z]"},
			&SameTest{"False", "PolynomialQ[y^x, y]"},
			&SameTest{"False", "PolynomialQ[x^y, y]"},
			&SameTest{"True", "PolynomialQ[x^y, z]"},
			&SameTest{"True", "PolynomialQ[2.5*x^2, 2.5]"},
			&SameTest{"True", "PolynomialQ[2.5*x, 2.5]"},
			&SameTest{"False", "PolynomialQ[2*x^2, 2]"},
			&SameTest{"True", "PolynomialQ[2*x, 2]"},
			&SameTest{"True", "PolynomialQ[x, 2]"},
			&SameTest{"True", "PolynomialQ[Cos[x*y], Cos[x*y]]"},
			&SameTest{"True", "PolynomialQ[x^2,2.]"},
			&SameTest{"False", "PolynomialQ[x^a,a]"},
			&SameTest{"False", "PolynomialQ[x^n,x]"},
			&SameTest{"True", "PolynomialQ[-x*Cos[y],x]"},
			&SameTest{"True", "PolynomialQ[x^y, 1]"},
		},
		KnownFailures: []TestInstruction{
			&SameTest{"True", "PolynomialQ[2*x^2-3x+2, 2]"},
			&SameTest{"True", "PolynomialQ[2*x^2-3x, 2]"},
		},
	})
	defs = append(defs, Definition{
		Name:  "Exponent",
		Usage:  "`Exponent[p, var]` returns the degree of `p` with respect to the variable `var`.",
		SimpleExamples: []TestInstruction{
			&SameTest{"{0,3,5}", "Exponent[3 + x^3 + k*x^5, x, List]"},
		},
		Tests: []TestInstruction{
			// Sorting of the input addition expression is off here, so we sort
			// the result so it does match up.
			&SameTest{"{0,3,5,x^x}", "Exponent[3 + \"hello\" + x^3 + a*x^5 + x^x^x, x, List]//Sort"},
			&SameTest{"{0}", "Exponent[1 + x^x^x, x^x, List]"},
			&SameTest{"{0}", "Exponent[2 + a, x, List]"},
			&SameTest{"{0}", "Exponent[a, x, List]"},
			&SameTest{"{2}", "Exponent[x^2, x, List]"},
			&SameTest{"{1}", "Exponent[x^2 - x*(a + x), x, List]"},
			&SameTest{"{0,1}", "Exponent[(1 + x)/(3 + x), x, List]"},
			&SameTest{"{0,2}", "Exponent[(1 + x^2)/(3 + x), x, List]"},
			&SameTest{"{0,1}", "Exponent[(1 + x)/(3 + x^3), x, List]"},
			&SameTest{"{0}", "Exponent[(3 + x^3)^(-1), x, List]"},
			&SameTest{"{-3}", "Exponent[x^(-3), x, List]"},
			&SameTest{"{-3}", "Exponent[a/x^3, x, List]"},
			&SameTest{"{-2}", "Exponent[x^(-2), x, List]"},
			&SameTest{"{1}", "Exponent[(a*x)/(3 + x^3), x, List]"},
			&SameTest{"{0,1}", "Exponent[1 + b*x + x^2 - (x*(1 + a*x))/a, x, List]"},
			&SameTest{"{0,1}", "Exponent[1 + x + x^2 - (x*(1 + 2*x))/2, x, List]"},
		},
		KnownFailures: []TestInstruction{
			&SameTest{"{0,1}", "Exponent[1 + x^x^x, x^x^x, List]"},
		},
	})
	defs = append(defs, Definition{
		Name:  "Coefficient",
		Usage:  "`Coefficient[p, form]` returns the coefficient of form `form` in polynomial `p`.",
		SimpleExamples: []TestInstruction{
			&SameTest{"3", "Coefficient[(a + b)^3, a*b^2]"},
		},
		Tests: []TestInstruction{
			&SameTest{"j", "Coefficient[c + j*a + k*b, a]"},
			&SameTest{"a", "Coefficient[c + k*x + a*x^3, x, 3]"},
			&SameTest{"24", "Coefficient[2*b*(2*a + 3*b)*(1 + 2*a + 3*b), a*b^2]"},
			&SameTest{"29", "Coefficient[(2 + x)^2 + (5 + x)^2, x, 0]"},
			&SameTest{"1", "Coefficient[a + x, x]"},
			&SameTest{"4", "Coefficient[2*b*(2*a + 3*b)*(1 + 2*a + 3*b), a*b]"},
			&SameTest{"1", "Coefficient[x^2, x^2]"},
			&SameTest{"-a", "Coefficient[x^2 - x*(a + x), x]"},
			&SameTest{"-(1/a)+b", "Coefficient[1 + b*x + x^2 - (x*(1 + a*x))/a, x]"},
			&SameTest{"1/2", "Coefficient[1 + x + x^2 - (x*(1 + 2*x))/2, x]"},
		},
	})
	defs = append(defs, Definition{
		Name:  "PolynomialQuotientRemainder",
		Usage:  "`PolynomialQuotientRemainder[poly_, div_, var_]` returns the quotient and remainder of `poly` divided by `div` treating `var` as the polynomial variable.",
		SimpleExamples: []TestInstruction{
			&SameTest{"{x^2/2,2}", "PolynomialQuotientRemainder[2 + x^2 + x^3, 2 + 2*x, x]"},
			&SameTest{"{x^2-x y+y^2,-y^3}", "PolynomialQuotientRemainder[x^3, x + y, x]"},
			&SameTest{"{x/a,1-x/a}", "PolynomialQuotientRemainder[1 + x^3, 1 + a*x^2, x]"},
		},
	})
	defs = append(defs, Definition{
		Name:  "PolynomialQuotient",
		Usage:  "`PolynomialQuotient[poly_, div_, var_]` returns the quotient of `poly` divided by `div` treating `var` as the polynomial variable.",
		SimpleExamples: []TestInstruction{
			&SameTest{"x^2/2", "PolynomialQuotient[2 + x^2 + x^3, 2 + 2*x, x]"},
			&SameTest{"x^2-x y+y^2", "PolynomialQuotient[x^3, x + y, x]"},
			&SameTest{"x/a", "PolynomialQuotient[1 + x^3, 1 + a*x^2, x]"},
		},
	})
	defs = append(defs, Definition{
		Name:  "PolynomialRemainder",
		Usage:  "`PolynomialRemainder[poly_, div_, var_]` returns the remainder of `poly` divided by `div` treating `var` as the polynomial variable.",
		SimpleExamples: []TestInstruction{
			&SameTest{"2", "PolynomialRemainder[2 + x^2 + x^3, 2 + 2*x, x]"},
			&SameTest{"-y^3", "PolynomialRemainder[x^3, x + y, x]"},
			&SameTest{"1-x/a", "PolynomialRemainder[1 + x^3, 1 + a*x^2, x]"},
		},
	})
	defs = append(defs, Definition{
		Name:  "FactorTermsList",
		Usage:  "`FactorTermsList[expr]` factors out the constant term of `expr` and puts the factored result into a `List`.",
		SimpleExamples: []TestInstruction{
			&SameTest{"{2,Sin[8 k]}", "FactorTermsList[2*Sin[8*k]]"},
			&SameTest{"{1/2,a+x}", "FactorTermsList[a/2 + x/2]"},
			&SameTest{"{1,a+x}", "FactorTermsList[a + x]"},
		},
		Tests: []TestInstruction{
			&SameTest{"{1,1}", "FactorTermsList[1]"},
			&SameTest{"{5,1}", "FactorTermsList[5]"},
			&SameTest{"{5.,1}", "FactorTermsList[5.]"},
			&SameTest{"{1,a}", "FactorTermsList[a]"},
			&SameTest{"{1/2,a}", "FactorTermsList[a/2]"},
			&SameTest{"{-(3/2),x}", "FactorTermsList[(-3*x)/2]"},
			&SameTest{"{2,a+x}", "FactorTermsList[2*a + 2*x]"},
			&SameTest{"{1/2,a/(2 b+2 c)+x/(2 b+2 c)}", "FactorTermsList[(a/2 + x/2)/(2*b + 2*c)]"},
			&SameTest{"{1,2+x^2}", "FactorTermsList[(8 + 4*x^2)/4]"},
			&SameTest{"{-(1/2),2+3 x+x^2}", "FactorTermsList[(-4 - 6*x - 2*x^2)/4]"},
			&SameTest{"{-(1/2),-2+3 x+x^2}", "FactorTermsList[(2 - 3*x - x^2)/2]"},
			&SameTest{"{-(1/2),-2-3 x+x^2}", "FactorTermsList[(2 + 3*x - x^2)/2]"},
			&SameTest{"{1/2,2+3 x+x^2}", "FactorTermsList[(2 + 3*x + x^2)/2]"},
			&SameTest{"{1/2,-2-3 x+x^2}", "FactorTermsList[(-2 - 3*x + x^2)/2]"},
			&SameTest{"{5,2+3 x+x^2}", "FactorTermsList[5*(1 + x)*(2 + x)]"},
			&SameTest{"{40,1+3 x+3 x^2+x^3}", "FactorTermsList[5*(2 + 2*x)^3]"},
			&SameTest{"{-6,1+x}", "FactorTermsList[(-12 - 12*x)/2]"},
			&SameTest{"{2/3,3+x}", "FactorTermsList[(18 + 6*x)/9]"},
			&SameTest{"{-(2800301/67344500),1-2 x+x^3}", "FactorTermsList[(-2800301/538756 + (2800301*x)/269378 - (2800301*x^3)/538756)/125]"},
		},
	})
	/*defs = append(defs, Definition{
		Name:  "PolynomialGCD",
		Usage:  "`PolynomialGCD[a, b]` calculates the polynomial GCD of `a` and `b`.",
		SimpleExamples: []TestInstruction{
			&SameTest{"2+3 x+x^2", "PolynomialGCD[8 + 22*x + 21*x^2 + 8*x^3 + x^4, 6 + 11*x + 6*x^2 + x^3]"},
			&SameTest{"2+x^2", "PolynomialGCD[-4 + x^4, 4 + 4*x^2 + x^4]"},
			&SameTest{"1-2 x+x^3", "PolynomialGCD[5 - 12*x + 4*x^2 + 5*x^3 - x^4 - 2*x^5 + x^7, 3 - 6*x - 7*x^2 + 17*x^3 + x^4 - 9*x^5 + x^7]"},
			&SameTest{"1+x", "PolynomialGCD[6 + 7*x + x^2, -6 - 5*x + x^2]"},
			&SameTest{"1", "PolynomialGCD[3 + 6*x + 2*x^2, 1 + 2*x]"},
			&SameTest{"3+x", "PolynomialGCD[6 - 28*x - 19*x^2 + 3*x^3 + 2*x^4, -18 - 9*x + 2*x^2 + x^3]"},
		},
	})*/
	return
}
