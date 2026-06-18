package pack

import (
	"math"
	"sort"

	"dt_go/he"
	"dt_go/tree"
	"dt_go/treeio"
)

type ComparisonPrep struct {
	ValsBitsList    [][]he.CT   // [F][32]
	BordersBitsList [][][]he.CT // [F][Lf][32]
}

type PackedData struct {
	Borders   [][]float32
	XSlots    [][]float32
	Forest    []*tree.Tree
	BaseScore float64
	Comp      ComparisonPrep
	InQuery   int
	OutQuery  int
}

func PackBundle(bundle treeio.ParsedBundle) PackedData {
	slots := he.Params.MaxSlots()

	xSlots := reshapeToFeatureSlots(bundle.Input, slots)

	borders, thrIndexMaps := buildBorders(bundle.Model, bundle.Input.F)

	D := modelMaxDepth(bundle.Model)

	forest := make([]*tree.Tree, len(bundle.Model.Trees))
	for i := range bundle.Model.Trees {
		forest[i] = convertRawTreeToMainTree(bundle.Model.Trees[i], D, thrIndexMaps, 0.0)
	}

	inquery := 0
	outquery := 1
	valsBitsList := make([][]he.CT, bundle.Input.F)
	for f := 0; f < bundle.Input.F; f++ {
		valsBitsList[f] = encryptBitsFromFloat32Slots(xSlots[f])
		inquery += len(valsBitsList[f])
	}

	bordersBitsList := make([][][]he.CT, bundle.Input.F)
	for f := 0; f < bundle.Input.F; f++ {
		bordersBitsList[f] = make([][]he.CT, len(borders[f]))
		for t := 0; t < len(borders[f]); t++ {
			u := f32ToOrderedU32(borders[f][t])
			bordersBitsList[f][t] = u32ToBits32ConstCT(u)
		}
	}

	return PackedData{
		Borders:   borders,
		XSlots:    xSlots,
		Forest:    forest,
		BaseScore: bundle.Model.BaseScore,
		OutQuery:  outquery,
		InQuery:   inquery,
		Comp: ComparisonPrep{
			ValsBitsList:    valsBitsList,
			BordersBitsList: bordersBitsList,
		},
	}
}

func reshapeToFeatureSlots(in treeio.RawInput, slots int) [][]float32 {
	if in.N < slots {
		panic("RawInput N < slots")
	}
	xSlots := make([][]float32, in.F)
	for f := 0; f < in.F; f++ {
		xSlots[f] = make([]float32, slots)
	}
	for i := 0; i < slots; i++ {
		base := i * in.F
		for f := 0; f < in.F; f++ {
			xSlots[f][i] = in.X[base+f]
		}
	}
	return xSlots
}

func buildBorders(m treeio.RawModel, F int) ([][]float32, []map[uint32]int) {
	sets := make([]map[uint32]struct{}, F)
	for f := 0; f < F; f++ {
		sets[f] = map[uint32]struct{}{}
	}

	for _, tr := range m.Trees {
		for nid := 0; nid < len(tr.SplitIndex); nid++ {
			if tr.Left[nid] == -1 {
				continue
			}
			fid := int(tr.SplitIndex[nid])
			thrBits := math.Float32bits(tr.SplitCond[nid])
			sets[fid][thrBits] = struct{}{}
		}
	}

	borders := make([][]float32, F)
	thrIndexMaps := make([]map[uint32]int, F)

	for f := 0; f < F; f++ {
		tmp := make([]float32, 0, len(sets[f]))
		for bits := range sets[f] {
			tmp = append(tmp, math.Float32frombits(bits))
		}
		sort.Slice(tmp, func(i, j int) bool { return tmp[i] < tmp[j] })
		if len(tmp) == 0 {
			tmp = []float32{0}
		}
		borders[f] = tmp

		mp := map[uint32]int{}
		for i := 0; i < len(tmp); i++ {
			mp[math.Float32bits(tmp[i])] = i
		}
		thrIndexMaps[f] = mp
	}

	return borders, thrIndexMaps
}

func modelMaxDepth(m treeio.RawModel) int {
	D := 0
	for _, tr := range m.Trees {
		d := rawTreeMaxDepth(tr, 0, 0)
		if d > D {
			D = d
		}
	}
	return D
}

