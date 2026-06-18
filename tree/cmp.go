package tree

import (
	"dt_go/he"
)

func largestPowerOfTwoLessThan(L int) int {
	p := 1
	for (p << 1) < L {
		p <<= 1
	}
	return p
}

func CmpGeBits(xBits []he.CT, tBits []he.CT) he.CT {
	//if tBits[0].Ct != nil {
	//	fmt.Println(tBits[0].Ct.Level())
	//}
	n := len(xBits)

	// a[i] = xBits[i] - tBits[i] ({-1,0,1})
	a := make([]he.CT, n)
	for i := 0; i < n; i++ {
		a[i] = he.SubNew(xBits[i], tBits[i])
		//if a[i].IsConst() {
		//	a[i] = he.AddNew(he.FromCt(he.Zeros()), a[i])
		//}
	}

	// if precision 2^k , don't called
	buildSingle := func(i int) (he.CT, he.CT) {
		half := he.MulNew(a[i], he.KFloat(0.5))
		he.Rescale(half)
		return half, he.CT{}
	}

	buildPair := func(i int, isTail bool) (he.CT, he.CT) {
		u := a[i]
		v := a[i+1]

		uvm := he.MulRelinNew(u, v) // compute a[i]*a[i+1]
		he.Rescale(uvm)

		uv := he.AddNew(u, v)              // u + v // compute a[i]+a[i+1]
		uv = he.MulNew(uv, he.KFloat(0.5)) // scale half

		u05 := he.MulNew(u, he.KFloat(0.5)) // compute a[i]/2
		he.Rescale(u05)

		uvmu := he.MulNew(uvm, u05) // (uvm)*u*0.5 // a[i]*a[i]*a[i+1]/2
		// S = (u+v) - uvm*u
		S := he.SubNew(uv, uvmu) // a[i]+a[i+1]-a[i]*a[i]*a[i+1] = a[i] + a[i+1]*(1-a[i]^2) = a[i] + m[i]*a[i+1]

		if isTail {
			he.Relin(S)
			he.Rescale(S)
			return S, he.CT{} // C=nil
		}
		uvs := he.SubNew(u, v) // u - v  // a[i]-a[i+1]
		// A = (1 + uvs) - uvm
		A := he.SubNew(he.AddNew(uvs, he.KInt(1)), uvm) // a[i]-a[i+1]+1 - a[i]*a[i+1]

		// B = (1 - uvs) - uvm
		B := he.SubNew(he.AddNew(he.MulNew(uvs, he.KInt(-1)), he.KInt(1)), uvm) // 1 -a[i]+a[i+1]- a[i]*a[i+1]

		C := he.MulNew(A, B) //
		return S, C
	}

	isNil := func(x he.CT) bool { return x.Ct == nil && x.K == nil }

	combine := func(lS, lC, rS, rC he.CT) (he.CT, he.CT) {
		he.Relin(lC)
		he.Rescale(lC)

		if isNil(rC) {
			S := he.AddNew(lS, he.MulNew(lC, rS))
			he.Relin(S)
			he.Rescale(S)
			return S, he.CT{} // C=nil
		}

		he.Relin(rS)
		he.Rescale(rS)

		he.Relin(rC)
		he.Rescale(rC)

		tmp := he.MulNew(lC, rS)
		S := he.AddNew(lS, tmp)

		C := he.MulNew(lC, rC)
		return S, C
	}

	var solve func(lo, hi int, isTailSeg bool) (he.CT, he.CT)
	solve = func(lo, hi int, isTailSeg bool) (he.CT, he.CT) {
		L := hi - lo
		if L == 1 {
			return buildSingle(lo)
		}
		if L == 2 {
			return buildPair(lo, isTailSeg)
		}

		leftLen := largestPowerOfTwoLessThan(L)
		mid := lo + leftLen

		lS, lC := solve(lo, mid, false)
		rS, rC := solve(mid, hi, isTailSeg)

		return combine(lS, lC, rS, rC)
	}

	s, _ := solve(0, n, true)

	// ge = (2 - s*(s-1))/2

	//ge = (0.5 + s*(0.5 - s)) * 2  # x >= t (0/1 근처)
	sMinus := he.SubNew(he.KFloat(0.5), s)
	ssm := he.MulRelinNew(s, sMinus)
	he.Rescale(ssm)

	num := he.AddNew(ssm, he.KFloat(0.5)) // 2 - ssm1
	ge := he.MulNew(num, he.KInt(2))      // /2
	//he.Rescale(ge)
	//fmt.Println(ge.Ct.Level())
	return ge
}
