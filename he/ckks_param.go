package he

type ParamsCustom struct {
	LogN            int // ring dimension
	H               int // Dense hamming
	h               int // sparse hamming
	LogMessageRatio int

	Logq0 int // base modulus

	LogSlotsToCoeffs int // slots to coefficients modulus
	SlotsToCoeffsLv  int // slots to coefficients Level

	LogDefaultScale int // circuit modulus
	CircuitLv       int // circuit level

	Logbase        int // Base for Bootstrapping
	LogLookUpTable int // Look-Up table modulus
	HermiteOrder   int // Order of hermite interpolation

	LogEvalMod int // Evaluation Mod1 modulus
	EvalModLv  int // Evaluation Mod1 Level

	LogCoeffsToSlots int // coefficients to slots modulus
	CoeffsToSlotsLv  int // coefficients to slots level

	LogP int // special modulus
	PLv  int // special modulus level
}

var DefaultParam = ParamsCustom{
	LogN:            16,
	H:               192,
	h:               32,
	LogMessageRatio: 10,

	Logq0: 55,

	LogSlotsToCoeffs: 39,
	SlotsToCoeffsLv:  3,

	LogDefaultScale: 45,
	CircuitLv:       11,

	Logbase:        -1, // -1 means that this parameter not support integer/batch-bit bootstrapping !!!
	LogLookUpTable: -1,
	HermiteOrder:   -1,

	LogEvalMod: 60,
	EvalModLv:  8,

	LogCoeffsToSlots: 56,
	CoeffsToSlotsLv:  3,

	LogP: 61,
	PLv:  5,
}

var Parameter19_H32768h32_LT33_B4H1 = ParamsCustom{
	LogN:            16,
	H:               32768,
	h:               32,
	LogMessageRatio: 0, // Don't Change!

	Logq0: 45,

	LogSlotsToCoeffs: 39,
	SlotsToCoeffsLv:  3,

	LogDefaultScale: 45,
	CircuitLv:       11, // Max level = 19

	Logbase:        4, // -1 means that this parameter not support integer/batch-bit bootstrapping !!!
	LogLookUpTable: 45,
	HermiteOrder:   1,

	LogEvalMod: 45,
	EvalModLv:  8,

	LogCoeffsToSlots: 42,
	CoeffsToSlotsLv:  3,

	LogP: 46,
	PLv:  5,
}

var Parameter19_H32768h32_LT33_B4H1_2 = ParamsCustom{
	LogN:            16,
	H:               32768,
	h:               32,
	LogMessageRatio: 0, // Don't Change!

	Logq0: 45,

	LogSlotsToCoeffs: 39,
	SlotsToCoeffsLv:  3,

	LogDefaultScale: 45,
	CircuitLv:       11, // Max level = 19

	Logbase:        1, // -1 means that this parameter not support integer/batch-bit bootstrapping !!!
	LogLookUpTable: 45,
	HermiteOrder:   1,

	LogEvalMod: 45,
	EvalModLv:  8,

	LogCoeffsToSlots: 42,
	CoeffsToSlotsLv:  3,

	LogP: 46,
	PLv:  5,
}

var Parameter18_H32768h32_LT33_B5H1 = ParamsCustom{
	LogN:            16,
	H:               32768,
	h:               32,
	LogMessageRatio: 0, // Don't Change!

	Logq0: 45,

	LogSlotsToCoeffs: 39,
	SlotsToCoeffsLv:  3,

	LogDefaultScale: 45,
	CircuitLv:       11, // Max level = 18

	Logbase:        5, // -1 means that this parameter not support integer/batch-bit bootstrapping !!!
	LogLookUpTable: 45,
	HermiteOrder:   1,

	LogEvalMod: 45,
	EvalModLv:  8,

	LogCoeffsToSlots: 42,
	CoeffsToSlotsLv:  3,

	LogP: 46,
	PLv:  5,
}

var Parameter15_H32768h32_LT33_B6H1 = ParamsCustom{
	LogN:            16,
	H:               32768,
	h:               32,
	LogMessageRatio: 0, // Don't Change!

	Logq0: 43,

	LogSlotsToCoeffs: 37,
	SlotsToCoeffsLv:  3,

	LogDefaultScale: 43,
	CircuitLv:       11, // Max level = 15

	Logbase:        6, // -1 means that this parameter not support integer/batch-bit bootstrapping !!!
	LogLookUpTable: 43,
	HermiteOrder:   1,

	LogEvalMod: 43,
	EvalModLv:  8,

	LogCoeffsToSlots: 40,
	CoeffsToSlotsLv:  3,

	LogP: 44,
	PLv:  5,
}

var Parameter14_H32768h32_LT33_B7H2 = ParamsCustom{
	LogN:            16,
	H:               32768,
	h:               32,
	LogMessageRatio: 0, // Don't Change!

	Logq0: 43,

	LogSlotsToCoeffs: 37,
	SlotsToCoeffsLv:  3,

	LogDefaultScale: 43,
	CircuitLv:       11, // Max level = 13

	Logbase:        7, // -1 means that this parameter not support integer/batch-bit bootstrapping !!!
	LogLookUpTable: 43,
	HermiteOrder:   2,

	LogEvalMod: 43,
	EvalModLv:  8,

	LogCoeffsToSlots: 40,
	CoeffsToSlotsLv:  3,

	LogP: 44,
	PLv:  5,
}

var Parameter12_H32768h32_LT33_B8H2 = ParamsCustom{
	LogN:            16,
	H:               32768,
	h:               32,
	LogMessageRatio: 0, // Don't Change!

	Logq0: 44,

	LogSlotsToCoeffs: 39,
	SlotsToCoeffsLv:  3,

	LogDefaultScale: 44,
	CircuitLv:       11, // Max level = 12

	Logbase:        8, // -1 means that this parameter not support integer/batch-bit bootstrapping !!!
	LogLookUpTable: 44,
	HermiteOrder:   2,

	LogEvalMod: 44,
	EvalModLv:  8,

	LogCoeffsToSlots: 42,
	CoeffsToSlotsLv:  3,

	LogP: 45,
	PLv:  4,
}
