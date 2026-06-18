package he

import (
	"fmt"
	"math"
	"time"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/tuneinsight/lattigo/v6/utils/bignum"
)

func Bootstrap(ct *rlwe.Ciphertext, isComplex bool) *rlwe.Ciphertext {
	var err error

	ct, err = Eval.SlotsToCoeffs(ct, nil)
	must(err)

	ct, _, err = Eval.ScaleDown(ct)
	must(err)

	ct, err = Eval.ModUp(ct)
	must(err)

	real, imag, err := Eval.CoeffsToSlots(ct)
	must(err)

	real, err = Eval.EvalMod(real)
	must(err)

	if !isComplex {
		return real
	}

	imag, err = Eval.EvalMod(imag)
	must(err)

	must(Eval.Evaluator.Mul(imag, 1i, imag))
	out := real.CopyNew()
	must(Eval.Evaluator.Add(real, imag, out))
	return out
}

func Bootstrap2(ct *rlwe.Ciphertext) *rlwe.Ciphertext {
	data := Decrypt(ct)

	for i := range data {
		r := math.Round(real(data[i]))
		im := math.Round(imag(data[i]))
		data[i] = complex(r, im)
	}
	ret := Encrypt(data, Params.MaxLevel()-1)
	return ret
}

func MultiBootstrap(cts []*rlwe.Ciphertext) []*rlwe.Ciphertext {
	out := make([]*rlwe.Ciphertext, len(cts))
	for i := range cts {
		//out[i] = Bootstrap(cts[i], isComplex)
		out[i] = Bootstrap2(cts[i])
	}
	return out
}

func MultiBootstrap_FeatureTree(cts []*rlwe.Ciphertext, numbatch int) []*rlwe.Ciphertext {
	out := make([]*rlwe.Ciphertext, len(cts))
	var err error

	if numbatch > 2*Logbase {
		panic("error : numbatch > 2*logbase")
	}

	//value := make([]complex128, Params.MaxSlots())
	//for i := 0; i < numbatch; i++ {
	//	printDebug(Params, cts[i], value, Decryptor, Encoder)
	//}

	if Logbase >= numbatch {
		//fmt.Println("real btp")
		if out, err = Batched_binBootstrap_Notcomplex(cts); err != nil {
			fmt.Println(err)
		}

	} else {
		//fmt.Println("complex btp")
		ciphertext_real := make([]*rlwe.Ciphertext, (numbatch+1)/2)
		ciphertext_imag := make([]*rlwe.Ciphertext, (numbatch)/2)
		for i := 0; i < (numbatch+1)/2; i++ {
			ciphertext_real[i] = cts[i]
			if i+(numbatch+1)/2 == numbatch {
				continue
			}
			ciphertext_imag[i] = cts[i+(numbatch+1)/2]
		}

		if ciphertext_real, ciphertext_imag, err = Batched_binBootstrap_complex(ciphertext_real, ciphertext_imag, true); err != nil {
			fmt.Println(err)
		}

		for i := 0; i < (numbatch+1)/2; i++ {
			out[i] = ciphertext_real[i]
			if i+(numbatch+1)/2 == numbatch {
				continue
			}
			out[i+(numbatch+1)/2] = ciphertext_imag[i]
		}
	}

	//value = make([]complex128, Params.MaxSlots())
	//for i := 0; i < len(out); i++ {
	//	printDebug(Params, out[i], value, Decryptor, Encoder)
	//}

	return out
}

func MultiBootstrapCT(xs []CT) []CT {
	cts := make([]*rlwe.Ciphertext, 0, len(xs))
	pos := make([]int, 0, len(xs))

	for i := range xs {
		if xs[i].IsConst() {
			continue
		}
		if xs[i].IsCt() {
			pos = append(pos, i)
			cts = append(cts, xs[i].Ct)
		}
	}

	if len(cts) == 0 {
		return xs
	}

	//fmt.Println("# mult cts : ", len(cts), "level :", cts[0].Level())

	//bts := MultiBootstrap(cts)
	bts := MultiBootstrap_FeatureTree(cts, len(cts))

	for j, i := range pos {
		xs[i] = FromCt(bts[j])
	}

	return xs
}

