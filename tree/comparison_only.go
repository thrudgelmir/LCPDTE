package tree

import (
	"dt_go/he"
	"math"
	"time"
)

type CmpErrStat struct {
	Ok   bool
	Tol  float64
	Tid  int
	Node int
	Feat int
	Thr  int

	MeanAbs float64
	MaxAbs  float64
	MaxIdx  int
	MaxGot  float64
	MaxRef  float64
}

func checkCmpAgainstThr(dec []complex128, xSlots []float32, thr float32, slots int, tol float64) (ok bool, meanAbs, maxAbs float64, maxIdx int, maxGot, maxRef float64) {
	ok = true
	maxIdx = -1

	var sumAbs float64
	for s := 0; s < slots; s++ {
		got := real(dec[s])

		ref := 0.0
		if xSlots[s] >= thr {
			ref = 1.0
		}

		d := math.Abs(got - ref)
		sumAbs += d
		if d > maxAbs {
			maxAbs = d
			maxIdx = s
			maxGot = got
			maxRef = ref
		}
	}

	meanAbs = sumAbs / float64(slots)
	if maxAbs > tol {
		ok = false
	}
	return
}

func checkCmpConstAgainstThr(gotConst float64, xSlots []float32, thr float32, slots int, tol float64) (ok bool, meanAbs, maxAbs float64, maxIdx int, maxGot, maxRef float64) {
	ok = true
	maxIdx = -1

	var sumAbs float64
	for s := 0; s < slots; s++ {
		ref := 0.0
		if xSlots[s] >= thr {
			ref = 1.0
		}

		d := math.Abs(gotConst - ref)
		sumAbs += d
		if d > maxAbs {
			maxAbs = d
			maxIdx = s
			maxGot = gotConst
			maxRef = ref
		}
	}

	meanAbs = sumAbs / float64(slots)
	if maxAbs > tol {
		ok = false
	}
	return
}

func EvalXGBForestHEOBOCompareOnlyAllNodes(
	valsBitsList [][]he.CT,
	bordersBitsList [][][]he.CT,
	xSlots [][]float32,
	borders [][]float32,
	forest []*Tree,
	slots int,
	tol float64,
	P int,
) ([]CmpErrStat, time.Duration) {

	total := 0
	for _, tr := range forest {
		total += (1 << tr.D) - 1
	}
	stats := make([]CmpErrStat, 0, total)

	var cmptime time.Duration

	for tid := range forest {
		tr := forest[tid]
		nodeN := (1 << tr.D) - 1

		for node := 0; node < nodeN; node++ {
			f := tr.NodeFeat[node]
			t := tr.NodeThr[node]


			xb := valsBitsList[f][:P/2] // complex pack half
			tb := bordersBitsList[f][t][:P]
			thr := borders[f][t]
			xF := xSlots[f]

			t0 := time.Now()
			sel := CmpGeBits(uncplx(xb), tb)
			cmptime += time.Since(t0)

			var ok bool
			var meanAbs, maxAbs float64
			var maxIdx int
			var maxGot, maxRef float64

			if sel.IsConst() {
				ok, meanAbs, maxAbs, maxIdx, maxGot, maxRef =
					checkCmpConstAgainstThr(sel.K.AsFloat64(), xF, thr, slots, tol)
			} else {
				dec := he.Decrypt(sel.Ct)
				ok, meanAbs, maxAbs, maxIdx, maxGot, maxRef =
					checkCmpAgainstThr(dec, xF, thr, slots, tol)
			}

			stats = append(stats, CmpErrStat{
				Ok:   ok,
				Tol:  tol,
				Tid:  tid,
				Node: node,
				Feat: f,
				Thr:  t,

				MeanAbs: meanAbs,
				MaxAbs:  maxAbs,
				MaxIdx:  maxIdx,
				MaxGot:  maxGot,
				MaxRef:  maxRef,
			})
		}
	}

	return stats, cmptime
}
