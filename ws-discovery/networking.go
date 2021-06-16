package wsdiscovery

/*******************************************************
 * Copyright (C) 2018 Palanjyan Zhorzhik
 *
 * This file is part of ws-discovery project.
 *
 * ws-discovery can be copied and/or distributed without the express
 * permission of Palanjyan Zhorzhik
 *******************************************************/

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"golang.org/x/net/ipv4"
)

const bufSize = 8192

//SendProbe to device
func SendProbe(interfaceName string, scopes, types []string, namespaces map[string]string) []string {
	// Creating UUID Version 4
	uuidV4 := uuid.Must(uuid.NewV4())
	//fmt.Printf("UUIDv4: %s\n", uuidV4)

	probeSOAP := buildProbeMessage(uuidV4.String(), scopes, types, namespaces)
	//probeSOAP = `<?xml version="1.0" encoding="UTF-8"?>
	//<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing">
	//<Header>
	//<a:Action mustUnderstand="1">http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</a:Action>
	//<a:MessageID>uuid:78a2ed98-bc1f-4b08-9668-094fcba81e35</a:MessageID><a:ReplyTo>
	//<a:Address>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
	//</a:ReplyTo><a:To mustUnderstand="1">urn:schemas-xmlsoap-org:ws:2005:04:discovery</a:To>
	//</Header>
	//<Body><Probe xmlns="http://schemas.xmlsoap.org/ws/2005/04/discovery">
	//<d:Types xmlns:d="http://schemas.xmlsoap.org/ws/2005/04/discovery" xmlns:dp0="http://www.onvif.org/ver10/network/wsdl">dp0:NetworkVideoTransmitter</d:Types>
	//</Probe>
	//</Body>
	//</Envelope>`
	return sendUDPMulticast(probeSOAP.String(), interfaceName)

}

func _NetGetIfaceIP(ip string) (iface *net.Interface) {
	interfaces, err := net.Interfaces() //all ifaces
	if err != nil {
		return nil
	}
	for _, interf := range interfaces {
		if addrs, err := interf.Addrs(); err == nil {
			for _, addr := range addrs {
				if strings.Contains(addr.String(), ip) {
					return &interf
				}
			}
		}
	}
	return nil
}

func _getDefaultIp() (string, error) {
	conn, err := net.DialTimeout("udp", "1.1.1.1:80", time.Second)
	if err != nil {
		log.Println(err)
		return "", err
	} else {
		defer conn.Close()
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}

func _getDefaultIface() (iface *net.Interface) {
	if dip, err := _getDefaultIp(); err == nil {
		return _NetGetIfaceIP(dip)
	}
	return nil
}

func sendUDPMulticast(msg string, interfaceName string) []string {
	var result []string
	data := []byte(msg)
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		if iface = _getDefaultIface(); iface == nil {
			fmt.Println("[Error]sendUDPMulticast", err)
		} else {
			fmt.Println("[Error]sendUDPMulticast", err)
		}
	}
	group := net.IPv4(239, 255, 255, 250)
	c, err := net.ListenPacket("udp4", "0.0.0.0:0")

	//	c, err := net.ListenPacket("udp4", "0.0.0.0:1080")
	if err != nil {
		fmt.Println(err)
		return result
	}
	defer c.Close()

	p := ipv4.NewPacketConn(c)
	if err := p.JoinGroup(iface, &net.UDPAddr{IP: group}); err != nil {
		fmt.Println(err)
	}

	dst := &net.UDPAddr{IP: group, Port: 3702}
	for _, ifi := range []*net.Interface{iface} {
		if err := p.SetMulticastInterface(ifi); err != nil {
			fmt.Println(err)
		}
		p.SetMulticastTTL(2)
		if _, err := p.WriteTo(data, nil, dst); err != nil {
			fmt.Println(err)
		}
	}

	if err := p.SetReadDeadline(time.Now().Add(time.Second * 2)); err != nil {
		log.Fatal(err)
	}

	for {
		b := make([]byte, bufSize)
		n, _, _, err := p.ReadFrom(b)
		if err != nil {
			//			fmt.Println("sendUDPMulticast:", err)
			break
		}
		result = append(result, string(b[0:n]))
	}
	return result
}
