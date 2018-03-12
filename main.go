package main

import (
	"bytes"
	"flag"
	"github.com/vmihailenco/msgpack"
	"log"
	"net"
)

type rpcCall struct {
	Method string
	Args   []string
}

type message struct {
	Sender  string
	Version int
	RPC     rpcCall
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
	flag.Parse()
	conn, err := net.Dial("tcp", *serverAddr)
	log.Printf("Connected to %v", *serverAddr)
	log.Printf("Proxying to %v", *proxyAddr)

	if err != nil {
		log.Fatal(err)
	}

	bytes, err := msgpack.Marshal(&message{
		Sender:  "client",
		Version: 1,
		RPC: rpcCall{
			Method: "net/register",
			Args:   []string{*customID},
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
				case "tcp/request":
					writeConn, err := net.Dial("tcp", *proxyAddr)
					log.Println("Proxy-pass the request to", *proxyAddr)

					if err != nil {
						log.Println("Proxy-pass connection failed:", err)
						continue
					}
					_, err = writeConn.Write([]byte(msgObj.RPC.Args[0]))
					if err != nil {
						log.Println("Proxy-pass write failed:", err)
						continue
					}
					n, err := writeConn.Read(reply)
					if err != nil {
						log.Println("Proxy-pass read failed:", err)
						continue
					}
					httpResponse := reply[:n]
					msgpRequest, err := msgpack.Marshal(&message{
						Sender:  "client",
						Version: 1,
						RPC: rpcCall{
							Method: "tcp/response",
							Args:   []string{string(httpResponse)},
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
