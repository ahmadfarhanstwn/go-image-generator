package apt

import (
	"math"
	"math/rand"
	"reflect"
	"strconv"

	"github.com/ahmadfarhanstwn/noise"
)

type Node interface {
	Eval(x, y float32) float32
	String() string
	SetParent(node Node)
	SetChildren(children []Node)
	AddRandom(node Node)
	AddLeaf(nodeLeaf Node) bool
	CountNode() int
	GetChildren() []Node
	GetParent() Node
}

type BaseNode struct {
	Parent Node
	Children []Node
}

func CopyTree(node, parent Node) Node {
	copy := reflect.New(reflect.ValueOf(node).Elem().Type()).Interface().(Node)
	switch n := node.(type) {
	case *OpConst:
		copy.(*OpConst).value = n.value
	}
	copy.SetParent(parent)
	copyChildren := make([]Node, len(node.GetChildren()))
	copy.SetChildren(copyChildren)
	for i := range copyChildren {
		copyChildren[i] = CopyTree(node.GetChildren()[i], copy)
	}
	return copy
}

func ReplaceNode(old, new Node) {
	oldParent := old.GetParent()
	if oldParent != nil {
		for i, child := range oldParent.GetChildren(){
			if child == old {
				oldParent.GetChildren()[i] = new
				break
			}
		}
	}
	new.SetParent(oldParent)
}

func GetNthChildren(node Node, n, count int) (Node, int) {
	if n == count {
		return node, count
	}
	var result Node
	for _, child := range node.GetChildren(){
		count++
		result, count = GetNthChildren(child, n, count)
		if result != nil {
			return result, count
		}
	}
	return nil, count
} 

func Mutate(node Node) Node {
	r := rand.Intn(23)
	var MutateNode Node
	if r <= 19 {
		MutateNode = GetRandomNodeOpt()
	} else {
		MutateNode = GetRandomLeafNode()
	}

	if node.GetParent() != nil {
		for i, parentChild := range node.GetParent().GetChildren() {
			if parentChild == node {
				node.GetParent().GetChildren()[i] = MutateNode
			}
		}
	}

	for i, child := range node.GetChildren() {
		if i >= len(MutateNode.GetChildren()) {
			break
		}
		MutateNode.GetChildren()[i] = child
		child.SetParent(MutateNode)
	}

	for i, child := range MutateNode.GetChildren() {
		if child == nil {
			leaf := GetRandomLeafNode()
			leaf.SetParent(MutateNode)
			MutateNode.GetChildren()[i] = leaf
		}
	}

	MutateNode.SetParent(node.GetParent())
	return MutateNode
}

func (base *BaseNode) GetParent() Node {
	return base.Parent
}

func (base *BaseNode) GetChildren() []Node {
	return base.Children
}

func (base *BaseNode) Eval(x, y float32) float32 {
	panic("tried to eval base node")
}

func (base *BaseNode) String() string {
	panic("tried to string basenode")
}

func (base *BaseNode) SetParent(parent Node) {
	base.Parent = parent
}

func (base *BaseNode) SetChildren(children []Node) {
	base.Children = children
}

func (base *BaseNode) AddRandom(node Node) {
	index := rand.Intn(len(base.Children))
	if base.Children[index] == nil {
		node.SetParent(base)
		base.Children[index] = node
	} else {
		base.Children[index].AddRandom(node)
	}
}

func (base *BaseNode) AddLeaf(leafNode Node) bool {
	for i, node := range base.Children {
		if node == nil {
			leafNode.SetParent(node)
			base.Children[i] = leafNode
			return true
		} else if base.Children[i].AddLeaf(leafNode) {
			return true
		}
	}
	return false
}

func (base *BaseNode) CountNode() int {
	count := 1
	for _, child := range base.Children {
		count += child.CountNode()
	}
	return count
}

type OpPlus struct {
	BaseNode
}

func NewOpPlus() *OpPlus {
	return &OpPlus{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpPlus) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) + op.Children[1].Eval(x, y)
}

func (op *OpPlus) String() string {
	return "( + " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpMinus struct {
	BaseNode
}

func NewOpMinus() *OpMinus {
	return &OpMinus{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpMinus) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) - op.Children[1].Eval(x, y)
}

