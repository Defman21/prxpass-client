package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/vmihailenco/msgpack"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
)

type RPC struct {
	Method string
	Args   []string
}

type message struct {
	Sender  string
	Version int
	RPC
}

func formatMessage(msgpackBytes []byte) []byte {
	return append([]byte("!msgpack:"), msgpackBytes...)
}

func parseMessage(msg []byte) (*message, bool, error) {
	if bytes.HasPrefix(msg, []byte("!msgpack:")) {
		var obj message
		err := msgpack.Unmarshal(msg[9:], &obj)
		if err != nil {
			return nil, false, err
		}
		if obj.Version == 0 {
			return nil, false, nil
		}
		return &obj, true, nil
	}
	return nil, false, nil
}

func main() {
	serverAddr := flag.String("server", "test.loc:30303", "PrxPass server address")
	proxyAddr := flag.String("proxy-to", "localhost:80", "Where to proxy requests")
	customID := flag.String("id", "", "Custom client ID")
	password := flag.String("password", "", "Server password")
	proxyHost := strings.Split(*proxyAddr, ":")[0]
	flag.Parse()
	conn, err := net.Dial("tcp", *serverAddr)
	log.Printf("Connected to %v", *serverAddr)
	log.Printf("Proxying to %v", *proxyAddr)
	hostRegexp := regexp.MustCompile("Host: .+")

	if err != nil {
		log.Fatal(err)
	}

	bytes, err := msgpack.Marshal(&message{
		Sender:  "client",
		Version: 1,
		RPC: RPC{
			Method: "net/register",
			Args:   []string{*customID, *password},
		},
	})

	_, err = conn.Write(formatMessage(bytes))

	reply := make([]byte, 2048)
	for {

		n, err := conn.Read(reply)

		if err != nil {
			log.Fatal(err)
		} else {
			if msgObj, isMsgpack, _ := parseMessage(reply[:n]); isMsgpack {
				switch msgObj.RPC.Method {
				case "net/notify":
					log.Println("Your ID:", msgObj.RPC.Args[0])
				case "net/auth-reject":
					log.Println(msgObj.RPC.Args[0])
				case "http/request":
					request := hostRegexp.ReplaceAllString(msgObj.RPC.Args[0], fmt.Sprintf("Host: %v", proxyHost))
					log.Printf("%+v", request)
					requestObj, _ := http.ReadRequest(bufio.NewReader(strings.NewReader(request)))
					requestObj.URL.Host = *proxyAddr
					requestObj.URL.Scheme = "http"
					log.Printf("%+v", requestObj)
					res, err := http.DefaultTransport.RoundTrip(requestObj)
					if err != nil {
						panic(err)
					}
					httpResponse, err := httputil.DumpResponse(res, true)
					if err != nil {
						panic(err)
					}
					msgpRequest, err := msgpack.Marshal(&message{
						Sender:  "client",
						Version: 1,
						RPC: RPC{
							Method: "http/response",
							Args:   []string{string(httpResponse)},
						},
					})
					if err != nil {
						log.Println("Failed to create a msgpack object:", err)
						continue
					}
					_, err = conn.Write(formatMessage(msgpRequest))
					if err != nil {
						log.Println("Failed to RPC:http/response:", err)
						continue
					}
				case "net/request":
					writeConn, err := net.Dial("tcp", *proxyAddr)
					request := msgObj.RPC.Args[0]
					log.Printf("%+v", request)

					if err != nil {
						log.Println("Proxy-pass connection failed:", err)
						continue
					}
					_, err = writeConn.Write([]byte(request))
					if err != nil {
						log.Println("Proxy-pass write failed:", err)
						continue
					}
					n, err := writeConn.Read(reply)
					if err != nil {
						log.Println("Proxy-pass read failed:", err)
						continue
					}
					tcpResponse := reply[:n]
					msgpRequest, err := msgpack.Marshal(&message{
						Sender:  "client",
						Version: 1,
						RPC: RPC{
							Method: "tcp/response",
							Args:   []string{string(tcpResponse)},
						},
					})
					if err != nil {
						log.Println("Failed to create a msgpack object:", err)
						continue
					}
					_, err = conn.Write(formatMessage(msgpRequest))
					if err != nil {
						log.Println("Failed to RPC:tcp/response:", err)
						continue
					}
					writeConn.Close()
				}
			}
		}
	}
}
