TEXT _rt0_amd64_atman(SB),NOSPLIT,$-8
	CLD
	MOVQ	$runtime·_atman_stack+0x10000(SB), SP
	ANDQ	$(~(0x1000-1)), SP
	LEAQ	8(SP), SI // argv
	MOVQ	0(SP), DI // argc
	MOVQ	$main(SB), AX
	JMP	AX

TEXT main(SB),NOSPLIT,$-8
	MOVQ	$runtime·rt0_go(SB), AX
	JMP	AX

DATA runtime·isatman(SB)/4, $1
GLOBL runtime·isatman(SB), NOPTR, $4