func MultiBootstrap_Debug_FeatureTree(xs []CT) []CT {
	cts := make([]*rlwe.Ciphertext, 0, len(xs))
	pos := make([]int, 0, len(xs))

	for i := range xs {
		if xs[i].IsConst() {
			continue
		}
		if xs[i].IsCt() {
			pos = append(pos, i)
			cts = append(cts, xs[i].Ct)
		}
	}

	if len(cts) == 0 {
		return xs
	}

	fmt.Println("# of ciphertext : ", len(cts))

	//bts := MultiBootstrap(cts, isComplex)
	fmt.Println("bts before : ", cts[0].Level())
	bts := MultiBootstrap_FeatureTree(cts, len(cts))
	fmt.Println("bts after : ", bts[0].Level())
	for j, i := range pos {
		xs[i] = FromCt(bts[j])
	}

	return xs
}

func DiBootstrap(ciphertext *rlwe.Ciphertext, LookUpTable func(int) int, iscomplex bool) (ctOut *rlwe.Ciphertext, ctOut2 *rlwe.Ciphertext, err error) {
	real_list, imag_list, err := DiBootstrapMany(ciphertext, []func(int) int{LookUpTable}, iscomplex)
	if err != nil {
		panic(err)
	}

	if iscomplex == false {
		return real_list[0], nil, err
	}
	return real_list[0], imag_list[0], err
}

func DiBootstrapMany(ciphertext *rlwe.Ciphertext, LookUpTable []func(int) int, iscomplex bool) (ctOut []*rlwe.Ciphertext, ctOut2 []*rlwe.Ciphertext, err error) {

	params := Params
	eval := Eval
	base := Base
	hermite_order := HermiteOrder
	//eval := Eval

	debug := false

	stc_start := time.Now()
	if ciphertext, err = eval.SlotsToCoeffs(ciphertext, nil); err != nil {
		panic(err)
	}
	stc_time := time.Since(stc_start)
	if debug {
		fmt.Println("StC time : ", stc_time)
	}

	// Step 3: scale to q/|m|
	if ciphertext, _, err = eval.ScaleDown(ciphertext); err != nil {
		panic(err)
	}

	// Step 4 : Extend the basis from q to Q
	if ciphertext, err = eval.ModUp(ciphertext); err != nil {
		panic(err)
	}

	// Step 5 : CoeffsToSlots (Homomorphic encoding)
	// Note: expects the result to be given in bit-reversed order
	// Also, we need the homomorphic encoding to split the real and
	// imaginary parts into two pure real ciphertexts, because the
	// homomorphic modular reduction is only defined on the reals.
	// The `imag` ciphertext can be ignored if the original input
	// is purely real.
	var real, imag *rlwe.Ciphertext
	var real_real, real_imag, imag_real, imag_imag *rlwe.Ciphertext
	var ciphertext_real, ciphertext_imag *rlwe.Ciphertext

	cts_start := time.Now()
	if real, imag, err = eval.CoeffsToSlots(ciphertext); err != nil {
		panic(err)
	}
	cts_time := time.Since(cts_start)
	if debug {
		fmt.Println("CtS time : ", cts_time)
	}

	// Step 6 : EvalMod (Homomorphic modular reduction)

	// compute real part
	Exp_start := time.Now()
	if real_imag, err = eval.EvalSinAndScale(real, 2*math.Pi); err != nil {
		panic(err)
	}

	if real_real, err = eval.EvalCosAndScale(real, 2*math.Pi); err != nil {
		panic(err)
	}

	// Recombines the real and imaginary part
	if err = eval.Evaluator.Mul(real_imag, 1i, real_imag); err != nil {
		panic(err)
	}

	if ciphertext_real, err = eval.Evaluator.AddNew(real_real, real_imag); err != nil {
		panic(err)
	}

	Exp_time := time.Since(Exp_start)
	if debug {
		fmt.Println("Exp time : ", Exp_time)
	}

	// compute imag part
	if iscomplex == true {
		if imag_imag, err = eval.EvalSinAndScale(imag, 2*math.Pi); err != nil {
			panic(err)
		}

		if imag_real, err = eval.EvalCosAndScale(imag, 2*math.Pi); err != nil {
			panic(err)
		}

		// Recombines the real and imaginary part
		if err = eval.Evaluator.Mul(imag_imag, 1i, imag_imag); err != nil {
			panic(err)
		}

		if ciphertext_imag, err = eval.Evaluator.AddNew(imag_real, imag_imag); err != nil {
			panic(err)
		}
	}

	polyEval := polynomial.NewEvaluator(params, eval.Evaluator)
	poly_list := make([]polynomial.Polynomial, len(LookUpTable))
	eval_poly_list := make([]bignum.Polynomial, len(LookUpTable))

	for i := 0; i < len(LookUpTable); i++ {
		eval_poly_list[i] = bignum.NewPolynomial(0, HermiteInterpolation(base, hermite_order, LookUpTable[i]), nil)
	}

	for i := 0; i < len(LookUpTable); i++ {
		if poly_list[i] = polynomial.NewPolynomial(eval_poly_list[i]); err != nil {
			panic(err)
		}
	}

	p_slice := make([]interface{}, len(LookUpTable))
	for i := 0; i < len(p_slice); i++ {
		p_slice[i] = poly_list[i]
	}

	LUT_start := time.Now()
	ctOut, _ = polyEval.EvaluateMultiPoly(ciphertext_real, p_slice, params.DefaultScale())
	LUT_time := time.Since(LUT_start)
	if debug {
		fmt.Println("LUT time : ", LUT_time)
	}
	if iscomplex == true {
		ctOut2, _ = polyEval.EvaluateMultiPoly(ciphertext_imag, p_slice, params.DefaultScale())
	}

	if iscomplex == false {
		return ctOut, nil, err
	}

	return ctOut, ctOut2, err
}

