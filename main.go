package main

import (
	"fmt"
	"math"
	"net"
	"os"
	"strings"
)

// Convert IP address to binary string representation
func ipToBinaryString(ip net.IP) string {
	binaryString := ""
	for _, octet := range ip.To4() {
		binaryString += fmt.Sprintf("%08b.", octet)
	}
	return strings.TrimRight(binaryString, ".")
}

// Calculate the network, broadcast, and range of host IP addresses
func calculateNetworkInfo(ip net.IP, mask net.IPMask) (net.IP, net.IP, net.IP, net.IP) {
	network := ip.Mask(mask)
	broadcast := make(net.IP, len(network))
	copy(broadcast, network)
	for i := range broadcast {
		broadcast[i] |= ^mask[i]
	}

	hostMin := make(net.IP, len(network))
	copy(hostMin, network)
	hostMin[len(hostMin)-1]++

	hostMax := make(net.IP, len(broadcast))
	copy(hostMax, broadcast)
	hostMax[len(hostMax)-1]--

	return network, broadcast, hostMin, hostMax
}

// Determine the class of the network
func getClass(ip net.IP) string {
	firstOctet := ip[0]
	var class, privacy string

	switch {
	case firstOctet <= 127:
		class = "Class A"
	case firstOctet >= 128 && firstOctet <= 191:
		class = "Class B"
	case firstOctet >= 192 && firstOctet <= 223:
		class = "Class C"
	case firstOctet >= 224 && firstOctet <= 239:
		class = "Class D (Multicast)"
	default:
		class = "Class E (Reserved)"
	}

	if isPrivate(ip) {
		privacy = "Private Internet"
	} else {
		privacy = "Public Internet"
	}

	return fmt.Sprintf("%s, %s", class, privacy)
}

// Determine if the network is private
func isPrivate(ip net.IP) bool {

	privateRanges := []struct {
		network *net.IPNet
	}{
		{parseCIDR("10.0.0.0/8")},
		{parseCIDR("172.16.0.0/12")},
		{parseCIDR("192.168.0.0/16")},
	}
	for _, r := range privateRanges {
		if r.network.Contains(ip) {
			return true
		}
	}
	return false
}

func parseCIDR(cidr string) *net.IPNet {
	_, network, _ := net.ParseCIDR(cidr)
	return network
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ipcalc <IP>/<mask>")
		return
	}

	ip, ipNet, err := net.ParseCIDR(os.Args[1])
	if err != nil {
		fmt.Println("Invalid CIDR notation")
		return
	}

	networkIp := ipNet.IP
	mask := ipNet.Mask
	network, broadcast, hostMin, hostMax := calculateNetworkInfo(networkIp, mask)

	netmaskFmt := fmt.Sprintf("%s = %d", net.IP(mask), maskSize(mask))
	networkFmt := fmt.Sprintf("%s /%d", network, maskSize(mask))

	fmt.Printf("Address:   %-20s %s\n", ip, ipToBinaryString(networkIp))
	fmt.Printf("Netmask:   %-20s %s\n", netmaskFmt, ipToBinaryString(net.IP(mask)))
	fmt.Printf("Wildcard:  %-20s %s\n", wildcard(mask), ipToBinaryString(wildcard(mask)))
	fmt.Println("=>")
	fmt.Printf("Network:   %-20s %s\n", networkFmt, ipToBinaryString(network))
	fmt.Printf("HostMin:   %-20s %s\n", hostMin, ipToBinaryString(hostMin))
	fmt.Printf("HostMax:   %-20s %s\n", hostMax, ipToBinaryString(hostMax))
	fmt.Printf("Broadcast: %-20s %s\n", broadcast, ipToBinaryString(broadcast))
	fmt.Printf("Hosts/Net: %-20d %s\n", hostsPerNetwork(mask), getClass(networkIp))
}

// Helper functions

func maskSize(mask net.IPMask) int {
	ones, _ := mask.Size()
	return ones
}

func wildcard(mask net.IPMask) net.IP {
	wildcard := make(net.IP, len(mask))
	for i := range mask {
		wildcard[i] = ^mask[i]
	}
	return wildcard
}

func hostsPerNetwork(mask net.IPMask) int {
	ones, bits := mask.Size()
	return int(math.Pow(2, float64(bits-ones)) - 2)
}
