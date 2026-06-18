package he

// ---------- const arithmetic ----------

func addConst(a, b *Scalar) *Scalar {
	if a != nil && b != nil && a.IsInt() && b.IsInt() {
		x := *a.I + *b.I
		return &Scalar{I: &x}
	}
	x := a.AsFloat64() + b.AsFloat64()
	return &Scalar{F: &x}
}

func subConst(a, b *Scalar) *Scalar {
	if a != nil && b != nil && a.IsInt() && b.IsInt() {
		x := *a.I - *b.I
		return &Scalar{I: &x}
	}
	x := a.AsFloat64() - b.AsFloat64()
	return &Scalar{F: &x}
}

func mulConst(a, b *Scalar) *Scalar {
	if a != nil && b != nil && a.IsInt() && b.IsInt() {
		x := (*a.I) * (*b.I)
		return &Scalar{I: &x}
	}
	x := a.AsFloat64() * b.AsFloat64()
	return &Scalar{F: &x}
}

func ConjugateNew(a CT) CT {
	if a.K.IsInt() {
		return KInt(*a.K.I)
	}
	if a.K.IsFloat() {
		return KFloat(*a.K.F)
	}
	ret, _ := Eval.ConjugateNew(a.Ct)
	return FromCt(ret)
}

// ---------- add ----------

func AddNew(a, b CT) CT {
	// const/const
	if a.IsConst() && b.IsConst() {
		return CT{K: addConst(a.K, b.K)}
	}
	// short-circuit 0
	if a.ConstIsZero() {
		return b
	}
	if b.ConstIsZero() {
		return a
	}
	// ct + const
	if a.IsCt() && b.IsConst() {
		ct, _ := Eval.AddNew(a.Ct, b.K.Operand())
		return FromCt(ct)
	}
	// const + ct
	if a.IsConst() && b.IsCt() {
		ct, _ := Eval.AddNew(b.Ct, a.K.Operand())
		return FromCt(ct)
	}
	// ct + ct
	ct, _ := Eval.AddNew(a.Ct, b.Ct)
	return FromCt(ct)
}

func Add(a, b CT, out *CT) {
	*out = AddNew(a, b)
}

// ---------- sub ----------

func SubNew(a, b CT) CT {
	if a.IsConst() && b.IsConst() {
		return CT{K: subConst(a.K, b.K)}
	}
	if b.ConstIsZero() {
		return a
	}
	// ct - const
	if a.IsCt() && b.IsConst() {
		ct, _ := Eval.SubNew(a.Ct, b.K.Operand())
		return FromCt(ct)
	}
	if a.IsConst() && b.IsCt() {
		neg, _ := Eval.MulNew(b.Ct, int64(-1))
		ct, _ := Eval.AddNew(neg, a.K.Operand())
		return FromCt(ct)
	}
	// ct - ct
	ct, _ := Eval.SubNew(a.Ct, b.Ct)
	return FromCt(ct)
}

func Sub(a, b CT, out *CT) {
	*out = SubNew(a, b)
}

// ---------- mul ----------

func MulNew(a, b CT, lazy ...bool) CT {
	// const/const
	if a.IsConst() && b.IsConst() {
		return CT{K: mulConst(a.K, b.K)}
	}
	// short-circuit 0/1
	if a.ConstIsZero() || b.ConstIsZero() {
		return KInt(0)
	}
	if a.ConstIsOne() {
		return b
	}
	if b.ConstIsOne() {
		return a
	}
	// ct * const
	if a.IsCt() && b.IsConst() {
		ct, _ := Eval.MulNew(a.Ct, b.K.Operand())
		return FromCt(ct)
	}
	// const * ct
	if a.IsConst() && b.IsCt() {
		ct, _ := Eval.MulNew(b.Ct, a.K.Operand())
		return FromCt(ct)
	}
	// ct * ct
	ct, _ := Eval.MulNew(a.Ct, b.Ct)
	return FromCt(ct)
}

func Mul(a, b CT, out *CT) {
	*out = MulNew(a, b)
}

// ---------- mulrelin ----------

func MulRelinNew(a, b CT) CT {
	// const/const
	if a.IsConst() && b.IsConst() {
		return CT{K: mulConst(a.K, b.K)}
	}
	// short-circuit 0/1
	if a.ConstIsZero() || b.ConstIsZero() {
		return KInt(0)
	}
	if a.ConstIsOne() {
		return b
	}
	if b.ConstIsOne() {
		return a
	}
	if a.IsCt() && b.IsConst() {
		ct, _ := Eval.MulNew(a.Ct, b.K.Operand())
		return FromCt(ct)
	}
	if a.IsConst() && b.IsCt() {
		ct, _ := Eval.MulNew(b.Ct, a.K.Operand())
		return FromCt(ct)
	}
	// ct * ct
	ct, _ := Eval.MulRelinNew(a.Ct, b.Ct)
	return FromCt(ct)
}

func MulRelin(a, b CT, out *CT) {
	*out = MulRelinNew(a, b)
}

// ---------- rescale / relin ----------

func Rescale(x CT) {
	if x.IsConst() {
		return
	}
	Eval.Rescale(x.Ct, x.Ct)
}

func RelinNew(x CT) CT {
	if x.IsConst() {
		return x
	}
	ct, _ := Eval.RelinearizeNew(x.Ct)
	return FromCt(ct)
}

func Relin(x CT) {
	if x.IsConst() {
		return
	}
	Eval.Relinearize(x.Ct, x.Ct)
}

// ---------- helpers ----------

func IsConstZero(x CT) bool { return x.ConstIsZero() }
func IsConstOne(x CT) bool  { return x.ConstIsOne() }