func printDebug(params ckks.Parameters, ciphertext *rlwe.Ciphertext, valuesWant []complex128, decryptor *rlwe.Decryptor, encoder *ckks.Encoder) (valuesTest []complex128) {

	slots := ciphertext.Slots()

	if !ciphertext.IsBatched {
		slots *= 2
	}

	valuesTest = make([]complex128, slots)

	if err := encoder.Decode(decryptor.DecryptNew(ciphertext), valuesTest); err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Printf("Level: %d (logQ = %d)\n", ciphertext.Level(), params.LogQLvl(ciphertext.Level()))

	fmt.Printf("Scale: 2^%f\n", math.Log2(ciphertext.Scale.Float64()))
	fmt.Printf("ValuesTest: %10.14f %10.14f %10.14f %10.14f...\n", valuesTest[0], valuesTest[1], valuesTest[2], valuesTest[3])
	fmt.Printf("ValuesWant: %10.14f %10.14f %10.14f %10.14f...\n", valuesWant[0], valuesWant[1], valuesWant[2], valuesWant[3])

	min_L2_err := 100.0
	avg_L2_err := 0.0

	for i := 0; i < params.MaxSlots(); i++ {

		// compute real err
		real_round := math.Round(2 * real(valuesTest[i]))
		realr := real(2 * valuesTest[i])
		real_err := math.Abs(real_round - realr)

		// compute imag err
		imag_round := math.Round(imag(valuesTest[i]))
		imag := imag(valuesTest[i])
		imag_err := math.Abs(imag_round - imag)

		L2_err := math.Sqrt(real_err*real_err + imag_err*imag_err)
		log_L2_err := -1 * math.Log2(L2_err)

		avg_L2_err += log_L2_err
		min_L2_err = min(min_L2_err, log_L2_err)
		if log_L2_err == min_L2_err {
			//min_idx = i
		}
	}

	//fmt.Println("B_in :", math.Log2(B_in/float64(len(valuesTest))))
	fmt.Println("Avg Precision :", avg_L2_err/float64(len(valuesTest)))
	fmt.Println("Max Precision :", min_L2_err)

	//precStats := ckks.GetPrecisionStats(params, encoder, nil, valuesWant, valuesTest, 0, false)

	//fmt.Println(precStats.String())
	fmt.Println()

	return
}

