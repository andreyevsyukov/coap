package coap

import (
	"net"
	"strconv"
	"strings"
	"net/url"
)

// Creates a New Request Instance
func NewRequest(messageType uint8, messageMethod CoapCode, messageID uint16) CoapRequest {
	msg := NewMessage(messageType, messageMethod, messageID)
	msg.Token = []byte(GenerateToken(8))

	return &DefaultCoapRequest{
		msg: msg,
	}
}

func NewConfirmableGetRequest() CoapRequest {
	msg := NewMessage(MessageConfirmable, Get, GenerateMessageID())
	msg.Token = []byte(GenerateToken(8))

	return &DefaultCoapRequest{
		msg: msg,
	}
}

func NewConfirmablePostRequest() CoapRequest {
	msg := NewMessage(MessageConfirmable, Post, GenerateMessageID())
	msg.Token = []byte(GenerateToken(8))

	return &DefaultCoapRequest{
		msg: msg,
	}
}

func NewConfirmablePutRequest() CoapRequest {
	msg := NewMessage(MessageConfirmable, Put, GenerateMessageID())
	msg.Token = []byte(GenerateToken(8))

	return &DefaultCoapRequest{
		msg: msg,
	}
}

func NewConfirmableDeleteRequest() CoapRequest {
	msg := NewMessage(MessageConfirmable, Delete, GenerateMessageID())
	msg.Token = []byte(GenerateToken(8))

	return &DefaultCoapRequest{
		msg: msg,
	}
}

// Creates a new request messages from a CoAP Message
func NewRequestFromMessage(msg *Message) CoapRequest {
	return &DefaultCoapRequest{
		msg: msg,
	}
}

//func NewClientRequestFromMessage(msg *Message, attrs map[string]string, conn *net.UDPConn, addr *net.UDPAddr) CoapRequest {
func NewClientRequestFromMessage(msg *Message, attrs map[string]string, conn *UDPConnection, addr *net.UDPAddr) CoapRequest {
	return &DefaultCoapRequest{
		msg:   msg,
		attrs: attrs,
		conn:  conn,
		addr:  addr,
	}
}

type CoapRequest interface {
	SetProxyURI(uri string)
	SetMediaType(mt MediaType)
	//GetConnection() *net.UDPConn
	GetConnection() *UDPConnection
	GetAddress() *net.UDPAddr
	GetAttributes() map[string]string
	GetAttribute(o string) string
	GetAttributeAsInt(o string) int
	GetMessage() *Message
	SetStringPayload(s string)
	SetRequestURI(uri string)
	SetConfirmable(con bool)
	SetToken(t string)
	GetURIQuery(q string) string
	SetURIQuery(k string, v string)
	GetRequestParamAsString(key string) string
	GetRequestParamAsInteger(key string) int
}

// Wraps a CoAP Message as a Request
// Provides various methods which proxies the Message object methods
type DefaultCoapRequest struct {
	msg    *Message
	attrs  map[string]string
	//conn   *net.UDPConn
	conn   *UDPConnection
	addr   *net.UDPAddr
	server *CoapServer
	parsedRequestBody	url.Values
}

func (c *DefaultCoapRequest) SetProxyURI(uri string) {
	c.msg.AddOption(OptionProxyURI, uri)
}

func (c *DefaultCoapRequest) SetMediaType(mt MediaType) {
	c.msg.AddOption(OptionContentFormat, mt)
}

//func (c *DefaultCoapRequest) GetConnection() *net.UDPConn {
func (c *DefaultCoapRequest) GetConnection() *UDPConnection {
	return c.conn
}

func (c *DefaultCoapRequest) GetAddress() *net.UDPAddr {
	return c.addr
}

func (c *DefaultCoapRequest) GetAttributes() map[string]string {
	return c.attrs
}

func (c *DefaultCoapRequest) GetAttribute(o string) string {
	return c.attrs[o]
}

func (c *DefaultCoapRequest) GetAttributeAsInt(o string) int {
	attr := c.GetAttribute(o)
	i, _ := strconv.Atoi(attr)

	return i
}

func (c *DefaultCoapRequest) GetMessage() *Message {
	return c.msg
}

func (c *DefaultCoapRequest) SetStringPayload(s string) {
	c.msg.Payload = NewPlainTextPayload(s)
}

func (c *DefaultCoapRequest) SetRequestURI(uri string) {
	c.msg.AddOptions(NewPathOptions(uri))
}

func (c *DefaultCoapRequest) SetConfirmable(con bool) {
	if con {
		c.msg.MessageType = MessageConfirmable
	} else {
		c.msg.MessageType = MessageNonConfirmable
	}
}

func (c *DefaultCoapRequest) SetToken(t string) {
	c.msg.Token = []byte(t)
}

func (c *DefaultCoapRequest) GetURIQuery(q string) string {
	qs := c.GetMessage().GetOptionsAsString(OptionURIQuery)

	for _, o := range qs {
		ps := strings.Split(o, "=")
		if len(ps) == 2 {
			if ps[0] == q {
				return ps[1]
			}
		}
	}
	return ""
}

func (c *DefaultCoapRequest) SetURIQuery(k string, v string) {
	c.GetMessage().AddOption(OptionURIQuery, k+"="+v)
}

func (c *DefaultCoapRequest) GetRequestParamAsString(key string) string {
	parseRequestBodyAsURLEncodedParams(c)

	return c.parsedRequestBody.Get(key)
}

func (c *DefaultCoapRequest) GetRequestParamAsInteger(key string) int {
	parseRequestBodyAsURLEncodedParams(c)

	if property, err := strconv.Atoi(c.parsedRequestBody.Get(key)); err == nil {
		return property
	} else {
		return 0
	}
}

func parseRequestBodyAsURLEncodedParams(request *DefaultCoapRequest) {
	if request.parsedRequestBody == nil {
		request.parsedRequestBody = make(url.Values)

		payload := request.GetMessage().Payload
		if payload != nil {
			/*query, err := url.QueryUnescape(payload.String())
			if err != nil {
				return
			}*/
			query := payload.String()

			query = strings.Trim(query, "\r\n ")
			queryMap, err := url.ParseQuery(query)

			if err == nil {
				request.parsedRequestBody = queryMap
			}
		}
	}
}