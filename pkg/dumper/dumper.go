// Package dumper converts `types.Value`s into `types.Op`s.
package dumper

import (
	"fmt"
	"math"
	"math/bits"

	"github.com/genkami/watson/pkg/lexer"
	"github.com/genkami/watson/pkg/types"
	"github.com/genkami/watson/pkg/vm"
)

// Dumper dumps `types.Value` as a sequence of `types.Op`s.
type Dumper struct {
	w lexer.OpWriter
}

// NewDumper creates a new Dumper.
func NewDumper(w lexer.OpWriter) *Dumper {
	return &Dumper{w: w}
}

// Dump converts v into a sequence of `types.Op`s and writes it to the underlying writer `lexer.OpWriter`.
func (d *Dumper) Dump(v *types.Value) error {
	switch v.Kind {
	case types.Int:
		return d.dumpInt(uint64(v.Int))
	case types.Uint:
		return d.dumpUint(v.Uint)
	case types.Float:
		return d.dumpFloat(v.Float)
	case types.String:
		return d.dumpString(v.String)
	case types.Object:
		return d.dumpObject(v.Object)
	case types.Array:
		return d.dumpArray(v.Array)
	case types.Bool:
		return d.dumpBool(v.Bool)
	case types.Nil:
		return d.dumpNil()
	default:
		panic(fmt.Errorf("unknown kind: %d", v.Kind))
	}
}

// dumpInt writes out a number from the most-significant to the least-significant bit.
func (d *Dumper) dumpInt(n uint64) error {
	var err error
	err = d.w.Write(vm.Inew)
	if err != nil {
		return err
	}

	if n == 0 {
		return nil
	}

	msb := 63 - bits.LeadingZeros64(n)

	// This bit is guaranteed to be one 1 and we can write it out directly.
	err = d.w.Write(vm.Iinc)
	if err != nil {
		return err
	}

	// Now we start checking from the next bit.
	for i := msb - 1; i >= 0; i-- {
		err = d.w.Write(vm.Ishl)
		if err != nil {
			return err
		}

		mask := uint64(1) << i
		if (n & mask) != 0 {
			err = d.w.Write(vm.Iinc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *Dumper) dumpUint(n uint64) error {
	var err error
	err = d.dumpInt(n)
	if err != nil {
		return err
	}
	err = d.w.Write(vm.Itou)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dumper) dumpFloat(x float64) error {
	var err error
	if math.IsNaN(x) {
		return d.w.Write(vm.Fnan)
	} else if math.IsInf(x, 1) {
		return d.w.Write(vm.Finf)
	} else if math.IsInf(x, -1) {
		err = d.w.Write(vm.Finf)
		if err != nil {
			return err
		}
		err = d.w.Write(vm.Fneg)
		if err != nil {
			return err
		}
	}
	err = d.dumpInt(math.Float64bits(x))
	if err != nil {
		return err
	}
	err = d.w.Write(vm.Itof)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dumper) dumpString(s []byte) error {
	var err error
	err = d.w.Write(vm.Snew)
	if err != nil {
		return err
	}
	for _, c := range s {
		err = d.dumpInt(uint64(c))
		if err != nil {
			return err
		}
		err = d.w.Write(vm.Sadd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Dumper) dumpObject(obj map[string]*types.Value) error {
	var err error
	err = d.w.Write(vm.Onew)
	if err != nil {
		return err
	}
	for k, v := range obj {
		err = d.dumpString([]byte(k))
		if err != nil {
			return err
		}
		err = d.Dump(v)
		if err != nil {
			return err
		}
		err = d.w.Write(vm.Oadd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Dumper) dumpArray(arr []*types.Value) error {
	var err error
	err = d.w.Write(vm.Anew)
	if err != nil {
		return err
	}
	for _, v := range arr {
		err = d.Dump(v)
		if err != nil {
			return err
		}
		err = d.w.Write(vm.Aadd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Dumper) dumpBool(b bool) error {
	var err error
	err = d.w.Write(vm.Bnew)
	if err != nil {
		return err
	}
	if b {
		err = d.w.Write(vm.Bneg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Dumper) dumpNil() error {
	return d.w.Write(vm.Nnew)
}
