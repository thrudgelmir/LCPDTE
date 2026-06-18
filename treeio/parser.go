package treeio

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

type RawInput struct {
	N int
	F int
	X []float32 // row-major
}

type RawModel struct {
	BaseScore      float64
	NumFeature     int
	NumOutputGroup int
	TreeInfo       []int
	Trees          []RawTree
}

type RawTree struct {
	Left        []int32
	Right       []int32
	SplitIndex  []int32
	SplitCond   []float32
	DefaultLeft []bool
	SplitType   []uint8
}

type ParsedBundle struct {
	Model RawModel
	Input RawInput
	Pred  []float32 // 없으면 nil
}

func ParseBundle(dir string, name string) (ParsedBundle, int) {
	model := ParseXGBModelJSON(filepath.Join(dir, "xgb_model_"+name+".json"))

	var in RawInput
	if fileExists(filepath.Join(dir, "x_test.bin")) {
		in = LoadXTestWithHeader(filepath.Join(dir, "x_test.bin"))
	} else {
		panic("no testdata found")
	}

	var pred []float32
	if fileExists(filepath.Join(dir, "pred_"+name+".bin")) {
		pred = LoadPredWithHeader(filepath.Join(dir, "pred_"+name+".bin"))
	} else {
		pred = nil
	}

	return ParsedBundle{Model: model, Input: in, Pred: pred}, 10
}

// -------- model(json) --------

type xgbRoot struct {
	Learner struct {
		LearnerModelParam map[string]string `json:"learner_model_param"`
		GradientBooster   struct {
			Model struct {
				TreeInfo []int             `json:"tree_info"`
				Trees    []json.RawMessage `json:"trees"`
			} `json:"model"`
		} `json:"gradient_booster"`
	} `json:"learner"`
}

func ParseXGBModelJSON(path string) RawModel {
	raw := mustReadFile(path)

	var root xgbRoot
	must(json.Unmarshal(raw, &root))

	lmp := root.Learner.LearnerModelParam
	numFeature := atoi0(lmp["num_feature"])
	numOutputGroup := atoi0(lmp["num_output_group"])

	baseScore := 0.0
	if s, ok := lmp["base_score"]; ok {
		baseScore = parseBaseScore(s)
	}

	treesRaw := root.Learner.GradientBooster.Model.Trees
	trees := make([]RawTree, len(treesRaw))
	for i := range treesRaw {
		trees[i] = parseOneTree(treesRaw[i])
	}

	return RawModel{
		BaseScore:      baseScore,
		NumFeature:     numFeature,
		NumOutputGroup: numOutputGroup,
		TreeInfo:       root.Learner.GradientBooster.Model.TreeInfo,
		Trees:          trees,
	}
}

func parseBaseScore(s string) float64 {
	if s == "" {
		return 0
	}
	var arr []float64
	if err := json.Unmarshal([]byte(s), &arr); err == nil && len(arr) > 0 {
		return arr[0]
	}
	v, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return v
	}
	return 0
}

func parseOneTree(raw json.RawMessage) RawTree {
	var m map[string]json.RawMessage
	must(json.Unmarshal(raw, &m))

	left := parseI32Arr(m["left_children"])
	right := parseI32Arr(m["right_children"])
	sidx := parseI32Arr(m["split_indices"])
	scond := parseF32Arr(m["split_conditions"])
	defLeft := parseBoolLikeArr(m["default_left"], len(left))
	sType := parseU8ArrOrZeros(m["split_type"], len(left))

	return RawTree{
		Left:        left,
		Right:       right,
		SplitIndex:  sidx,
		SplitCond:   scond,
		DefaultLeft: defLeft,
		SplitType:   sType,
	}
}

func parseI32Arr(raw json.RawMessage) []int32 {
	if len(raw) == 0 {
		return nil
	}
	var out []int32
	must(json.Unmarshal(raw, &out))
	return out
}

func parseF32Arr(raw json.RawMessage) []float32 {
	if len(raw) == 0 {
		return nil
	}
	var out []float32
	must(json.Unmarshal(raw, &out))
	return out
}

func parseBoolLikeArr(raw json.RawMessage, n int) []bool {
	if len(raw) == 0 {
		return make([]bool, n)
	}
	var b []bool
	if err := json.Unmarshal(raw, &b); err == nil {
		return b
	}
	var x []int32
	must(json.Unmarshal(raw, &x))
	out := make([]bool, len(x))
	for i := range x {
		out[i] = (x[i] != 0)
	}
	return out
}

func parseU8ArrOrZeros(raw json.RawMessage, n int) []uint8 {
	if len(raw) == 0 {
		return make([]uint8, n)
	}
	var x []int32
	must(json.Unmarshal(raw, &x))
	out := make([]uint8, len(x))
	for i := range x {
		out[i] = uint8(x[i])
	}
	return out
}

// -------- input/pred(bin) --------
// x_test.bin: [int32 N][int32 F][float32 N*F]
// pred.bin:   [int32 N][float32 N]

func LoadXTestWithHeader(path string) RawInput {
	b := mustReadFile(path)
	r := bytes.NewReader(b)

	var n32 int32
	var f32 int32
	must(binary.Read(r, binary.LittleEndian, &n32))
	must(binary.Read(r, binary.LittleEndian, &f32))

	N := int(n32)
	F := int(f32)

	x := make([]float32, N*F)
	must(binary.Read(r, binary.LittleEndian, &x))
	return RawInput{N: N, F: F, X: x}
}

func LoadPredWithHeader(path string) []float32 {
	b := mustReadFile(path)
	r := bytes.NewReader(b)

	var n32 int32
	must(binary.Read(r, binary.LittleEndian, &n32))

	p := make([]float32, int(n32))
	must(binary.Read(r, binary.LittleEndian, &p))
	return p
}

// fallback (no header)
func LoadXTestNoHeader(path string, F int) RawInput {
	b := mustReadFile(path)
	r := bytes.NewReader(b)

	x := make([]float32, len(b)/4)
	must(binary.Read(r, binary.LittleEndian, &x))

	N := len(x) / F
	return RawInput{N: N, F: F, X: x}
}

func LoadPredNoHeader(path string) []float32 {
	b := mustReadFile(path)
	r := bytes.NewReader(b)

	p := make([]float32, len(b)/4)
	must(binary.Read(r, binary.LittleEndian, &p))
	return p
}

// -------- utils --------

func mustReadFile(path string) []byte {
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return b
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
func atoi0(s string) int {
	if s == "" {
		return 0
	}
	v, _ := strconv.Atoi(s)
	return v
}
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
