//+build !noasm !appengine
// AUTO-GENERATED BY C2GOASM -- DO NOT EDIT

TEXT ·_crc32u64(SB), $0-24

    MOVQ a+0(FP), DI
    MOVQ b+8(FP), SI
    MOVQ result+16(FP), DX

    LONG $0x380f48f2; WORD $0xfef1 // crc32    rdi, rsi
    WORD $0x8948; BYTE $0x3a     // mov    qword [rdx], rdi
    RET


// func cpuid(eaxArg, ecxArg uint32) (eax, ebx, ecx, edx uint32)
TEXT ·cpuid(SB), $0-24
	MOVL eaxArg+0(FP), AX
	MOVL ecxArg+4(FP), CX
	CPUID
	MOVL AX, eax+8(FP)
	MOVL BX, ebx+12(FP)
	MOVL CX, ecx+16(FP)
	MOVL DX, edx+20(FP)
	RET
