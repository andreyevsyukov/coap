package coap

import (
	"log"
	"net"
	"fmt"
	"github.com/andreyevsyukov/coap/Logger"
)

//func handleRequest(s CoapServer, err error, msg *Message, conn *net.UDPConn, addr *net.UDPAddr) {
func handleRequest(s CoapServer, err error, msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	if msg.MessageType != MessageReset {
		// Unsupported Method
		if msg.Code != Get && msg.Code != Post && msg.Code != Put && msg.Code != Delete {
			handleReqUnsupportedMethodRequest(s, msg, conn, addr)
			return
		}

		if err != nil {
			s.GetEvents().Error(err)
			if err == ErrUnknownCriticalOption {
				handleReqUnknownCriticalOption(msg, conn, addr)
				return
			}
		}

		// Proxy
		if IsProxyRequest(msg) {
			Logger.Debug("coap.handleRequest", "we got PROXY request")
			handleReqProxyRequest(s, msg, conn, addr)
		} else {
			route, attrs, err := MatchingRoute(msg.GetURIPath(), MethodString(msg.Code), msg.GetOptions(OptionContentFormat), s.GetRoutes())
			if err != nil {
				s.GetEvents().Error(err)
				if err == ErrNoMatchingRoute {
					handleReqNoMatchingRoute(s, msg, conn, addr)
					return
				}

				if err == ErrNoMatchingMethod {
					handleReqNoMatchingMethod(s, msg, conn, addr)
					return
				}

				if err == ErrUnsupportedContentFormat {
					handleReqUnsupportedContentFormat(s, msg, conn, addr)
					return
				}

				log.Println("Error occured parsing inbound message")
				return
			}

			// Duplicate Message ID Check
			if s.IsDuplicateMessage(msg) {
				log.Println("Duplicate Message ID ", msg.MessageID)
				if msg.MessageType == MessageConfirmable {
					handleReqDuplicateMessageID(s, msg, conn, addr)
				}
				return
			}

			s.UpdateMessageTS(msg)

			// Auto acknowledge
			if msg.MessageType == MessageConfirmable && route.AutoAck {
				handleRequestAutoAcknowledge(s, msg, conn, addr)
			}

			req := NewClientRequestFromMessage(msg, attrs, conn, addr)
			if msg.MessageType == MessageConfirmable {

				// Observation Request
				obsOpt := msg.GetOption(OptionObserve)
				if obsOpt != nil {
					handleReqObserve(s, req, msg, conn, addr)
				}
			}

			resp := route.Handler(req)
			_, nilresponse := resp.(NilResponse)
			if !nilresponse {
				respMsg := resp.GetMessage()
				respMsg.Token = req.GetMessage().Token

				// TODO: Validate Message before sending (e.g missing messageId)
				err := ValidateMessage(respMsg)
				if err == nil {
					s.GetEvents().Message(respMsg, false)

					//SendMessageTo(respMsg, NewUDPConnection(conn), addr)
					SendMessageTo(respMsg, conn, addr)
				} else {
					fmt.Println("MESSAGE IS NOT VALID: ", err)
				}
			}
		}
	}
}

func handleReqUnknownCriticalOption(msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	if msg.MessageType == MessageConfirmable {
		SendMessageTo(BadOptionMessage(msg.MessageID, MessageAcknowledgment), conn, addr)
	}
	return
}

func handleReqUnsupportedMethodRequest(s CoapServer, msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	ret := NotImplementedMessage(msg.MessageID, MessageAcknowledgment)
	ret.CloneOptions(msg, OptionURIPath, OptionContentFormat)

	s.GetEvents().Message(ret, false)
	SendMessageTo(ret, conn, addr)
}

