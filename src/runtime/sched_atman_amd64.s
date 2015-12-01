#include "go_asm.h"
#include "go_tls.h"
#include "textflag.h"

// func taskstart(fn, mp, gp unsafe.Pointer)
TEXT ·taskstart(SB),NOSPLIT,$0
	MOVQ	(SP), R12
	MOVQ	8(SP), R8
	MOVQ	16(SP), R9

	// set m->procid to current task ID
	MOVQ	$runtime·taskcurrent(SB), BX
	MOVQ	(BX), AX
	MOVQ	AX, m_procid(R8)
	
	// Set FS to point at m->tls.
	LEAQ	m_tls(R8), DI
	CALL	runtime·settls(SB)

	// Set up new stack
	get_tls(CX)
	MOVQ	R8, g_m(R9)
	MOVQ	R9, g(CX)
	CALL	runtime·stackcheck(SB)

	// Call fn
	CALL	R12

	// Exit if function returns
	CALL	runtime·taskexit(SB)

	RET // unreachable

// func contextsave(*Context)
TEXT ·contextsave(SB),NOSPLIT,$0-8
	MOVQ	ctx+0(FP), DI
	MOVQ	(SP), CX
	MOVQ	CX, 128(DI)	// save ip to rip
	MOVQ	8(SP), CX
	MOVQ	CX, 152(DI)	// save sp to rsp
	get_tls(CX)
	MOVQ	CX, 184(DI)	// save tls
	RET

// func contextload(*Context)
TEXT ·contextload(SB),NOSPLIT,$0
	MOVQ	ctx+0(FP), DI
	MOVQ	152(DI), R8	// save sp
	MOVQ	128(DI), R9	// save ip
	MOVQ	184(DI), DI
	CALL	runtime·settls(SB) // restore tls
	MOVQ	R8, SP
	MOVQ	R9, (SP) // set return address
	RET
