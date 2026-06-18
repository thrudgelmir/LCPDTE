package he

import (
	"fmt"
	"math"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

func Encrypt(values []complex128, level ...int) *rlwe.Ciphertext {
	lvl := MaxLv
	//fmt.Println(lvl)
	if len(level) > 0 {
		lvl = level[0]
	}
	pt := ckks.NewPlaintext(Params, lvl)
	must(Encoder.Encode(values, pt))
	ct, err := Encryptor.EncryptNew(pt)
	must(err)
	return ct
}

func Zeros(level ...int) *rlwe.Ciphertext {
	lvl := MaxLv
	if len(level) > 0 {
		lvl = level[0]
	}
	pt := ckks.NewPlaintext(Params, lvl)
	values := make([]complex128, Params.MaxSlots())
	must(Encoder.Encode(values, pt))
	ct, err := Encryptor.EncryptNew(pt)
	must(err)
	return ct
}

func PrintCt(x CT, slots []int, r bool) {
	if x.IsConst() {
		if x.K.IsInt() {
			fmt.Printf("(%d)", *x.K.I)
		} else {
			fmt.Printf("(%.6f)", *x.K.F)
		}
		return
	}
	for _, s := range slots {
		if r {
			fmt.Printf("(%.6f)", real(Decrypt(x.Ct)[s]))
		} else {
			fmt.Printf("(%.6f, %.6f)", real(Decrypt(x.Ct)[s]), imag(Decrypt(x.Ct)[s]))
		}
	}
}

func ConstCT(v int) CT {
	x := int64(v)
	return CT{Ct: nil, K: &Scalar{I: &x}}
}

func Decrypt(ct *rlwe.Ciphertext) []complex128 {
	slots := ct.Slots()
	if !ct.IsBatched {
		slots *= 2
	}
	out := make([]complex128, slots)
	must(Encoder.Decode(Decryptor.DecryptNew(ct), out))
	return out
}

func Show(c *rlwe.Ciphertext, k ...int) {
	kk := 10
	pp := 2
	isComplex := false

	if len(k) >= 1 {
		kk = k[0]
	}
	if len(k) >= 2 {
		pp = k[1]
	}
	if len(k) >= 3 {
		isComplex = (k[2] != 0)
	}

	vals := Decrypt(c)
	if kk > len(vals) {
		kk = len(vals)
	}
	if kk < 0 {
		kk = 0
	}

	scale := math.Pow10(pp)
	round := func(x float64) float64 {
		return math.Round(x*scale) / scale
	}

	if !isComplex {
		fmt.Print("[")
		for i := 0; i < kk; i++ {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%.*f", pp, round(real(vals[i])))
		}
		fmt.Println("]")
		return
	}

	fmt.Print("[")
	for i := 0; i < kk; i++ {
		if i > 0 {
			fmt.Print(", ")
		}
		re := round(real(vals[i]))
		im := round(imag(vals[i]))
		fmt.Printf("(%.*f%+.*fi)", pp, re, pp, im)
	}
	fmt.Println("]")
}
