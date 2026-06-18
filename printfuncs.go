package main

import (
	"dt_go/he"
	"dt_go/pack"
	"dt_go/tree"
	"dt_go/treeio"
	"fmt"
	"io"
	"math"
	"os"
	"time"
	"path/filepath"

)

type QueryStat struct {
	InCt  int
	OutCt int
}

type PredStat struct {
	Ok      bool
	Checked int
	Tol     float64

	MeanAbs   float64
	MaxAbs    float64
	MaxIdx    int
	MaxGot    float64
	MaxRef    float64
	MatchRate float64

	Time time.Duration
}

type RunResult struct {
	K int

	PackTime time.Duration
	TreeTime time.Duration
	TravTime time.Duration
	BtsTime  time.Duration
	CmpTime  time.Duration

	Query QueryStat
	Pred  PredStat

	D int
}

func ctBytes(ct int) int { return ct * he.CtSize }

func PrintSummary(results []RunResult, slots int, name string) {
	dir := filepath.Dir(name)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Println("failed to create result directory:", err)
			return
		}
	}

	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("failed to open summary file:", err)
		return
	}
	defer f.Close()

	w := io.MultiWriter(os.Stdout, f)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "==================================== SUMMARY ====================================")
	fmt.Fprintf(w, "%-4s %-3s | %10s %10s %10s %10s | %12s %12s %12s %12s | %10s %10s | %12s %12s | %10s\n",
		"K", "D",
		"CMP(s)", "TREE(s)", "BTS(s)", "TRAV(s)",
		"CMP(ms/slot)", "TREE(ms/slot)", "BTS(ms/slot)", "TRAV(ms/slot)",
		"In(MB)", "Out(MB)",
		"In(KB/slot)", "Out(KB/slot)",
		"Match(%)",
	)
	fmt.Fprintln(w, "--------------------------------------------------------------------------------")

	var (
		sumMean  float64
		maxMax   float64
		sumMatch float64
		cnt      int

		totalCmp  time.Duration
		totalTree time.Duration
		totalBts  time.Duration
	)

	for _, r := range results {
		inBytes := ctBytes(r.Query.InCt)
		outBytes := ctBytes(r.Query.OutCt)

		cmp := r.CmpTime
		tree := r.TreeTime
		bts := r.BtsTime
		trav := r.TravTime

		// 일반 시간: s
		cmpS := cmp.Seconds()
		treeS := tree.Seconds()
		btsS := bts.Seconds()
		travS := trav.Seconds()

		// amort 시간: ms/slot
		den := float64(slots)
		cmpAm := cmpS * 1000.0 / den
		treeAm := treeS * 1000.0 / den
		btsAm := btsS * 1000.0 / den
		travAm := travS * 1000.0 / den

		// query: MB, amort: KB/slot
		inMB := float64(inBytes) / 1024.0 / 1024.0
		outMB := float64(outBytes) / 1024.0 / 1024.0
		inAmKB := float64(inBytes) / 1024.0 / den
		outAmKB := float64(outBytes) / 1024.0 / den

		fmt.Fprintf(w, "%-4d %-3d | %10.2f %10.2f %10.2f %10.2f | %12.3f %12.3f %12.3f %12.3f | %10.2f %10.2f | %12.3f %12.3f | %10.2f\n",
			r.K, r.D,
			cmpS, treeS, btsS, travS,
			cmpAm, treeAm, btsAm, travAm,
			inMB, outMB,
			inAmKB, outAmKB,
			r.Pred.MatchRate,
		)

		totalCmp += cmp
		totalTree += tree
		totalBts += bts

		if math.IsNaN(r.Pred.MeanAbs) || math.IsNaN(r.Pred.MaxAbs) || math.IsNaN(r.Pred.MatchRate) {
			continue
		}
		sumMean += r.Pred.MeanAbs
		if r.Pred.MaxAbs > maxMax {
			maxMax = r.Pred.MaxAbs
		}
		sumMatch += r.Pred.MatchRate
		cnt++
	}

	fmt.Fprintln(w, "--------------------------------------------------------------------------------")

	if cnt > 0 {
		fmt.Fprintf(w, "[SUMMARY] runs=%d slots=%d | avg_meanAbs=%.6g global_maxAbs=%.6g avg_matchRate=%.2f%%\n",
			cnt, slots, sumMean/float64(cnt), maxMax, sumMatch/float64(cnt))
	}

	if len(results) > 0 {
		travTotal := totalTree - totalBts

		cmpS := totalCmp.Seconds()
		treeS := totalTree.Seconds()
		btsS := totalBts.Seconds()
		travS := travTotal.Seconds()

		den := float64(slots * len(results))
		cmpAm := cmpS * 1000.0 / den
		treeAm := treeS * 1000.0 / den
		btsAm := btsS * 1000.0 / den
		travAm := travS * 1000.0 / den

		fmt.Fprintf(w, "[TIME] total: CMP=%.2fs TREE=%.2fs BTS=%.2fs TRAV=%.2fs\n", cmpS, treeS, btsS, travS)
		fmt.Fprintf(w, "[TIME] amort: CMP=%.3fms/slot TREE=%.3fms/slot BTS=%.3fms/slot TRAV=%.3fms/slot\n",
			cmpAm, treeAm, btsAm, travAm)
	}

	fmt.Fprintln(w, "================================================================================")
	fmt.Fprintln(w)
}

