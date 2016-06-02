package coap

import (
	//"errors"
	"net"
	//"time"
	//"fmt"
)

// SendMessageTo sends a CoAP Message to UDP address
func SendMessageTo(msg *Message, conn Connection, addr *net.UDPAddr) (CoapResponse, error) {
	//fmt.Println("SendMessageTo: ", msg.MessageType, msg.MessageID)
	if conn == nil {
		return nil, ErrNilConn
	}

	if msg == nil {
		return nil, ErrNilMessage
	}

	if addr == nil {
		return nil, ErrNilAddr
	}

	//fmt.Println("SendMessageTo::WriteTo")
	b, _ := MessageToBytes(msg)
	_, err := conn.WriteTo(b, addr)

	if err != nil {
		return nil, err
	}

	if msg.MessageType == MessageNonConfirmable {
		//fmt.Println("SendMessageTo::return New Empty Response")
		return NewResponse(NewEmptyMessage(msg.MessageID), err), err
	}

	/*fmt.Println("getResponser()")
	responser := getResponser()
	for {
		fmt.Println("for loop")
		select {
			case respMsg := <-responser.Msg:
				fmt.Println("Got Response Message From Responser: ", respMsg.MessageID, msg.MessageID)
				if respMsg != nil && respMsg.MessageID == msg.MessageID {
					resp := NewResponse(respMsg, nil)
					return resp, nil
				}
			case <-responser.Quit:
				return nil, errors.New("No response recevied :(")
		}
	}*/

	//fmt.Println("SendMessageTo::return")
	return NewResponse(NewEmptyMessage(msg.MessageID), nil), nil

	/*

	// conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	buf, n, err := conn.Read()
	if err != nil {
		return nil, err
	}
	msg, err = BytesToMessage(buf[:n])
	resp := NewResponse(msg, err)

	return resp, err

	*/
}

// We don't really need it...

// SendMessage sends a CoAP Message to a UDP Connection
/*func SendMessage(msg *Message, conn Connection) (CoapResponse, error) {
	if conn == nil {
		return nil, ErrNilConn
	}

	b, _ := MessageToBytes(msg)
	_, err := conn.Write(b)

	if err != nil {
		return nil, err
	}

	if msg.MessageType == MessageNonConfirmable {
		return nil, err
	}

	var buf = make([]byte, 1500)
	conn.SetReadDeadline(time.Now().Add(time.Second * DefaultAckTimeout))
	buf, n, err := conn.Read()

	if err != nil {
		return nil, err
	}

	msg, err = BytesToMessage(buf[:n])

	resp := NewResponse(msg, err)

	return resp, err
}*/