func (op *OpMinus) String() string {
	return "( - " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpMultiplies struct {
	BaseNode
}

func NewOpMultiplies() *OpMultiplies {
	return &OpMultiplies{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpMultiplies) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) * op.Children[1].Eval(x, y)
}

func (op *OpMultiplies) String() string {
	return "( * " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpDivide struct {
	BaseNode
}

func NewOpDivide() *OpDivide {
	return &OpDivide{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpDivide) Eval(x, y float32) float32 {
	return op.Children[0].Eval(x, y) / op.Children[1].Eval(x, y)
}

func (op *OpDivide) String() string {
	return "( / " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpAtan2 struct {
	BaseNode
}

func NewOpAtan2() *OpAtan2 {
	return &OpAtan2{BaseNode{nil, make([]Node, 2)}}
}

func (op *OpAtan2) Eval(x, y float32) float32 {
	return float32(math.Atan2(float64(y),float64(x)))
}

func (op *OpAtan2) String() string {
	return "( atan2 " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpSin struct {
	BaseNode
}

func NewOpSin() *OpSin {
	return &OpSin{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpSin) Eval(x, y float32) float32 {
	return float32(math.Sin(float64(op.Children[0].Eval(x,y))))
}

func (op *OpSin) String() string {
	return "( sin " + op.Children[0].String() + " )"
}

type OpCos struct {
	BaseNode
}

func NewOpCos() *OpCos {
	return &OpCos{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpCos) Eval(x, y float32) float32 {
	return float32(math.Cos(float64(op.Children[0].Eval(x,y))))
}

func (op *OpCos) String() string {
	return "( cos " + op.Children[0].String() + " )"
}

type OpAtan struct {
	BaseNode
}

func NewOpAtan() *OpAtan {
	return &OpAtan{BaseNode{nil, make([]Node, 1)}}
}

func (op *OpAtan) Eval(x, y float32) float32 {
	return float32(math.Atan(float64(op.Children[0].Eval(x,y))))
}

func (op *OpAtan) String() string {
	return "( atan " + op.Children[0].String() + " )"
}

type opNoise struct {
	BaseNode
}

func NewOpNoise() *opNoise {
	return &opNoise{BaseNode{nil, make([]Node, 2)}}
}

func (op *opNoise) Eval(x, y float32) float32 {
	return 80*noise.Snoise2(op.Children[0].Eval(x,y),op.Children[1].Eval(x,y))-2.0
}

func (op *opNoise) String() string {
	return "( snoise2 " + op.Children[0].String() + " " + op.Children[1].String() + " )"
}

type OpSquare struct {
	BaseNode
}

func NewOpSquare() *OpSquare {
	return &OpSquare{BaseNode{nil, make([]Node, 1)}}
}

func (opsquare *OpSquare) Eval(x, y float32) float32 {
	val := opsquare.Children[0].Eval(x,y)
	return val*val
}

func (opsquare *OpSquare) String() string {
	return "( square " + opsquare.Children[0].String() + " )"
}

type OpNegate struct {
	BaseNode
}

func NewOpNegate() *OpNegate {
	return &OpNegate{BaseNode{nil, make([]Node, 1)}}
}

func (opnegate *OpNegate) Eval(x, y float32) float32 {
	return -opnegate.Children[0].Eval(x,y)
}

func (opnegate *OpNegate) String() string {
	return "( negate " + opnegate.Children[0].String() + " )"
}

type OpCeil struct {
	BaseNode
}

func NewOpCeil() *OpCeil {
	return &OpCeil{BaseNode{nil, make([]Node, 1)}}
}

func (opceil *OpCeil) Eval(x, y float32) float32 {
	return float32(math.Ceil(float64(opceil.Children[0].Eval(x,y))))
}

func (opceil *OpCeil) String() string {
	return "( ceil " + opceil.Children[0].String() + " )"
}

type OpFloor struct {
	BaseNode
}

func NewOpFloor() *OpFloor {
	return &OpFloor{BaseNode{nil, make([]Node, 1)}}
}

func (opfloor *OpFloor) Eval(x, y float32) float32 {
	return float32(math.Floor(float64(opfloor.Children[0].Eval(x,y))))
}

func (opfloor *OpFloor) String() string {
	return "( floor " + opfloor.Children[0].String() + " )"
}

type OpAbs struct {
	BaseNode
}

func NewOpAbs() *OpAbs {
	return &OpAbs{BaseNode{nil, make([]Node, 1)}}
}

func (opabs *OpAbs) Eval(x, y float32) float32 {
	return float32(math.Abs(float64(opabs.Children[0].Eval(x,y))))
}

func (opabs *OpAbs) String() string {
	return "( abs " + opabs.Children[0].String() + " )"
}

type OpFbm struct {
	BaseNode
}

func NewOpFbm() *OpFbm {
	return &OpFbm{BaseNode{nil, make([]Node, 3)}}
}

func (opfbm *OpFbm) Eval(x, y float32) float32 {
	return noise.Fbm2(opfbm.Children[0].Eval(x,y), opfbm.Children[1].Eval(x,y), opfbm.Children[2].Eval(x,y), 0.5, 2, 3)
}

func (opfbm *OpFbm) String() string {
	return "( fbm " + opfbm.Children[0].String() + " " + opfbm.Children[1].String() + " " + opfbm.Children[2].String() + " )"
}

type OpTurbulence struct {
	BaseNode
}

func NewTurbulence() *OpFbm {
	return &OpFbm{BaseNode{nil, make([]Node, 3)}}
}

func (opturbulence *OpTurbulence) Eval(x, y float32) float32 {
	return noise.Turbulence(opturbulence.Children[0].Eval(x,y), opturbulence.Children[1].Eval(x,y), opturbulence.Children[2].Eval(x,y), 0.5, 2, 3)
}

func (opturbulence *OpTurbulence) String() string {
	return "( turbulence " + opturbulence.Children[0].String() + " " + opturbulence.Children[1].String() + " " + opturbulence.Children[2].String() + " )"
}

type OpX struct {
	BaseNode
}

func NewOpX() *OpX {
	return &OpX{BaseNode{nil, make([]Node, 0)}}
}

func (opx *OpX) Eval(x, y float32) float32 {
	return x
}

func (opx *OpX) String() string {
	return "x"
}

type OpY struct {
	BaseNode
}

func NewOpY() *OpY {
	return &OpY{BaseNode{nil, make([]Node, 0)}}
}

func (opy *OpY) Eval(x, y float32) float32 {
	return y
}

func (opy *OpY) String() string {
	return "y"
}

type OpConst struct {
	BaseNode
	value float32
}

func NewOpConst() *OpConst {
	return &OpConst{BaseNode{nil, make([]Node, 0)}, rand.Float32()*2-1}
}

func (opconst *OpConst) Eval(x, y float32) float32 {
	return opconst.value
}

func (opconst *OpConst) String() string {
	return strconv.FormatFloat(float64(opconst.value),'f',9,32)
}

type OpPict struct {
	BaseNode
}

func NewOpPict() *OpPict {
	return &OpPict{BaseNode{nil, make([]Node, 3)}}
}

func (oppict *OpPict) Eval(x, y float32) float32 {
	panic("tried to eval root of pict")
}

func (oppict *OpPict) String() string {
	return "( picture\n" + oppict.Children[0].String() + "\n" + oppict.Children[1].String() + "\n" + oppict.Children[2].String() + " )"
}

func GetRandomNodeOpt() Node {
	r := rand.Intn(16)
	switch r {
	case 0:
		return NewOpPlus()
	case 1:
		return NewOpMinus()
	case 2:
		return NewOpMultiplies()
	case 3:
		return NewOpDivide()
	case 4:
		return NewOpAtan2()
	case 5:
		return NewOpAtan()
	case 6:
		return NewOpSin()
	case 7:
		return NewOpCos()
	case 8:
		return NewOpNoise()
	case 9:
		return NewTurbulence()
	case 10:
		return NewOpCeil()
	case 11:
		return NewOpFbm()
	case 12:
		return NewOpFloor()
	case 13:
		return NewOpNegate()
	case 14:
		return NewOpSquare()
	case 15:
		return NewOpAbs()
	}
	panic("get random operation failed")
}

func GetRandomLeafNode() Node {
	r := rand.Intn(3)
	switch r {
	case 0:
		return NewOpX()
	case 1:
		return NewOpY()
	case 2:
		return NewOpConst()
	}
	panic("get random leaf node failed")
}