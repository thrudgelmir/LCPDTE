package tree

import (
	"dt_go/he"
	"math"
	"math/bits"
)

func BalancedProducts(prev []he.CT, bit he.CT) []he.CT {
	N0 := len(prev) // = 2^(k-1)
	N := N0 * 2
	prods := make([]he.CT, N)

	for i, v := range prev {
		prods[i<<1] = v
	}
	prods[1] = bit

	for mask := 0; mask < N; mask++ {
		if !isNilCT(prods[mask]) {
			continue
		}
		pc := bits.OnesCount(uint(mask))
		half := pc / 2

		left := 0
		cnt := half
		tmp := mask
		idx := 0
		for cnt > 0 && tmp > 0 {
			if (tmp & 1) == 1 {
				left |= (1 << idx)
				cnt--
			}
			tmp >>= 1
			idx++
		}
		right := mask ^ left

		prods[mask] = he.MulRelinNew(prods[left], prods[right])
		he.Rescale(prods[mask])
	}
	return prods
}

func OneHotFromProdsIterative(prods []he.CT) []he.CT {
	arr := make([]he.CT, len(prods))
	copy(arr, prods)

	N := len(arr)
	d := int(math.Log2(float64(N)))

	for lvl := d - 1; lvl >= 0; lvl-- {
		size := 1 << lvl
		for i := 0; i < N; i += 2 * size {
			for j := 0; j < size; j++ {
				a := arr[i+j]
				b := arr[i+size+j]
				arr[i+j] = he.SubNew(a, b)
				arr[i+size+j] = b
			}
		}
	}
	return arr
}

func OneHotUpdateFromPrev(prevOh []he.CT, prodsNew []he.CT) []he.CT {
	N := len(prevOh)

	rightProds := make([]he.CT, N)
	for i := 0; i < N; i++ {
		rightProds[i] = prodsNew[(i<<1)|1]
	}
	ohRight := OneHotFromProdsIterative(rightProds)
	ohNew := make([]he.CT, 2*N)
	for i := 0; i < N; i++ {
		r := ohRight[i]
		ohNew[2*i+1] = r
		ohNew[2*i] = he.SubNew(prevOh[i], r)
	}
	return ohNew
}
