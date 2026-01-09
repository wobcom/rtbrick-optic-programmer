package rtbrick

const (
	Bit0 Bitmask = 1 << iota
	Bit1
	Bit2
	Bit3
	Bit4
	Bit5
	Bit6
	Bit7
)

type Bitmask uint8

func (b Bitmask) Has(flag Bitmask) bool { return b&flag != 0 }
func (b *Bitmask) Set(flag Bitmask)     { *b |= flag }
func (b *Bitmask) Clear(flag Bitmask)   { *b &= ^flag }
func (b *Bitmask) Toggle(flag Bitmask)  { *b ^= flag }
