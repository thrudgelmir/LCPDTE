package he

import (
	"math"
	"math/cmplx"
)

func HermiteInterpolation(n int, hermiteOrder int, LUT func(int) int) []complex128 {
	k := hermiteOrder
	numConstraints := n * (k + 1)

	A := make([][]complex128, numConstraints)
	b := make([]complex128, numConstraints)

	z := cmplx.Exp(2 * math.Pi * 1i / complex(float64(n), 0))

	row := 0
	for i := 0; i < n; i++ {
		zi := cmplx.Pow(z, complex(float64(i), 0))

		for m := 0; m <= k; m++ {
			A[row] = make([]complex128, numConstraints)

			for j := 0; j < numConstraints; j++ {
				if j < m {
					A[row][j] = 0
				} else {
					prod := complex(1.0, 0)
					for p := 0; p < m; p++ {
						prod *= complex(float64(j-p), 0)
					}
					zi_pow := cmplx.Pow(zi, complex(float64(j-m), 0))
					A[row][j] = prod * zi_pow
				}
			}

			if m == 0 {
				b[row] = complex(float64(LUT(i)), 0)
			} else {
				b[row] = 0
			}

			row++
		}
	}

	list := solveLinearSystem(A, b)
	return list
}

func HermiteInterpolation_exp2symbol(n int, deg int, base int) []complex128 {
	z := cmplx.Exp(2 * math.Pi * 1i / complex(float64(n), 0))

	A := make([][]complex128, 2*n)
	b := make([]complex128, 2*n)

	for i := 0; i < n; i++ {
		zi := cmplx.Pow(z, complex(float64(i), 0))

		A[i] = make([]complex128, 2*n)
		for j := 0; j < 2*n; j++ {
			A[i][j] = cmplx.Pow(zi, complex(float64(j), 0))
		}

		if i >= base {
			b[i] = complex(0, float64(1))
		} else if i == base-1 {
			b[i] = complex(float64(0.5), 0)
		} else {
			b[i] = complex(float64(0), 0)
		}

		A[n+i] = make([]complex128, 2*n)
		for j := 1; j < 2*n; j++ {
			A[n+i][j] = complex(float64(j), 0) * cmplx.Pow(zi, complex(float64(j-1), 0))
		}
		b[n+i] = 0
	}

	list := solveLinearSystem(A, b)
	coeffs := make([]complex128, deg+1)
	for i := 0; i < 2*n; i++ {
		coeffs[i] = list[i]
	}
	for i := 2 * n; i <= deg; i++ {
		coeffs[i] = complex(0.0, 0.0)
	}
	return coeffs
}

func HermiteInterpolation_exp2negatesymbol(n int, deg int, base int) []complex128 {
	z := cmplx.Exp(2 * math.Pi * 1i / complex(float64(n), 0))

	A := make([][]complex128, 2*n)
	b := make([]complex128, 2*n)

	for i := 0; i < n; i++ {
		zi := cmplx.Pow(z, complex(float64(i), 0))

		A[i] = make([]complex128, 2*n)
		for j := 0; j < 2*n; j++ {
			A[i][j] = cmplx.Pow(zi, complex(float64(j), 0))
		}

		if i >= base {
			b[i] = complex(0, float64(1))
		} else if i == 0 {
			b[i] = complex(float64(0.5), 0)
		} else {
			b[i] = complex(float64(0), 0)
		}

		A[n+i] = make([]complex128, 2*n)
		for j := 1; j < 2*n; j++ {
			A[n+i][j] = complex(float64(j), 0) * cmplx.Pow(zi, complex(float64(j-1), 0))
		}
		b[n+i] = 0
	}

	list := solveLinearSystem(A, b)
	coeffs := make([]complex128, deg+1)
	for i := 0; i < 2*n; i++ {
		coeffs[i] = list[i]
	}
	for i := 2 * n; i <= deg; i++ {
		coeffs[i] = complex(0.0, 0.0)
	}
	return coeffs
}

