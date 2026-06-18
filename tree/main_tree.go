package tree

import (
	"container/heap"
	"dt_go/he"
	"time"
)

type pqItem struct {
	pri int
	uid int
	tid int
}
type featPQ []*pqItem

func (h featPQ) Len() int { return len(h) }
func (h featPQ) Less(i, j int) bool {
	if h[i].pri == h[j].pri {
		return h[i].uid < h[j].uid
	}
	return h[i].pri > h[j].pri
}
func (h featPQ) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *featPQ) Push(x any)   { *h = append(*h, x.(*pqItem)) }
func (h *featPQ) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func isNilCT(x he.CT) bool { return x.Ct == nil && x.K == nil }

type Tree struct {
	D          int
	NodeFeat   []int
	NodeThr    []int
	NodeIsLeaf []bool
	LeafValues []float64
}

type treeState struct {
	id    int
	D     int
	depth int

	prodsEven []he.CT
	prodsOdd  []he.CT
	ohEven    []he.CT
	ohOdd     []he.CT

	pendingSel he.CT
	done       bool
	out        he.CT

	nodeFeat   []int
	nodeThr    []int
	nodeIsLeaf []bool
	leafValues []float64
}

func remaining(s *treeState) int { return s.D - s.depth }

type treePQ []*pqItem

func (h treePQ) Len() int { return len(h) }
func (h treePQ) Less(i, j int) bool {
	if h[i].pri == h[j].pri {
		return h[i].uid < h[j].uid
	}
	return h[i].pri > h[j].pri
}
func (h treePQ) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *treePQ) Push(x any)   { *h = append(*h, x.(*pqItem)) }
func (h *treePQ) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func select1D(vals, oh []he.CT, rs bool) he.CT {
	m := len(vals)
	if m == 1 {
		return vals[0]
	}

	needRelin := false
	needRescale := false

	for i := 0; i < m; i++ {
		if oh[i].IsCt() && vals[i].IsCt() {
			needRelin = true
			needRescale = true
			break
		}
		if oh[i].IsCt() && vals[i].K.IsFloat() {
			needRescale = true
		}
		if oh[i].K.IsFloat() && vals[i].IsCt() {
			needRescale = true
		}
	}
	res := he.KInt(0)
	for i := 0; i < m; i++ {
		res = he.AddNew(res, he.MulNew(oh[i], vals[i]))
	}
	if rs && needRelin {
		he.Relin(res)
	}
	if rs && needRescale {
		he.Rescale(res)
	}
	return res
}

func bitsMSB(x, n int) []int {
	out := make([]int, n)
	for i := 0; i < n; i++ {
		out[i] = (x >> (n - 1 - i)) & 1
	}
	return out
}

func interleaveIndex(ie, io, depth, ne, no int) int {
	be := bitsMSB(ie, ne)
	bo := bitsMSB(io, no)
	e, o := 0, 0
	idx := 0
	for p := 0; p < depth; p++ {
		b := 0
		if (p & 1) == 0 {
			b = be[e]
			e++
		} else {
			b = bo[o]
			o++
		}
		idx = (idx << 1) | b
	}
	return idx
}

var idxMapCache = map[int][][]int{}

func buildIdxMap(depth int) [][]int {
	if mp, ok := idxMapCache[depth]; ok {
		return mp
	}
	ne := (depth + 1) / 2
	no := depth / 2
	M := 1 << ne
	S := 1 << no
	mp := make([][]int, M)
	for ie := 0; ie < M; ie++ {
		mp[ie] = make([]int, S)
		for io := 0; io < S; io++ {
			mp[ie][io] = interleaveIndex(ie, io, depth, ne, no)
		}
	}
	idxMapCache[depth] = mp
	return mp
}

