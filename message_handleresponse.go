package coap

import "net"

//func handleResponse(s CoapServer, msg *Message, conn *net.UDPConn, addr *net.UDPAddr) {
func handleResponse(s CoapServer, msg *Message, conn *UDPConnection, addr *net.UDPAddr) {
	if msg.GetOption(OptionObserve) != nil {
		handleAcknowledgeObserveRequest(s, msg)
		return
	}

	//responser := GetResponser()
	//responser.Msg <- msg
	RunAwaitResponseHandler(msg)
}
func handleAcknowledgeObserveRequest(s CoapServer, msg *Message) {
	s.GetEvents().Notify(msg.GetURIPath(), msg.Payload, msg)
}
