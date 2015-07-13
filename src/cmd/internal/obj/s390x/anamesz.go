package s390x

var cnamesz = []string{
	"NONE",
	"REG",
	"FREG",
	"AREG",
	"ZCON",
	"SCON",
	"UCON",
	"ADDCON",
	"ANDCON",
	"LCON",
	"DCON",
	"SACON",
	"SECON",
	"LACON",
	"LECON",
	"DACON",
	"SBRA",
	"LBRA",
	"SAUTO",
	"LAUTO",
	"ZOREG",
	"SOREG",
	"LOREG",
	"LR",
	"ANY",
	"GOK",
	"ADDR",
	"TEXTSIZE",
	"NCLASS",
}

// const ( // comments from func aclass in asmz.go
//     C_NONE = iota
//     C_REG // general-purpose register
//     C_FREG // floating-point register
//     C_AREG // access register
//     C_ZCON // constant == 0
//     C_SCON // 0 <= constant <= 0x7fff (positive int16)
//     C_UCON // constant & 0xffff == 0 (int32 or uint32)
//     C_ADDCON // 0 > constant >= -0x8000 (negative int16)
//     C_ANDCON // constant <= 0xffff
//     C_LCON // constant (int32 or uint32)
//     C_DCON // constant (int64 or uint64)
//     C_SACON // computed address, 16-bit displacement, possibly SP-relative
//     C_SECON // computed address, 16-bit displacement, possibly SB-relative, unused?
//     C_LACON // computed address, 32-bit displacement, possibly SP-relative
//     C_LECON // computed address, 32-bit displacement, possibly SB-relative, unused?
//     C_DACON // computed address, 64-bit displacment?
//     C_SBRA // short branch
//     C_LBRA // long branch
//     C_SAUTO // short auto
//     C_LAUTO // long auto
//     C_ZOREG // heap address, register-based, displacement == 0
//     C_SOREG // heap address, register-based, int16 displacement
//     C_LOREG // heap address, register-based, int32 displacement
//     C_ANY
//     C_GOK // general address
//     C_ADDR // relocation for extern or static symbols
//     C_TEXTSIZE // text size
//     C_NCLASS
// )
