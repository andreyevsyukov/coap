package coap

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
)

// Proxy Filter
type ProxyFilter func(*Message, *net.UDPAddr) bool

func NullProxyFilter(*Message, *net.UDPAddr) bool {
	return true
}

//type ProxyHandler func(msg *Message, conn *net.UDPConn, addr *net.UDPAddr)
type ProxyHandler func(msg *Message, conn *UDPConnection, addr *net.UDPAddr)

// The default handler when proxying is disabled
//func NullProxyHandler(msg *Message, conn *net.UDPConn, addr *net.UDPAddr) {
func NullProxyHandler(msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	//SendMessageTo(ProxyingNotSupportedMessage(msg.MessageID, MessageAcknowledgment), NewUDPConnection(conn), addr)
	SendMessageTo(ProxyingNotSupportedMessage(msg.MessageID, MessageAcknowledgment), conn, addr)
}

//func COAPProxyHandler(msg *Message, conn *net.UDPConn, addr *net.UDPAddr) {
func COAPProxyHandler(msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	proxyURI := msg.GetOption(OptionProxyURI).StringValue()

	parsedURL, err := url.Parse(proxyURI)
	if err != nil {
		log.Println("Error parsing proxy URI")
		//SendMessageTo(BadGatewayMessage(msg.MessageID, MessageAcknowledgment), NewUDPConnection(conn), addr)
		SendMessageTo(BadGatewayMessage(msg.MessageID, MessageAcknowledgment), conn, addr)
		return
	}

	client := NewCoapClient()
	client.OnStart(func(server CoapServer) {
		client.Dial(parsedURL.Host)

		msg.RemoveOptions(OptionProxyURI)
		req := NewRequestFromMessage(msg)
		req.SetRequestURI(parsedURL.RequestURI())

		response, err := client.Send(req)
		if err != nil {
			//SendMessageTo(BadGatewayMessage(msg.MessageID, MessageAcknowledgment), NewUDPConnection(conn), addr)
			SendMessageTo(BadGatewayMessage(msg.MessageID, MessageAcknowledgment), conn, addr)
			client.Stop()
			return
		}

		//_, err = SendMessageTo(response.GetMessage(), NewUDPConnection(conn), addr)
		_, err = SendMessageTo(response.GetMessage(), conn, addr)
		if err != nil {
			log.Println("Error occured responding to proxy request")
			client.Stop()
			return
		}
		client.Stop()

	})
	client.Start()
}

// Handles requests for proxying from CoAP to HTTP
//func HTTPProxyHandler(msg *Message, conn *net.UDPConn, addr *net.UDPAddr) {
func HTTPProxyHandler(msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	proxyURI := msg.GetOption(OptionProxyURI).StringValue()
	requestMethod := msg.Code

	client := &http.Client{}
	req, err := http.NewRequest(MethodString(CoapCode(msg.GetMethod())), proxyURI, nil)
	if err != nil {
		//SendMessageTo(BadGatewayMessage(msg.MessageID, MessageAcknowledgment), NewUDPConnection(conn), addr)
		SendMessageTo(BadGatewayMessage(msg.MessageID, MessageAcknowledgment), conn, addr)
		return
	}

	etag := msg.GetOption(OptionEtag)
	if etag != nil {
		req.Header.Add("ETag", etag.StringValue())
	}

	// TODO: Set timeout handler, and on timeout return 5.04
	resp, err := client.Do(req)
	if err != nil {
		//SendMessageTo(BadGatewayMessage(msg.MessageID, MessageAcknowledgment), NewUDPConnection(conn), addr)
		SendMessageTo(BadGatewayMessage(msg.MessageID, MessageAcknowledgment), conn, addr)
		return
	}

	defer resp.Body.Close()

	contents, _ := ioutil.ReadAll(resp.Body)
	msg.Payload = NewBytesPayload(contents)
	respMsg := NewRequestFromMessage(msg)

	if requestMethod == Get {
		etag := resp.Header.Get("ETag")
		if etag != "" {
			msg.AddOption(OptionEtag, etag)
		}
	}

	// TODO: Check payload length against Size1 options
	if len(respMsg.GetMessage().Payload.String()) > MaxPacketSize {
		//SendMessageTo(BadGatewayMessage(msg.MessageID, MessageAcknowledgment), NewUDPConnection(conn), addr)
		SendMessageTo(BadGatewayMessage(msg.MessageID, MessageAcknowledgment), conn, addr)
		return
	}

	//_, err = SendMessageTo(respMsg.GetMessage(), NewUDPConnection(conn), addr)
	_, err = SendMessageTo(respMsg.GetMessage(), conn, addr)
	if err != nil {
		println(err.Error())
	}
}

// Handles requests for proxying from HTTP to CoAP
//func HTTPCOAPProxyHandler(msg *Message, conn *net.UDPConn, addr *net.UDPAddr) {
func HTTPCOAPProxyHandler(msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	log.Println("HttpCoapProxyHandler Proxy Handler")
}
