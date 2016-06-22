package coap

import (
	"net"
)

// SendMessageTo sends a CoAP Message to UDP address
func SendMessageTo(msg *Message, conn *UDPConnection, addr *net.UDPAddr) (CoapResponse, error) {
	if conn == nil {
		return nil, ErrNilConn
	}

	if msg == nil {
		return nil, ErrNilMessage
	}

	if addr == nil {
		return nil, ErrNilAddr
	}

	b, _ := MessageToBytes(msg)
	_, err := conn.WriteTo(b, addr)

	if err != nil {
		return nil, err
	}

	if msg.MessageType == MessageNonConfirmable {
		return NewResponse(NewEmptyMessage(msg.MessageID), err), err
	}

	return NewResponse(NewEmptyMessage(msg.MessageID), nil), nil
}