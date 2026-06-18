// tree/precompute_border_float32.go
package tree

import (
	"dt_go/he"
	"math"
	"sort"
)

// float32 -> order-preserving uint32 (total order key)
func f32ToOrderedU32(x float32) uint32 {
	u := math.Float32bits(x)
	if (u & 0x80000000) != 0 {
		return ^u
	}
	return u ^ 0x80000000
}

// uint32 key -> [32] const bits (MSB first) as he.CT
func u32ToBits32CT(u uint32) []he.CT {
	out := make([]he.CT, 32)
	if u == 0 {
		for i := 0; i < 32; i++ {
			out[i] = he.KInt(0) // CT.K.I = 0 (정수 const)
		}
		return out
	}
	for i := 0; i < 32; i++ {
		b := int((u >> (31 - uint(i))) & 1)
		out[i] = he.ConstCT(b)
	}
	return out
}

func padBordersU32(borders []uint32) ([]uint32, int) {
	L := len(borders)
	n := 1
	for (1<<n)-1 < L {
		n++
	}
	target := (1 << n) - 1
	pad := target - L

	out := make([]uint32, 0, target)
	for i := 0; i < pad; i++ {
		//out = append(out, borders[0])
		out = append(out, 0)
	}
	out = append(out, borders...)
	return out, pad
}

func makeBorderLevelsU32(borders []uint32) [][]uint32 {
	res := [][]uint32{}
	var helper func([]uint32, int)
	helper = func(arr []uint32, level int) {
		if len(arr) == 0 {
			return
		}
		mid := len(arr) / 2
		if level == len(res) {
			res = append(res, []uint32{})
		}
		res[level] = append(res[level], arr[mid])
		helper(arr[:mid], level+1)
		helper(arr[mid+1:], level+1)
	}
	helper(borders, 0)
	return res
}

// PrecomputeBorderFloat32CT:
// float32 borders -> level별 threshold bits(he.CT const) + padding b
// return levelsBits: [H][..][32]he.CT , bpad
func PrecomputeBorderFloat32CT(borders []float32) ([][][]he.CT, int) {
	cp := make([]float32, len(borders))
	copy(cp, borders)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })

	for i := 1; i < len(cp); i++ {
		if !(cp[i] > cp[i-1]) {
			cp[i] = math.Nextafter32(cp[i-1], float32(math.Inf(1)))
		}
	}

	keys := make([]uint32, len(cp))
	for i := range cp {
		keys[i] = f32ToOrderedU32(cp[i])
	}

	padded, bpad := padBordersU32(keys)
	levelKeys := makeBorderLevelsU32(padded)

	levelsBits := make([][][]he.CT, len(levelKeys))
	for lv := range levelKeys {
		levelsBits[lv] = make([][]he.CT, len(levelKeys[lv]))
		for j := range levelKeys[lv] {
			levelsBits[lv][j] = u32ToBits32CT(levelKeys[lv][j])
		}
	}
	return levelsBits, bpad
}