func checkPred(dec []complex128, ref []float32, slots int, tol float64) (ok bool, meanAbs, maxAbs float64, maxIdx int, maxGot, maxRef float64, matchRate float64) {
	checkSlots := slots

	ok = true
	maxIdx = -1
	var sumAbs float64
	var matchCnt int

	for s := 0; s < checkSlots; s++ {
		got := real(dec[s])
		r := float64(ref[s])

		d := math.Abs(got - r)
		sumAbs += d
		if d > maxAbs {
			maxAbs = d
			maxIdx = s
			maxGot = got
			maxRef = r
		}

		gotCls := 0
		if got >= 0 {
			gotCls = 1
		}
		refCls := 0
		if r >= 0 {
			refCls = 1
		}
		if gotCls == refCls {
			matchCnt++
		}
	}

	meanAbs = sumAbs / float64(checkSlots)
	if maxAbs > tol {
		ok = false
	}

	matchRate = 100.0 * float64(matchCnt) / float64(checkSlots)
	return
}

func runOnce(bundle treeio.ParsedBundle, slots int, K int, hook tree.ProgressHook, tol float64, MODE string, P int) RunResult {
	res := RunResult{K: K}

	t1 := time.Now()
	pk := pack.PackBundle(bundle)
	res.PackTime = time.Since(t1)
	res.D = pk.Forest[0].D

	fmt.Printf("[PACK] trees=%d D=%d time=%s\n", len(pk.Forest), pk.Forest[0].D, res.PackTime)

	if MODE == "obo" {
		fmt.Println("[OBO-CMP] start")

		stats, cmptime := tree.EvalXGBForestHEOBOCompareOnlyAllNodes(
			pk.Comp.ValsBitsList,
			pk.Comp.BordersBitsList,
			pk.XSlots,
			pk.Borders,
			pk.Forest,
			slots,
			tol,
			P,
		)

		res.CmpTime = cmptime
		res.TreeTime = 0
		res.BtsTime = 0
		res.TravTime = 0

		res.Query = QueryStat{InCt: pk.InQuery * (32 / P), OutCt: pk.OutQuery}

		okAll := true
		sumMean := 0.0
		globalMax := 0.0
		worst := tree.CmpErrStat{MaxIdx: -1}

		for _, st := range stats {
			sumMean += st.MeanAbs
			if st.MaxAbs > globalMax {
				globalMax = st.MaxAbs
				worst = st
			}
			if !st.Ok {
				okAll = false
			}
		}

		avgMean := math.NaN()
		if len(stats) > 0 {
			avgMean = sumMean / float64(len(stats))
		}

		res.Pred = PredStat{
			Ok:      okAll,
			Checked: slots * len(stats),
			Tol:     tol,

			MeanAbs: avgMean,
			MaxAbs:  globalMax,
			MaxIdx:  worst.MaxIdx,
			MaxGot:  worst.MaxGot,
			MaxRef:  worst.MaxRef,

			Time: 0,
		}

		fmt.Printf("[OBO-CMP] comps=%d cmptime=%s | avg_meanAbs=%.6g global_maxAbs=%.6g\n",
			len(stats), cmptime, avgMean, globalMax)
		if worst.MaxIdx >= 0 {
			fmt.Printf("[OBO-CMP] worst: tid=%d node=%d (f=%d,t=%d) slot=%d got=%.6g ref=%.6g err=%.6g\n",
				worst.Tid, worst.Node, worst.Feat, worst.Thr,
				worst.MaxIdx, worst.MaxGot, worst.MaxRef, worst.MaxAbs)
		}

		return res
	}

	fmt.Println("[MAIN_TREE] start")
	t3 := time.Now()
	ablation := 0
	if MODE == "bsgs" {
		ablation = 1
	}
	outs, btstime, cmptime := tree.EvalXGBForestHEProdParityParallel(
		pk.Comp.ValsBitsList,
		pk.Comp.BordersBitsList,
		pk.Forest,
		K,
		hook,
		ablation,
		P,
	)
	sum := outs[0]
	for i := 1; i < len(outs); i++ {
		he.Add(sum, outs[i], &sum)
	}
	he.Relin(sum)
	he.Rescale(sum)
	BaseMargin := treeio.BaseMarginFromBaseScore(pk.BaseScore)
	if pk.BaseScore != 0 {
		he.Add(sum, he.KFloat(BaseMargin), &sum)
	}
	res.TreeTime = time.Since(t3)
	res.BtsTime = btstime
	res.CmpTime = cmptime
	res.TravTime = res.TreeTime - btstime - cmptime
	res.Query = QueryStat{InCt: pk.InQuery, OutCt: pk.OutQuery}
	fmt.Printf("[MAIN_TREE] time=%s\n", res.TreeTime)

	if !sum.IsCt() || sum.Ct == nil {
		res.Pred = PredStat{
			Ok:      false,
			Checked: slots,
			Tol:     tol,
			MeanAbs: math.NaN(),
			MaxAbs:  math.NaN(),
			MaxIdx:  -1,
			MaxGot:  math.NaN(),
			MaxRef:  math.NaN(),
			Time:    0,
		}
		fmt.Printf("[END2END] final_sum is not ct (const/nil). const=%.6f\n", sum.K.AsFloat64())
		return res
	}

	dec := he.Decrypt(sum.Ct)

	t4 := time.Now()
	ok, meanAbs, maxAbs, maxIdx, maxGot, maxRef, matchRate := checkPred(dec, bundle.Pred, slots, tol)
	elapsed := time.Since(t4)

	res.Pred = PredStat{
		Ok:        ok,
		Checked:   slots,
		Tol:       tol,
		MeanAbs:   meanAbs,
		MaxAbs:    maxAbs,
		MaxIdx:    maxIdx,
		MaxGot:    maxGot,
		MaxRef:    maxRef,
		MatchRate: matchRate,
		Time:      elapsed,
	}
	return res
}
