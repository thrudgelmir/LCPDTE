package treeio

import (
	"math"
	"math/rand"
)

type GenConfig struct {
	Seed int64
	N    int
	F    int

	NumTrees int
	Depth    int

	BaseScore float64

	XMean   float64
	XStd    float64
	NanProb float64

	ThrMean float64
	ThrStd  float64
	LeafStd float64

	// SparseRatio controls how "sparse" (pruned) the generated trees are.
	// - 1.0: perfect/full tree (same as before).
	// - r in (0,1): each internal node is kept with prob r; otherwise it becomes a leaf (pruning).
	// If this option is omitted (zero value), it defaults to 1.0.
	SparseRatio float64
}

func BaseMarginFromBaseScore(p float64) float64 {
	const eps = 1e-15
	if p < eps {
		p = eps
	} else if p > 1.0-eps {
		p = 1.0 - eps
	}
	return math.Log(p / (1.0 - p))
}

func GenBundle(cfg GenConfig) ParsedBundle {
	rng := rand.New(rand.NewSource(cfg.Seed))

	in := genInput(cfg, rng)
	model := genModel(cfg, rng)

	pred := make([]float32, in.N)
	for i := 0; i < in.N; i++ {
		pred[i] = evalRawModelOneRow(model, in, i)
	}

	return ParsedBundle{Model: model, Input: in, Pred: pred}
}

func genInput(cfg GenConfig, rng *rand.Rand) RawInput {
	X := make([]float32, cfg.N*cfg.F)
	for i := 0; i < cfg.N; i++ {
		for f := 0; f < cfg.F; f++ {
			v := float32(cfg.XMean + rng.NormFloat64()*cfg.XStd + 0.05*float64(f))
			if cfg.NanProb > 0 && rng.Float64() < cfg.NanProb {
				v = float32(math.NaN())
			}
			X[i*cfg.F+f] = v
		}
	}
	return RawInput{N: cfg.N, F: cfg.F, X: X}
}

func genModel(cfg GenConfig, rng *rand.Rand) RawModel {
	sr := normSparseRatio(cfg.SparseRatio)

	trees := make([]RawTree, cfg.NumTrees)
	for i := 0; i < cfg.NumTrees; i++ {
		trees[i] = genTreeWithSparsity(cfg.F, cfg.Depth, cfg.ThrMean, cfg.ThrStd, cfg.LeafStd, sr, rng)
	}
	return RawModel{
		BaseScore:      cfg.BaseScore,
		NumFeature:     cfg.F,
		NumOutputGroup: 0,
		TreeInfo:       make([]int, cfg.NumTrees),
		Trees:          trees,
	}
}

func normSparseRatio(x float64) float64 {
	if x <= 0 {
		return 1.0
	}
	if x > 1.0 {
		return 1.0
	}
	return x
}

func genPerfectTree(F, depth int, thrMean, thrStd, leafStd float64, rng *rand.Rand) RawTree {
	return genTreeWithSparsity(F, depth, thrMean, thrStd, leafStd, 1.0, rng)
}

func genTreeWithSparsity(F, depth int, thrMean, thrStd, leafStd, sparseRatio float64, rng *rand.Rand) RawTree {
	nodeN := (1 << (depth + 1)) - 1
	leafBase := (1 << depth) - 1

	left := make([]int32, nodeN)
	right := make([]int32, nodeN)
	sidx := make([]int32, nodeN)
	scond := make([]float32, nodeN)
	defLeft := make([]bool, nodeN)
	sType := make([]uint8, nodeN)

	keep := make([]bool, nodeN)
	if sparseRatio < 1.0 {
		leaf := leafBase + rng.Intn(1<<depth) // [leafBase, leafBase + 2^depth)
		nid := leaf
		for {
			keep[nid] = true
			if nid == 0 {
				break
			}
			nid = (nid - 1) / 2
		}
	}

	for nid := 0; nid < nodeN; nid++ {
		// 마지막 레벨은 항상 leaf
		if nid >= leafBase {
			left[nid], right[nid] = -1, -1
			sidx[nid] = -1
			scond[nid] = float32(rng.NormFloat64() * leafStd) // leaf value
			defLeft[nid] = false
			sType[nid] = 0
			continue
		}

		if sparseRatio < 1.0 && !keep[nid] && rng.Float64() >= sparseRatio {
			left[nid], right[nid] = -1, -1
			sidx[nid] = -1
			scond[nid] = float32(rng.NormFloat64() * leafStd) // leaf value
			defLeft[nid] = false
			sType[nid] = 0
			continue
		}

		// internal node
		left[nid] = int32(2*nid + 1)
		right[nid] = int32(2*nid + 2)
		sidx[nid] = int32(rng.Intn(F))
		scond[nid] = float32(thrMean + rng.NormFloat64()*thrStd)
		defLeft[nid] = false
		sType[nid] = 0
	}

	return RawTree{
		SplitIndex:  sidx,
		SplitCond:   scond,
		DefaultLeft: defLeft,
		Left:        left,
		Right:       right,
		SplitType:   sType,
	}
}

func evalRawModelOneRow(m RawModel, in RawInput, row int) float32 {
	xrow := in.X[row*in.F : (row+1)*in.F]
	y := float32(BaseMarginFromBaseScore(m.BaseScore))
	for i := 0; i < len(m.Trees); i++ {
		y += evalRawTreeOneRow(m.Trees[i], xrow, in.F)
	}
	return y
}

func evalRawTreeOneRow(t RawTree, x []float32, F int) float32 {
	nid := 0
	for {
		if t.Left[nid] == -1 {
			return t.SplitCond[nid]
		}
		fid := int(t.SplitIndex[nid])
		thr := t.SplitCond[nid]

		v := float32(math.NaN())
		if 0 <= fid && fid < F {
			v = x[fid]
		}
		if v != v {
			if t.DefaultLeft[nid] {
				nid = int(t.Left[nid])
			} else {
				nid = int(t.Right[nid])
			}
			continue
		}
		if v < thr {
			nid = int(t.Left[nid])
		} else {
			nid = int(t.Right[nid])
		}
	}
}
