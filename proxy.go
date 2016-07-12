package coap

import (
	"github.com/andreyevsyukov/coap/Logger"
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

type ProxyHandler func(s CoapServer, msg *Message, conn *UDPConnection, addr *net.UDPAddr)

// The default handler when proxying is disabled
func NullProxyHandler(s CoapServer, msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	SendMessageTo(ProxyingNotSupportedMessage(msg.MessageID, MessageAcknowledgment), conn, addr)
}

func COAPProxyHandler(s CoapServer, msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	proxyURI := msg.GetOption(OptionProxyURI).StringValue()

	parsedURL, err := url.Parse(proxyURI)
	if err != nil {
		log.Println("Error parsing proxy URI")
		SendMessageTo(BadGatewayMessage(msg.MessageID, MessageAcknowledgment), conn, addr)
		return
	}

	/*client := NewServer("0", parsedURL.Host)
	client.OnStart(func(server CoapServer) {

		msg.RemoveOptions(OptionProxyURI)
		req := NewRequestFromMessage(msg)
		req.SetRequestURI(parsedURL.RequestURI())

		err := client.SendAndWaitForCallback(req, func(respMsg *Message) {
			Logger.Debug("Proxy RESP", CoapCodeToString(respMsg.Code), respMsg.String())

			_, err = SendMessageTo(respMsg, conn, addr)
			if err != nil {
				Logger.Error("Error occured responding to proxy request")
			}

			client.Stop()
		})

		if err != nil {
			Logger.Debug("Proxy ERROR", err)
			SendMessageTo(BadGatewayMessage(msg.MessageID, MessageAcknowledgment), conn, addr)
			client.Stop()
		}
	})

	client.Start()*/

	s.Dial(parsedURL.Host)

	msg.RemoveOptions(OptionProxyURI)
	req := NewRequestFromMessage(msg)
	req.SetRequestURI(parsedURL.RequestURI())

	err = s.SendAndWaitForCallback(req, func(respMsg *Message) {
		Logger.Debug("Proxy RESP", CoapCodeToString(respMsg.Code), respMsg.String())

		_, err = SendMessageTo(respMsg, conn, addr)
		if err != nil {
			Logger.Error("Error occured responding to proxy request")
		}
	})

	if err != nil {
		Logger.Debug("Proxy ERROR", err)
		SendMessageTo(BadGatewayMessage(msg.MessageID, MessageAcknowledgment), conn, addr)
	}
}

// Handles requests for proxying from CoAP to HTTP
func HTTPProxyHandler(s CoapServer, msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	proxyURI := msg.GetOption(OptionProxyURI).StringValue()
	requestMethod := msg.Code

	client := &http.Client{}
	req, err := http.NewRequest(MethodString(CoapCode(msg.GetMethod())), proxyURI, nil)
	if err != nil {
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
func HTTPCOAPProxyHandler(s CoapServer, msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	log.Println("HttpCoapProxyHandler Proxy Handler")
}
