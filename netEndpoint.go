package qnet

import (
	"errors"
	"fmt"
	"net"
	"regexp"

	"github.com/spf13/cast"
)

// TODO: if merge bus plugin & net plugin, this file can be moved to plugin/netPlugin dir

type ProtoType string

const (
	ProtoTypeUnknown ProtoType = "unknown"
	ProtoTypeTcp     ProtoType = "tcp"
	ProtoTypeUdp     ProtoType = "udp"
	ProtoTypeHttp    ProtoType = "http"
	ProtoTypeHttps   ProtoType = "https"
	ProtoTypeWs      ProtoType = "ws"
	ProtoTypeWss     ProtoType = "wss"
)

func ProtoTypeToStr(t ProtoType) string {
	return cast.ToString(t)
}

func StrToProtoType(t string) ProtoType {
	switch t {
	case "tcp":
		return ProtoTypeTcp
	case "udp":
		return ProtoTypeUdp
	case "http":
		return ProtoTypeHttp
	case "https":
		return ProtoTypeHttps
	case "ws":
		return ProtoTypeWs
	case "wss":
		return ProtoTypeWss
	default:
		return ProtoTypeUnknown
	}
}

type Endpoint struct {
	isIpv6 bool
	proto  ProtoType
	ip     string
	port   uint16
	path   string
}

func NewFromString(url string) (*Endpoint, error) {
	if url == "" {
		return nil, errors.New("AFEndpoint url is empty")
	}

	r, err := regexp.Compile("((.*)://)?([^:/]+)(:(\\d+))?(/.*)?")
	if err != nil {
		return nil, err
	}

	if !r.MatchString(url) {
		return nil, errors.New("unmatched url ` " + url + " `")
	}

	strArr := r.FindStringSubmatch(url)

	port, err := cast.ToUint16E(strArr[5])
	if err != nil {
		return nil, err
	}

	ep := &Endpoint{
		isIpv6: false,
		proto:  StrToProtoType(strArr[2]),
		ip:     strArr[3],
		port:   port,
		path:   strArr[6],
	}
	return ep, nil
}

func (ep *Endpoint) ToString() string {
	var url string
	if ep.proto != ProtoTypeUnknown {
		url += string(ep.proto)
	}

	url += ep.GetIP() + ":" + cast.ToString(ep.GetPort()) + ep.GetPath()

	return url
}

//******* net.Addr ********
func (ep *Endpoint) TCPAddr() (*net.TCPAddr, error) {
	return net.ResolveTCPAddr(string(ep.Proto()), fmt.Sprintf("%s:%d", ep.GetIP(), ep.GetPort()))
}

func (ep *Endpoint) UDPAddr() (*net.UDPAddr, error) {
	return net.ResolveUDPAddr(string(ep.Proto()), fmt.Sprintf("%s:%d", ep.GetIP(), ep.GetPort()))
}

//******* GET & SET ********
func (ep *Endpoint) Proto() ProtoType {
	return ep.proto
}

func (ep *Endpoint) SetProto(proto ProtoType) {
	ep.proto = proto
}

func (ep *Endpoint) GetIP() string {
	return ep.ip
}

func (ep *Endpoint) SetIP(ip string) {
	ep.ip = ip
}

func (ep *Endpoint) GetPath() string {
	return ep.path
}

func (ep *Endpoint) SetPath(path string) {
	ep.path = path
}

func (ep *Endpoint) GetPort() uint16 {
	return ep.port
}

func (ep *Endpoint) SetPort(port uint16) {
	ep.port = port
}

func (ep *Endpoint) IsV6() bool {
	return ep.isIpv6
}

func (ep *Endpoint) SetIsV6(v6 bool) {
	ep.isIpv6 = v6
}