func rawTreeMaxDepth(tr treeio.RawTree, nid int32, depth int) int {
	if nid < 0 || tr.Left[int(nid)] == -1 {
		return depth
	}
	ld := rawTreeMaxDepth(tr, tr.Left[int(nid)], depth+1)
	rd := rawTreeMaxDepth(tr, tr.Right[int(nid)], depth+1)
	if ld > rd {
		return ld
	}
	return rd
}

func convertRawTreeToMainTree(raw treeio.RawTree, D int, thrIndexMaps []map[uint32]int, padValue float64) *tree.Tree {
	nodeN := (1 << D) - 1
	leafN := 1 << D
	leafBase := (1 << D) - 1

	out := &tree.Tree{
		D:          D,
		NodeFeat:   make([]int, nodeN),
		NodeThr:    make([]int, nodeN),
		NodeIsLeaf: make([]bool, nodeN),
		LeafValues: make([]float64, leafN),
	}

	dummyF, dummyT := 0, 0

	var fillConstSubtree func(bfsIdx int, depth int, v float64)
	fillConstSubtree = func(bfsIdx int, depth int, v float64) {
		if depth == D {
			leafPos := bfsIdx - leafBase
			out.LeafValues[leafPos] = v
			return
		}
		out.NodeFeat[bfsIdx] = dummyF
		out.NodeThr[bfsIdx] = dummyT
		out.NodeIsLeaf[bfsIdx] = true

		fillConstSubtree(2*bfsIdx+1, depth+1, v)
		fillConstSubtree(2*bfsIdx+2, depth+1, v)
	}

	var rec func(rawNid int32, depth int, bfsIdx int)
	rec = func(rawNid int32, depth int, bfsIdx int) {
		if depth == D {
			leafPos := bfsIdx - leafBase
			if rawNid >= 0 && raw.Left[int(rawNid)] == -1 {
				out.LeafValues[leafPos] = float64(raw.SplitCond[int(rawNid)])
			} else {
				out.LeafValues[leafPos] = padValue
			}
			return
		}

		if rawNid < 0 {
			fillConstSubtree(bfsIdx, depth, padValue)
			return
		}

		if raw.Left[int(rawNid)] == -1 {
			fillConstSubtree(bfsIdx, depth, float64(raw.SplitCond[int(rawNid)]))
			return
		}

		// raw internal
		fid := int(raw.SplitIndex[int(rawNid)])
		thr := raw.SplitCond[int(rawNid)]
		tidx, ok := thrIndexMaps[fid][math.Float32bits(thr)]
		if !ok {
			panic("threshold not found in borders")
		}

		out.NodeFeat[bfsIdx] = fid
		out.NodeThr[bfsIdx] = tidx

		rec(raw.Left[int(rawNid)], depth+1, 2*bfsIdx+1)
		rec(raw.Right[int(rawNid)], depth+1, 2*bfsIdx+2)
	}

	rec(0, 0, 0)
	return out
}

func f32ToOrderedU32(x float32) uint32 {
	u := math.Float32bits(x)
	if (u & 0x80000000) != 0 {
		return ^u
	}
	return u ^ 0x80000000
}

func u32ToBits32ConstCT(u uint32) []he.CT {
	out := make([]he.CT, 32)
	for bi := 0; bi < 32; bi++ {
		shift := 31 - uint(bi)
		b := int64((u >> shift) & 1)
		out[bi] = he.KInt(b)
	}
	return out
}

func encryptBitsFromFloat32Slots(xs []float32) []he.CT {
	slots := he.Params.MaxSlots()
	out := make([]he.CT, 16)

	for bi := 0; bi < 16; bi++ {
		shift := 31 - uint(bi*2)
		shift2 := 31 - uint(bi*2+1)
		vals := make([]complex128, slots)
		for i := 0; i < slots; i++ {
			u := f32ToOrderedU32(xs[i])
			b := (u >> shift) & 1
			b2 := (u >> shift2) & 1
			vals[i] = complex(float64(b)/2.0, float64(b2)/2.0) //real = 0,2,4,6.., imag = 1,3,5..
		}
		out[bi] = he.FromCt(he.Encrypt(vals))
	}
	return out
}
