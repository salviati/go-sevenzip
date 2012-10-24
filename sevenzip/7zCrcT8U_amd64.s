.macro CRC1b
movzbl (%rsi),%edx
inc    %rsi
movzbl %al,%ebx
xor    %ebx,%edx
shr    $0x8,%eax
xor    (%rdi,%rdx,4),%eax
dec    %r8
.endm

.align 16
.global CrcUpdateT8
CrcUpdateT8:
push   %rbx
push   %rbp
mov    %edi,%eax /* EAX = CRC */
mov    %rdx,%r8  /* R8 = LEN */
mov    %rcx,%rdi /* RDI = table */
test   %r8,%r8
je     1f

0: /* sl */
test   $0x7,%rsi
je     1f
CRC1b
jne    0b
1: /* sl_end */

cmp    $0x10,%r8
jb     3f
mov    %r8,%r9
and    $0x7,%r8
add    $0x8,%r8
sub    %r8,%r9

add    %rsi,%r9
xor    (%rsi),%eax
mov    0x4(%rsi),%ebx
movzbl %bl,%ecx

.align 16

2: /* main_loop */
mov    0xc00(%rdi,%rcx,4),%edx
movzbl %bh,%ebp
xor    0x800(%rdi,%rbp,4),%edx
shr    $0x10,%ebx
movzbl %bl,%ecx
xor    0x8(%rsi),%edx
xor    0x400(%rdi,%rcx,4),%edx
movzbl %al,%ecx
movzbl %bh,%ebp
xor    (%rdi,%rbp,4),%edx

mov    0xc(%rsi),%ebx

xor    0x1c00(%rdi,%rcx,4),%edx
movzbl %ah,%ebp
shr    $0x10,%eax
movzbl %al,%ecx
xor    0x1800(%rdi,%rbp,4),%edx
movzbl %ah,%ebp
mov    0x1400(%rdi,%rcx,4),%eax
add    $0x8,%rsi
xor    0x1000(%rdi,%rbp,4),%eax
movzbl %bl,%ecx
xor    %edx,%eax

cmp    %r9,%rsi
jne    2b
xor    (%rsi),%eax

3: /* crc_end */

test   %r8,%r8
je     5f
4: /* fl */
CRC1b
jne    4b
5: /* fl_end */
pop    %rbp
pop    %rbx
retq