func selectByParity(values, ohEven, ohOdd []he.CT, depth int, ablation int, rlOpt ...bool) he.CT {
	if depth == 0 {
		return values[0]
	}
	ne := (depth + 1) / 2
	no := depth / 2
	rs := rlOpt[0]
	if ablation == 1 {
		return select1D(values, ohEven, rs)
	}

	if no == 0 {
		return select1D(values, ohEven, rs)
	}
	if ne == 0 {
		return select1D(values, ohOdd, rs)
	}

	mp := buildIdxMap(depth)
	M := 1 << ne
	S := 1 << no

	if len(ohEven) >= len(ohOdd) {
		cols := make([]he.CT, S)
		for io := 0; io < S; io++ {
			col := make([]he.CT, M)
			for ie := 0; ie < M; ie++ {
				col[ie] = values[mp[ie][io]]
			}
			cols[io] = select1D(col, ohEven, true)
		}
		return select1D(cols, ohOdd, rs)
	}

	rows := make([]he.CT, M)
	for ie := 0; ie < M; ie++ {
		row := make([]he.CT, S)
		for io := 0; io < S; io++ {
			row[io] = values[mp[ie][io]]
		}
		rows[ie] = select1D(row, ohOdd, true)
	}
	return select1D(rows, ohEven, rs)
}

func uncplx(x []he.CT) []he.CT {
	ret := make([]he.CT, len(x)*2)
	if x[0].IsConst() {
		return ret
	}
	for i := 0; i < len(x); i++ {
		cplx := he.ConjugateNew(x[i])
		real := he.AddNew(x[i], cplx)
		imag := he.SubNew(cplx, x[i])
		tmp, _ := he.Eval.MulNew(imag.Ct, complex(0.0, 1.0))
		imag = he.FromCt(tmp)
		ret[i*2] = real
		ret[i*2+1] = imag
	}
	return ret
}

