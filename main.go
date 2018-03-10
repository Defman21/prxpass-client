package main

import (
    "net"
    "log"
    "fmt"
    "flag"
)

func main() {
    serverAddr := flag.String("server", "localhost:30303", "ngrak instanse address")
    proxyAddr  := flag.String("proxy", "localhost:80", "proxy requests to")
    flag.Parse()
    conn, err := net.Dial("tcp", *serverAddr)
    log.Printf("Connected to %v", *serverAddr)
    log.Printf("Proxying to %v", *proxyAddr)

    if err != nil {
        log.Fatal(err)
    }

    reply := make([]byte, 2048)
    for {

        n, err := conn.Read(reply)

        if err != nil {
            log.Fatal(err)
        } else {
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