func Batched_binBootstrap_Notcomplex(ciphertext_reallists []*rlwe.Ciphertext) (ctOut []*rlwe.Ciphertext, err error) {
	eval := Eval

	var ciphertext_real, ciphertext *rlwe.Ciphertext

	//for i := 0; i < logbase; i++ {
	for i := 0; i < len(ciphertext_reallists); i++ {
		if i == 0 {
			ciphertext_real = ciphertext_reallists[i].CopyNew()
			//ciphertext_imag = ciphertext_imaglists[i].CopyNew()
		} else {
			eval.Mul(ciphertext_reallists[i], 1<<i, ciphertext_reallists[i])
			eval.Add(ciphertext_reallists[i], ciphertext_real, ciphertext_real)

			//eval.Mul(ciphertext_imaglists[i], 1<<i, ciphertext_imaglists[i])
			//eval.Add(ciphertext_imaglists[i], ciphertext_imag, ciphertext_imag)
		}
	}

	//eval.Mul(ciphertext_imag, 1i, ciphertext_imag)
	//ciphertext, _ = eval.AddNew(ciphertext_real, ciphertext_imag)
	ciphertext = ciphertext_real.CopyNew()

	bit_extraction := make([]func(int) int, len(ciphertext_reallists))
	for i := 0; i < len(ciphertext_reallists); i++ {
		bit_extraction[i] = func(x int) int { return (x >> i) & 1 }
	}

	ciphertext_list, _, _ := DiBootstrapMany(ciphertext, bit_extraction, false)

	return ciphertext_list, err
}

func Batched_binBootstrap_complex(ciphertext_reallists []*rlwe.Ciphertext, ciphertext_imaglists []*rlwe.Ciphertext, iscomplex bool) (ctOut []*rlwe.Ciphertext, ctOut2 []*rlwe.Ciphertext, err error) {

	eval := Eval
	debug := true
	var ciphertext_real, ciphertext_imag, ciphertext *rlwe.Ciphertext

	//comb_start := time.Now()
	for i := 0; i < len(ciphertext_reallists); i++ {
		if i == 0 {
			ciphertext_real = ciphertext_reallists[i].CopyNew()
			ciphertext_imag = ciphertext_imaglists[i].CopyNew()
		} else {
			eval.Mul(ciphertext_reallists[i], 1<<i, ciphertext_reallists[i])
			eval.Add(ciphertext_reallists[i], ciphertext_real, ciphertext_real)
			if len(ciphertext_imaglists) == i {
				continue
			}
			eval.Mul(ciphertext_imaglists[i], 1<<i, ciphertext_imaglists[i])
			eval.Add(ciphertext_imaglists[i], ciphertext_imag, ciphertext_imag)
		}
	}

	eval.Mul(ciphertext_imag, 1i, ciphertext_imag)
	ciphertext, _ = eval.AddNew(ciphertext_real, ciphertext_imag)

	//comb_time := time.Since(comb_start)
	if debug {
		//fmt.Println("combine time : ", comb_time)
	}

	bit_extraction := make([]func(int) int, len(ciphertext_reallists))
	for i := 0; i < len(ciphertext_reallists); i++ {
		bit_extraction[i] = func(x int) int { return (x >> i) & 1 }
	}

	ctOut = make([]*rlwe.Ciphertext, len(ciphertext_reallists))
	ctOut2 = make([]*rlwe.Ciphertext, len(ciphertext_imaglists))

	ctOut, ctOut2, err = DiBootstrapMany(ciphertext, bit_extraction, true)
	//fmt.Println(ctOut[0].Level(), ctOut2[0].Level())

	return //ciphertext_list, ciphertext_list2, err
}
