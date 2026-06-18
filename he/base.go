package he

import (
	"fmt"
	"math/big"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/dft"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/mod1"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
)

var (
	Params ckks.Parameters

	BtpParams bootstrapping.Parameters

	SlotsToCoeffsParams dft.MatrixLiteral
	CoeffsToSlotsParams dft.MatrixLiteral
	Mod1Params          mod1.ParametersLiteral

	SK *rlwe.SecretKey
	PK *rlwe.PublicKey

	Encoder   *ckks.Encoder
	Encryptor *rlwe.Encryptor
	Decryptor *rlwe.Decryptor
	Eval      *bootstrapping.Evaluator

	MaxLv int

	Logbase      int
	Base         int
	HermiteOrder int

	CtSize int
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func Preset(short bool, opt int, clv int) {

	debug := false

	//var params = DefaultParam
	var params = Parameter19_H32768h32_LT33_B4H1
	if opt == 1 {
		params = Parameter19_H32768h32_LT33_B4H1_2
	}
	params.CircuitLv = clv

	Logbase = params.Logbase
	Base = 1 << Logbase
	HermiteOrder = params.HermiteOrder

	LogN := params.LogN

	if short {
		LogN -= 3
	}

	q0 := []int{params.Logq0} // ScaleDown & ModUp

	qiSlotsToCoeffs := []int{} // SlotsToCoeffs
	StCLevels := []int{}
	for i := 0; i < params.SlotsToCoeffsLv; i++ {
		qiSlotsToCoeffs = append(qiSlotsToCoeffs, params.LogSlotsToCoeffs)
		StCLevels = append(StCLevels, 1)
	}

	qiCircuitSlots := []int{} // Circuit in slots
	for i := 0; i < params.CircuitLv; i++ {
		qiCircuitSlots = append(qiCircuitSlots, params.LogDefaultScale)
	}

	qiLookUpTables := []int{} // Look-Up Table
	for i := 0; i < params.HermiteOrder+params.Logbase; i++ {
		qiLookUpTables = append(qiLookUpTables, params.LogLookUpTable)
	}

	qiEvalMod := []int{} // EvalMod1
	for i := 0; i < params.EvalModLv; i++ {
		qiEvalMod = append(qiEvalMod, params.LogEvalMod)
	}

	CtSLevels := []int{}
	qiCoeffsToSlots := []int{} // CoeffsToSlots
	for i := 0; i < params.CoeffsToSlotsLv; i++ {
		qiCoeffsToSlots = append(qiCoeffsToSlots, params.LogCoeffsToSlots)
		CtSLevels = append(CtSLevels, 1)
	}

	LogQ := append([]int{}, q0...)
	LogQ = append(LogQ, qiSlotsToCoeffs...)
	LogQ = append(LogQ, qiCircuitSlots...)
	MaxLv = len(LogQ) - 1
	LogQ = append(LogQ, qiLookUpTables...)
	LogQ = append(LogQ, qiEvalMod...)
	LogQ = append(LogQ, qiCoeffsToSlots...)

	LogP := []int{}
	for i := 0; i < params.PLv; i++ {
		LogP = append(LogP, params.LogP)
	}

	var err error
	Params, err = ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN:            LogN,
		LogQ:            LogQ,
		LogP:            LogP,
		LogDefaultScale: params.LogDefaultScale,
		Xs:              ring.Ternary{H: params.H},
	})
	must(err)

	// === Bootstrapping sub-params ===
	CoeffsToSlotsParams = dft.MatrixLiteral{
		Type:         dft.HomomorphicEncode,
		Format:       dft.RepackImagAsReal,
		LogSlots:     Params.LogMaxSlots(),
		LevelQ:       Params.MaxLevelQ(),
		LevelP:       Params.MaxLevelP(),
		LogBSGSRatio: 1,
		Levels:       CtSLevels,
	}

	Mod1Params = mod1.ParametersLiteral{
		LevelQ:          Params.MaxLevel() - CoeffsToSlotsParams.Depth(true),
		LogScale:        params.LogEvalMod,
		Mod1Type:        mod1.CosDiscrete,
		Mod1Degree:      30,
		DoubleAngle:     3,
		K:               16,
		LogMessageRatio: params.LogMessageRatio,
		Mod1InvDegree:   0,
	}

	Scaling := big.NewFloat(1.0)
	if params.Logbase != -1 {
		base := 1 << params.Logbase
		Scaling = big.NewFloat(1.0 / float64(base))
	}
	SlotsToCoeffsParams = dft.MatrixLiteral{
		Type:         dft.HomomorphicDecode,
		LogSlots:     Params.LogMaxSlots(),
		LogBSGSRatio: 1,
		LevelP:       Params.MaxLevelP(),
		Levels:       StCLevels,
		Scaling:      Scaling,
	}
	SlotsToCoeffsParams.LevelQ = len(SlotsToCoeffsParams.Levels)

	BtpParams = bootstrapping.Parameters{
		ResidualParameters:      Params,
		BootstrappingParameters: Params,
		SlotsToCoeffsParameters: SlotsToCoeffsParams,
		Mod1ParametersLiteral:   Mod1Params,
		CoeffsToSlotsParameters: CoeffsToSlotsParams,
		EphemeralSecretWeight:   params.h,
		CircuitOrder:            bootstrapping.DecodeThenModUp,
	}
	fmt.Printf("Bootstrapping parameters: logN=%d, logSlots=%d, H(%d; %d), sigma=%f, logQP=%f, levels=%d, scale=2^%d\n",
		BtpParams.BootstrappingParameters.LogN(),
		BtpParams.BootstrappingParameters.LogMaxSlots(),
		BtpParams.BootstrappingParameters.XsHammingWeight(),
		BtpParams.EphemeralSecretWeight,
		BtpParams.BootstrappingParameters.Xe(),
		BtpParams.BootstrappingParameters.LogQP(),
		BtpParams.BootstrappingParameters.QCount(),
		BtpParams.BootstrappingParameters.LogDefaultScale())

	// debug
	if debug {
		fmt.Println("Base : ", q0)
		fmt.Println("StC : ", qiSlotsToCoeffs)
		fmt.Println("Circuit : ", qiCircuitSlots)
		fmt.Println("LookUpTable : ", qiLookUpTables)
		fmt.Println("EvalMod : ", qiEvalMod)
		fmt.Println("CtS : ", qiCoeffsToSlots)
		fmt.Println("Special : ", LogP)
	}

	//fmt.Printf("Maximum Level : %d, Minimum Level : %d\n", MaxLv)
	// === Keys + basic objects ===
	kgen := rlwe.NewKeyGenerator(Params)
	SK, PK = kgen.GenKeyPairNew()

	Encoder = ckks.NewEncoder(Params)
	Decryptor = rlwe.NewDecryptor(Params, SK)
	Encryptor = rlwe.NewEncryptor(Params, PK)

	fmt.Println("Generating bootstrapping evaluation keys...")
	evk, _, err := BtpParams.GenEvaluationKeys(SK)
	must(err)

	Eval, err = bootstrapping.NewEvaluator(BtpParams, evk)
	must(err)
	test := Zeros()
	CtSize = test.BinarySize()
	fmt.Println("Done")
}
