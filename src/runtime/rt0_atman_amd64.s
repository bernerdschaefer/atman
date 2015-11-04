#define _PAGE_ROUND_UP(REGISTER)		\
	ADDQ	$0x0000000000000fff, REGISTER	\
	ANDQ	$0xfffffffffffff000, REGISTER

#define _HYPERCALL(OFFSET)				\
	MOVQ	$runtime·_atman_hypercall_page(SB), BX	\
	_PAGE_ROUND_UP(BX)				\
	ADDQ	OFFSET, BX				\
	BYTE $0xff; BYTE $0xd3 // callq *%rbx

#define _HYPERVISOR_console_io(OP, SIZE, DATA_PTR) \
	MOVQ	OP, DI		\
	MOVQ	SIZE, SI	\
	MOVQ	DATA_PTR, DX	\
	_HYPERCALL($0x240)

TEXT _rt0_amd64_atman(SB),NOSPLIT,$-8
	CLD
	MOVQ	$runtime·_atman_stack+0x8000(SB), SP

	_HYPERVISOR_console_io($0, $7, $runtime·_atman_hello(SB))

loop:
	JMP	loop

DATA runtime·_atman_hello(SB)/8, $"hello\n"
GLOBL runtime·_atman_hello(SB), NOPTR, $8
