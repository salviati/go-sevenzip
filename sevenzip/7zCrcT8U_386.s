.equ data_size, 28
.equ crc_table, (data_size + 4)

.macro CRC1b
movzbl (%esi),%edx
inc    %esi
movzbl %al,%ebx
xor    %ebx,%edx
shr    $0x8,%eax
xor    0x0(%ebp,%edx,4),%eax
dec    %edi
.endm

.align 16
.global CrcUpdateT8
CrcUpdateT8:
push   %ebx
push   %esi
push   %edi
push   %ebp

mov    0x14(%esp),%eax  # CRC
mov    0x18(%esp),%esi  # buf
mov    data_size(%esp),%edi # size
mov    crc_table(%esp),%ebp # tables

test   %edi,%edi
je     1f

0: # sl
test   $0x7,%esi
je     1f
CRC1b
jne    0b
1: # sl_end

cmp    $0x10,%edi
jb     3f
mov    %edi,data_size(%esp)
sub    $0x8,%edi
and    $0xfffffff8,%edi # ~7
sub    %edi,data_size(%esp)

add    %esi,%edi
xor    (%esi),%eax
mov    0x4(%esi),%ebx
movzbl %bl,%ecx

.align 16
2: # main_loop
mov    0xc00(%ebp,%ecx,4),%edx
movzbl %bh,%ecx
xor    0x800(%ebp,%ecx,4),%edx
shr    $0x10,%ebx
movzbl %bl,%ecx
xor    0x400(%ebp,%ecx,4),%edx
xor    0x8(%esi),%edx
movzbl %al,%ecx
movzbl %bh,%ebx
xor    0x0(%ebp,%ebx,4),%edx

mov    0xc(%esi),%ebx

xor    0x1c00(%ebp,%ecx,4),%edx
movzbl %ah,%ecx
add    $0x8,%esi
shr    $0x10,%eax
xor    0x1800(%ebp,%ecx,4),%edx
movzbl %al,%ecx
xor    0x1400(%ebp,%ecx,4),%edx
movzbl %ah,%ecx
mov    0x1000(%ebp,%ecx,4),%eax
movzbl %bl,%ecx
xor    %edx,%eax

cmp    %edi,%esi
jne    2b
xor    (%esi),%eax

mov    data_size(%esp),%edi

3: #crc_end
test   %edi,%edi
je     5f
4: # fl
CRC1b
jne    4b
5: #fl_end
pop    %ebp
pop    %edi
pop    %esi
pop    %ebx
ret    