func handleReqProxyRequest(s CoapServer, msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	if !s.AllowProxyForwarding(msg, addr) {
		Logger.Debug("PROXY IS NOT ALLOWED")
		SendMessageTo(ForbiddenMessage(msg.MessageID, MessageAcknowledgment), conn, addr)
		/*msg := ContentMessage(msg.MessageID, MessageAcknowledgment)
		msg.SetStringPayload("PROXY IS NOT ALLOWED")
		SendMessageTo(msg, conn, addr)*/
	}

	proxyURI := msg.GetOption(OptionProxyURI).StringValue()

	Logger.Debug("PROXYING THE MSG!", "URI:", proxyURI)

	proxySchemeOption := msg.GetOption(OptionProxyScheme)

	proxyScheme := ""
	if proxySchemeOption != nil {
		proxyScheme = proxySchemeOption.StringValue()
	}

	switch {
		case proxyScheme == "coap" || IsCoapURI(proxyURI):
			s.ForwardCoap(msg, conn, addr)

		case proxyScheme == "http" || IsHTTPURI(proxyURI):
			s.ForwardHTTP(msg, conn, addr)

		default:
			Logger.Error("UNKNOWN PROXY SCHEME!")
			SendMessageTo(BadRequestMessage(msg.MessageID, MessageAcknowledgment), conn, addr)
	}
}

//func handleReqNoMatchingRoute(s CoapServer, msg *Message, conn *net.UDPConn, addr *net.UDPAddr) {
func handleReqNoMatchingRoute(s CoapServer, msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	ret := NotFoundMessage(msg.MessageID, MessageAcknowledgment)
	ret.CloneOptions(msg, OptionURIPath, OptionContentFormat)
	ret.Token = msg.Token

	//SendMessageTo(ret, NewUDPConnection(conn), addr)
	SendMessageTo(ret, conn, addr)
}

//func handleReqNoMatchingMethod(s CoapServer, msg *Message, conn *net.UDPConn, addr *net.UDPAddr) {
func handleReqNoMatchingMethod(s CoapServer, msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	ret := MethodNotAllowedMessage(msg.MessageID, MessageAcknowledgment)
	ret.CloneOptions(msg, OptionURIPath, OptionContentFormat)

	s.GetEvents().Message(ret, false)
	//SendMessageTo(ret, NewUDPConnection(conn), addr)
	SendMessageTo(ret, conn, addr)
}

//func handleReqUnsupportedContentFormat(s CoapServer, msg *Message, conn *net.UDPConn, addr *net.UDPAddr) {
func handleReqUnsupportedContentFormat(s CoapServer, msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	ret := UnsupportedContentFormatMessage(msg.MessageID, MessageAcknowledgment)
	ret.CloneOptions(msg, OptionURIPath, OptionContentFormat)

	s.GetEvents().Message(ret, false)
	//SendMessageTo(ret, NewUDPConnection(conn), addr)
	SendMessageTo(ret, conn, addr)
}

//func handleReqDuplicateMessageID(s CoapServer, msg *Message, conn *net.UDPConn, addr *net.UDPAddr) {
func handleReqDuplicateMessageID(s CoapServer, msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	ret := EmptyMessage(msg.MessageID, MessageReset)
	ret.CloneOptions(msg, OptionURIPath, OptionContentFormat)

	s.GetEvents().Message(ret, false)
	//SendMessageTo(ret, NewUDPConnection(conn), addr)
	SendMessageTo(ret, conn, addr)
}

//func handleRequestAutoAcknowledge(s CoapServer, msg *Message, conn *net.UDPConn, addr *net.UDPAddr) {
func handleRequestAutoAcknowledge(s CoapServer, msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	ack := NewMessageOfType(MessageAcknowledgment, msg.MessageID)

	s.GetEvents().Message(ack, false)
	//SendMessageTo(ack, NewUDPConnection(conn), addr)
	SendMessageTo(ack, conn, addr)
}

//func handleReqObserve(s CoapServer, req CoapRequest, msg *Message, conn *net.UDPConn, addr *net.UDPAddr) {
func handleReqObserve(s CoapServer, req CoapRequest, msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	// TODO: if server doesn't allow observing, return error

	// TODO: Check if observation has been registered, if yes, remove it (observation == cancel)
	resource := msg.GetURIPath()
	if s.HasObservation(resource, addr) {
		// Remove observation of client
		s.RemoveObservation(resource, addr)

		// Observe Cancel Request & Fire OnObserveCancel Event
		s.GetEvents().ObserveCancelled(resource, msg)
	} else {
		// Register observation of client
		s.AddObservation(msg.GetURIPath(), string(msg.Token), addr)

		// Observe Request & Fire OnObserve Event
		s.GetEvents().Observe(resource, msg)
	}

	req.GetMessage().AddOption(OptionObserve, 1)
}