func HermiteInterpolation_exp2bin(n int, deg int, base int) []complex128 {
	z := cmplx.Exp(2 * math.Pi * 1i / complex(float64(n), 0))

	A := make([][]complex128, 2*n)
	b := make([]complex128, 2*n)

	for i := 0; i < n; i++ {
		zi := cmplx.Pow(z, complex(float64(i), 0))

		A[i] = make([]complex128, 2*n)
		for j := 0; j < 2*n; j++ {
			A[i][j] = cmplx.Pow(zi, complex(float64(j), 0))
		}

		if i == 1 {
			b[i] = complex(float64(0), 0)
		} else {
			b[i] = complex(float64(1), 0)
		}

		A[n+i] = make([]complex128, 2*n)
		for j := 1; j < 2*n; j++ {
			A[n+i][j] = complex(float64(j), 0) * cmplx.Pow(zi, complex(float64(j-1), 0))
		}
		b[n+i] = 0
	}

	list := solveLinearSystem(A, b)
	coeffs := make([]complex128, deg+1)
	for i := 0; i < 2*n; i++ {
		coeffs[i] = list[i]
	}
	for i := 2 * n; i <= deg; i++ {
		coeffs[i] = complex(0.0, 0.0)
	}
	return coeffs
}

// n = 3 : 0 0.5 i
func HermiteInterpolation_symbol2symbol() []complex128 {

	z := cmplx.Exp(2 * math.Pi * 1i / complex(float64(32), 0))

	A := make([][]complex128, 4)
	b := make([]complex128, 4)

	for i := 0; i < 2; i++ {
		zi := cmplx.Pow(z, complex(float64(i), 0))

		A[i] = make([]complex128, 4)
		for j := 0; j < 4; j++ {
			A[i][j] = cmplx.Pow(zi, complex(float64(j), 0))
		}

		if i == 0 {
			b[i] = complex(float64(0), 0)
		} else {
			b[i] = complex(float64(0.5), 0)
		}

		A[2+i] = make([]complex128, 4)
		for j := 1; j < 4; j++ {
			A[2+i][j] = complex(float64(j), 0) * cmplx.Pow(zi, complex(float64(j-1), 0))
		}
		b[2+i] = 0
	}

	list := solveLinearSystem(A, b)
	//fmt.Println(len(list))
	coeffs := make([]complex128, 4)
	for i := 0; i < 4; i++ {
		coeffs[i] = list[i]
	}

	return coeffs
}

func HermiteInterpolation_symbol2symbol_imag() []complex128 {

	z := cmplx.Exp(2 * math.Pi * 1i / complex(float64(16), 0))

	A := make([][]complex128, 4)
	b := make([]complex128, 4)

	for i := 0; i < 2; i++ {
		zi := cmplx.Pow(z, complex(float64(i), 0))

		A[i] = make([]complex128, 4)
		for j := 0; j < 4; j++ {
			A[i][j] = cmplx.Pow(zi, complex(float64(j), 0))
		}

		if i == 0 {
			b[i] = complex(float64(0), 0)
		} else {
			b[i] = complex(float64(0), 1.0)
		}

		A[2+i] = make([]complex128, 4)
		for j := 1; j < 4; j++ {
			A[2+i][j] = complex(float64(j), 0) * cmplx.Pow(zi, complex(float64(j-1), 0))
		}
		b[2+i] = 0
	}

	list := solveLinearSystem(A, b)
	//fmt.Println(len(list))
	coeffs := make([]complex128, 4)
	for i := 0; i < 4; i++ {
		coeffs[i] = list[i]
	}

	return coeffs
}

func Decomp(x uint32, chunksize int) []uint32 {
	len := 32 / chunksize
	bits := make([]uint32, len)
	for i := 0; i < len; i++ {
		bits[i] = (x >> uint32(chunksize)) & (0b1111)
	}
	return bits
}

func Int2Exp(x uint32, chunksize int) complex128 {
	return cmplx.Exp(2 * math.Pi * 1i * complex(float64(x), 0) / complex(float64(chunksize), 0))
}

func solveLinearSystem(A [][]complex128, b []complex128) []complex128 {
	n := len(b)
	for i := 0; i < n; i++ {
		maxRow := i
		for k := i + 1; k < n; k++ {
			if cmplx.Abs(A[k][i]) > cmplx.Abs(A[maxRow][i]) {
				maxRow = k
			}
		}
		A[i], A[maxRow] = A[maxRow], A[i]
		b[i], b[maxRow] = b[maxRow], b[i]

		pivot := A[i][i]
		for j := i; j < n; j++ {
			A[i][j] /= pivot
		}
		b[i] /= pivot

		for k := i + 1; k < n; k++ {
			factor := A[k][i]
			for j := i; j < n; j++ {
				A[k][j] -= factor * A[i][j]
			}
			b[k] -= factor * b[i]
		}
	}

	x := make([]complex128, n)
	for i := n - 1; i >= 0; i-- {
		x[i] = b[i]
		for j := i + 1; j < n; j++ {
			x[i] -= A[i][j] * x[j]
		}
	}

	return x
}