func EvalXGBForestHEProdParityParallel(valsBitsList [][]he.CT, bordersBitsList [][][]he.CT, forest []*Tree, k int, hook ProgressHook, ablation int, P int) ([]he.CT, time.Duration, time.Duration) {
	states := make([]*treeState, len(forest))
	var btstime time.Duration
	var cmptime time.Duration
	for tid := range forest {
		tr := forest[tid]
		states[tid] = &treeState{
			id:        tid,
			D:         tr.D,
			depth:     0,
			prodsEven: []he.CT{he.KInt(1)},
			prodsOdd:  []he.CT{he.KInt(1)},
			ohEven:    []he.CT{he.KInt(1)},
			ohOdd:     []he.CT{he.KInt(1)},
			nodeFeat:  tr.NodeFeat, nodeThr: tr.NodeThr, nodeIsLeaf: tr.NodeIsLeaf, leafValues: tr.LeafValues,
		}
	}

	done := 0
	doneFlag := make([]bool, len(forest))
	markDone := func(tid int, note string) {
		if doneFlag[tid] {
			return
		}
		doneFlag[tid] = true
		done++
		emit(hook, ProgressEvent{
			Where: "main_tree.done",
			Done:  done,
			Total: len(forest),
			Note:  note,
		})
	}

	h := &treePQ{}
	heap.Init(h)
	uid := 0
	for tid := range states {
		uid++
		heap.Push(h, &pqItem{pri: remaining(states[tid]), uid: uid, tid: tid})
	}

	for h.Len() > 0 {
		picked := make([]int, 0, k)
		for h.Len() > 0 && len(picked) < k {
			it := heap.Pop(h).(*pqItem)
			if states[it.tid].done {
				continue
			}
			picked = append(picked, it.tid)
		}
		if len(picked) == 0 {
			break
		}

		pendingTids := make([]int, 0, len(picked))
		pendingSel := make([]he.CT, 0, len(picked))
		for _, tid := range picked {
			s := states[tid]
			d := s.depth
			if d >= s.D {
				continue
			}

			cnt := 1 << d
			base := (1 << d) - 1

			// xs, ts: (cnt, 32)
			xs := make([][]he.CT, cnt)
			ts := make([][]he.CT, cnt)
			for i := 0; i < cnt; i++ {
				xs[i] = make([]he.CT, P/2)
				ts[i] = make([]he.CT, P)

				nodeIdx := base + i
				if s.nodeIsLeaf[nodeIdx] {
					for bi := 0; bi < P/2; bi++ {
						xs[i][bi] = he.KInt(0)
					}
					for bi := 0; bi < P; bi++ {
						ts[i][bi] = he.KInt(1)
					}
					continue
				}

				f := s.nodeFeat[nodeIdx]
				t := s.nodeThr[nodeIdx]
				xb := valsBitsList[f]
				tb := bordersBitsList[f][t]
				for bi := 0; bi < P/2; bi++ {
					xs[i][bi] = xb[bi]
				}
				for bi := 0; bi < P; bi++ {
					ts[i][bi] = tb[bi]
				}
			}
			if d == 0 {
				if s.nodeIsLeaf[0] {
					s.pendingSel = he.KInt(0)
				} else {
					timeCmpStart := time.Now()
					s.pendingSel = CmpGeBits(uncplx(xs[0]), ts[0])
					timeCmpend := time.Since(timeCmpStart)
					cmptime += timeCmpend
				}
			} else {
				selectedX := make([]he.CT, P/2)
				selectedT := make([]he.CT, P)

				for bi := 0; bi < P/2; bi++ {
					colX := make([]he.CT, cnt)
					for i := 0; i < cnt; i++ {
						colX[i] = xs[i][bi]
					}
					selectedX[bi] = selectByParity(colX, s.ohEven, s.ohOdd, d, ablation, true)
				}
				for bi := 0; bi < P; bi++ {
					colT := make([]he.CT, cnt)
					for i := 0; i < cnt; i++ {
						colT[i] = ts[i][bi]
					}
					selectedT[bi] = selectByParity(colT, s.ohEven, s.ohOdd, d, ablation, true)
				}
				timeCmpStart := time.Now()
				s.pendingSel = CmpGeBits(uncplx(selectedX), selectedT)
				timeCmpend := time.Since(timeCmpStart)
				cmptime += timeCmpend

			}
			pendingTids = append(pendingTids, tid)
			pendingSel = append(pendingSel, s.pendingSel)
		}

		if len(pendingSel) > 0 {
			timeBtsStart := time.Now()
			bts := he.MultiBootstrapCT(pendingSel)
			timeBtsend := time.Since(timeBtsStart)
			btstime += timeBtsend
			for i := range bts {
				tid := pendingTids[i]
				s := states[tid]
				d := s.depth

				selBts := bts[i]
				if ablation == 1 {
					s.prodsEven = BalancedProducts(s.prodsEven, selBts)
					s.ohEven = OneHotFromProdsIterative(s.prodsEven)
				} else {
					if (d & 1) == 0 {
						s.prodsEven = BalancedProducts(s.prodsEven, selBts)
						s.ohEven = OneHotFromProdsIterative(s.prodsEven)
					} else {
						s.prodsOdd = BalancedProducts(s.prodsOdd, selBts)
						s.ohOdd = OneHotFromProdsIterative(s.prodsOdd)
					}
				}

				s.depth++
				s.pendingSel = he.CT{}

				if s.depth == s.D {
					leafCT := make([]he.CT, len(s.leafValues))
					for j := range s.leafValues {
						leafCT[j] = he.KFloat(s.leafValues[j]) // 상수 encrypt 안 함
					}
					s.out = selectByParity(leafCT, s.ohEven, s.ohOdd, s.D, ablation, false)
					s.done = true
					markDone(tid, "finalize")
				}
			}
		}

		for _, tid := range picked {
			if states[tid].done {
				continue
			}
			uid++
			heap.Push(h, &pqItem{pri: remaining(states[tid]), uid: uid, tid: tid})
		}
	}

	out := make([]he.CT, len(states))
	for i := range states {
		out[i] = states[i].out
	}
	return out, btstime, cmptime
}
