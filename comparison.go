package cas

import "bytes"

func (this *Expression) EvalEqual(es *EvalState) Ex {
	if len(this.Parts) != 3 {
		return this
	}

	var isequal string = this.Parts[1].Eval(es).IsEqual(this.Parts[2].Eval(es), es)
	if isequal == "EQUAL_UNK" {
		return this
	} else if isequal == "EQUAL_TRUE" {
		return &Symbol{"True"}
	} else if isequal == "EQUAL_FALSE" {
		return &Symbol{"False"}
	}

	return &Error{"Unexpected equality return value."}
}

func (this *Expression) ToStringEqual() string {
	var buffer bytes.Buffer
	buffer.WriteString("(")
	buffer.WriteString(this.Parts[1].ToString())
	buffer.WriteString(") == (")
	buffer.WriteString(this.Parts[2].ToString())
	buffer.WriteString(")")
	return buffer.String()
}

type SameQ struct {
	Lhs Ex
	Rhs Ex
}

func (this *SameQ) Eval(es *EvalState) Ex {
	var issame bool = this.Lhs.Eval(es).IsSameQ(this.Rhs.Eval(es), es)
	if issame {
		return &Symbol{"True"}
	} else {
		return &Symbol{"False"}
	}
}

func (this *SameQ) Replace(r *Rule, es *EvalState) Ex {
	if this.IsMatchQ(r.Lhs, es) {
		return r.Rhs
	}
	this.Lhs = this.Lhs.Replace(r, es)
	this.Rhs = this.Rhs.Replace(r, es)
	return this.Eval(es)
}

func (this *SameQ) ToString() string {
	var buffer bytes.Buffer
	buffer.WriteString("(")
	buffer.WriteString(this.Lhs.ToString())
	buffer.WriteString(") === (")
	buffer.WriteString(this.Rhs.ToString())
	buffer.WriteString(")")
	return buffer.String()
}

func (this *SameQ) IsEqual(otherEx Ex, es *EvalState) string {
	return "EQUAL_UNK"
}

func (this *SameQ) IsSameQ(otherEx Ex, es *EvalState) bool {
	return false
}

func (this *SameQ) IsMatchQ(otherEx Ex, es *EvalState) bool {
	return this.IsSameQ(otherEx, es)
}

func (this *SameQ) DeepCopy() Ex {
	return &SameQ{
		this.Lhs.DeepCopy(),
		this.Rhs.DeepCopy(),
	}
}

type MatchQ struct {
	Expr Ex
	Form Ex
}

func (this *MatchQ) Eval(es *EvalState) Ex {
	oldVars := es.GetDefinedSnapshot()
	var issame bool = this.Expr.Eval(es).IsMatchQ(this.Form.Eval(es), es)
	es.ClearPD()
	es.defined = oldVars
	if issame {
		return &Symbol{"True"}
	} else {
		return &Symbol{"False"}
	}
}

func (this *MatchQ) Replace(r *Rule, es *EvalState) Ex {
	if this.IsMatchQ(r.Lhs, es) {
		return r.Rhs
	}
	this.Expr = this.Expr.Replace(r, es)
	this.Form = this.Form.Replace(r, es)
	return this.Eval(es)
}

func (this *MatchQ) ToString() string {
	var buffer bytes.Buffer
	buffer.WriteString("MatchQ[")
	buffer.WriteString(this.Expr.ToString())
	buffer.WriteString(", ")
	buffer.WriteString(this.Form.ToString())
	buffer.WriteString("]")
	return buffer.String()
}

func (this *MatchQ) IsEqual(otherEx Ex, es *EvalState) string {
	return "EQUAL_UNK"
}

func (this *MatchQ) IsSameQ(otherEx Ex, es *EvalState) bool {
	return false
}

func (this *MatchQ) IsMatchQ(otherEx Ex, es *EvalState) bool {
	return this.IsSameQ(otherEx, es)
}

func (this *MatchQ) DeepCopy() Ex {
	return &MatchQ{
		this.Expr.DeepCopy(),
		this.Form.DeepCopy(),
	}
}
