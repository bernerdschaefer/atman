TEXT _rt0_amd64_atman(SB),NOSPLIT,$-8
	CLD
	MOVQ	$runtime·_atman_stack+0x8000(SB), SP

	MOVQ	$runtime·_atman_hello(SB), DX
	MOVQ	$7, SI // strlen
	MOVQ	$0, DI // CONSOLEIO_write
	MOVQ	$runtime·_atman_hypercall_page+0x240(SB), AX
	BYTE $0xFF; BYTE $0xd0 // callq *%rax
loop:
	JMP	loop

DATA runtime·_atman_hello(SB)/8, $"hello\n"
GLOBL runtime·_atman_hello(SB), NOPTR, $8
