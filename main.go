package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"regexp"
)

func main() {
	serverAddr := flag.String("server", "localhost:30303", "ngrak instanse address")
	proxyAddr := flag.String("proxy", "localhost:80", "proxy requests to")
	flag.Parse()
	conn, err := net.Dial("tcp", *serverAddr)
	log.Printf("Connected to %v", *serverAddr)
	log.Printf("Proxying to %v", *proxyAddr)

	if err != nil {
		log.Fatal(err)
	}

	idRegex, _ := regexp.Compile("^~!@=([a-z0-9]+)=@!~$")

	reply := make([]byte, 2048)
	for {

		n, err := conn.Read(reply)

		if err != nil {
			log.Fatal(err)
		} else {
			log.Println(string(reply[:n]))

			if id := idRegex.FindSubmatch(reply[:n]); id != nil {
				log.Printf("Your ID: %v", string(id[1]))
				continue
			}
			writeconn, err := net.Dial("tcp", *proxyAddr)
			log.Printf("Proxy-pass data to %v", *proxyAddr)
			if err != nil {
				fmt.Println(err)
			} else {
				_, err = writeconn.Write(reply[:n])
				if err != nil {
					fmt.Println(err)
				} else {
					n, err := writeconn.Read(reply)
					if err != nil {
						fmt.Println(err)
					} else {
						data := reply[:n]
						log.Printf("Sending data back: %v", string(data))
						_, err := conn.Write(data)
						if err != nil {
							log.Fatal("zulul")
						}
					}
					writeconn.Close()
				}
			}
		}
	}
}
