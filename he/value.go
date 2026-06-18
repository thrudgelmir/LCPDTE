package he

import "github.com/tuneinsight/lattigo/v6/core/rlwe"

type CT struct {
	Ct *rlwe.Ciphertext
	K  *Scalar
}

type Scalar struct {
	I *int64
	F *float64
}

func KInt(x int64) CT               { return CT{Ct: nil, K: &Scalar{I: &x}} }
func KFloat(x float64) CT           { return CT{Ct: nil, K: &Scalar{F: &x}} }
func FromCt(ct *rlwe.Ciphertext) CT { return CT{Ct: ct, K: nil} }
func (x CT) IsNil() bool            { return x.Ct == nil && x.K == nil }

func (s *Scalar) IsInt() bool   { return s != nil && s.I != nil }
func (s *Scalar) IsFloat() bool { return s != nil && s.F != nil }

func (s *Scalar) AsFloat64() float64 {
	if s == nil {
		return 0
	}
	if s.F != nil {
		return *s.F
	}
	if s.I != nil {
		return float64(*s.I)
	}
	return 0
}

func (s *Scalar) IsZero() bool { return s.AsFloat64() == 0 }
func (s *Scalar) IsOne() bool  { return s.AsFloat64() == 1 }

func (s *Scalar) Operand() rlwe.Operand {
	if s == nil {
		return int64(0)
	}
	if s.I != nil {
		return *s.I
	}
	if s.F != nil {
		return *s.F
	}
	return int64(0)
}

func (x CT) IsConst() bool { return x.K != nil && x.Ct == nil }
func (x CT) IsCt() bool    { return x.Ct != nil && x.K == nil }

func (x CT) ConstIsZero() bool { return x.IsConst() && x.K.IsZero() }
func (x CT) ConstIsOne() bool  { return x.IsConst() && x.K.IsOne() }
