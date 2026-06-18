package main

import (
	"dt_go/he"
	"dt_go/tree"
	"dt_go/treeio"
	"fmt"
	"math"
	"flag"
)

func main() {
	MODE := "tree" // "parse" | "boosting" | "obo" | "bsgs" | "tree"
	flag.StringVar(&MODE, "mode", MODE, "execution mode: parse | boosting | obo | bsgs | tree")
	flag.StringVar(&MODE, "m", MODE, "shorthand for -mode")
	flag.Parse()

	switch MODE {
	case "parse", "boosting", "obo", "bsgs", "tree":
	default:
		fmt.Printf("invalid mode: %s\n", MODE)
		fmt.Println("available modes: parse | boosting | obo | bsgs | tree")
		return
	}
	loop := 1
	DIR := "xgbdata"

	clv := []int{6, 7, 8, 9, 10, 10, 11, 11, 11, 11, 12, 12, 12, 12, 12, 12, 12, 12, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13}
	params := 0
	K := 8
	T := 8
	P := 32
	D := 12
	TOL := 1e-3
	short := false // make to true for short check
	slots := 1 << 15
	if short {
		slots = 1 << 12
	}

	GEN := treeio.GenConfig{
		Seed:        0,
		N:           slots,
		F:           1,
		NumTrees:    T,
		Depth:       1,
		SparseRatio: 1.0,

		BaseScore: 0.5,

		XMean:   0,
		XStd:    1,
		NanProb: 0,

		ThrMean: 0,
		ThrStd:  1,
		LeafStd: 1,
	}
	Models := []string{"d8", "d10", "d12"}
	dest := "results/" + MODE + ".txt"
	fmt.Println("Mode : ",MODE, ", Result dest : ",dest)
	for L := 0; L < loop; L++ {
		if MODE == "parse" {
			for _, Model := range Models {
				bundle, d := treeio.ParseBundle(DIR, Model)

				fmt.Printf("[BUNDLE] mode=%s N=%d F=%d trees=%d base_score=%.6f slots=%d\n",
					MODE, bundle.Input.N, bundle.Input.F, len(bundle.Model.Trees), bundle.Model.BaseScore, slots)

				he.Preset(short, params, clv[d]) 

				hook := func(e tree.ProgressEvent) { _ = e }
				r := runOnce(bundle, slots, K, hook, TOL, MODE, P)
				PrintSummary([]RunResult{r}, slots, dest)
			}
		} else {
			for d := 2; d <= D; d++ {
				fmt.Printf("Loop : %d, Mode : %s, Depth : %d\n",L,MODE,d)
				if MODE == "obo" || MODE == "tree" || MODE == "bsgs" {
					K = 1
					T = 1
					params = 1 //ckks parameter for non-batch bootstrapping
				} else {
					K = 8
					T = 8
					params = 0 //ckks parameter for batch bootstrapping
				}
				GEN.Depth = d
				GEN.NumTrees = T
				bundle := treeio.GenBundle(GEN)
				fmt.Printf("[BUNDLE] mode=%s N=%d F=%d trees=%d base_score=%.6f slots=%d\n",
					MODE, bundle.Input.N, bundle.Input.F, len(bundle.Model.Trees), bundle.Model.BaseScore, slots)

				he.Preset(short, params, clv[d]-int(math.Log2(float64(32/P))))

				hook := func(e tree.ProgressEvent) { _ = e }
				r := runOnce(bundle, slots, K, hook, TOL, MODE, P)
				PrintSummary([]RunResult{r}, slots, dest)
			}
		}
	}
}
