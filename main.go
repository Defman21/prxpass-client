package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vmihailenco/msgpack"
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
	var serverAddr string
	var proxyAddr string
	var customID string
	var password string
	rootCmd := &cobra.Command{
		Use:   "prxpass-client [address]",
		Short: "Expose your localhost to the Internet",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			proxyAddr = args[0]
			proxyHost := strings.Split(proxyAddr, ":")[0]
			conn, err := net.Dial("tcp", serverAddr)
			log.Printf("Connected to %v", serverAddr)
			log.Printf("Proxying to %v", proxyAddr)
			hostRegexp := regexp.MustCompile("Host: .+")

			if err != nil {
				log.Fatal(err)
			}

			bytes, err := msgpack.Marshal(&message{
				Sender:  "client",
				Version: 1,
				RPC: RPC{
					Method: "net/register",
					Args:   []string{customID, password},
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
							log.Println("Your public URL:", msgObj.RPC.Args[1])
						case "net/auth-reject":
							log.Println(msgObj.RPC.Args[0])
						case "http/request":
							request := hostRegexp.ReplaceAllString(msgObj.RPC.Args[0], fmt.Sprintf("Host: %v", proxyHost))
							requestObj, _ := http.ReadRequest(bufio.NewReader(strings.NewReader(request)))
							requestObj.URL.Host = proxyAddr
							requestObj.URL.Scheme = "http"
							log.Printf("%s %s", requestObj.Method, requestObj.RequestURI)
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
							writeConn, err := net.Dial("tcp", proxyAddr)
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
		},
	}
	rootCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "localhost:30303", "PrxPass server address")
	rootCmd.PersistentFlags().StringVar(&customID, "id", "", "Custom client ID")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Server password")
	rootCmd.Execute()
}
