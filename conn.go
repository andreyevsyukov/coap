package coap

import (
	//"errors"
	"net"
	"time"
)

// ----------------------------------------------------------------

// Creates a new default CanopousConnect
func NewUDPConnection(c *net.UDPConn) *UDPConnection {
	return &UDPConnection{
		conn: c,
	}
}

// Creates a new Connect given an existing UDP
// connection and address
func NewUDPConnectionWithAddr(c *net.UDPConn, a net.Addr) *UDPConnection {
	return &UDPConnection{
		conn: c,
		addr: a,
	}
}

type UDPConnection struct {
	conn *net.UDPConn
	addr net.Addr
}

func (c *UDPConnection) GetConnection() net.Conn {
	return c.conn
}

func (c *UDPConnection) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *UDPConnection) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *UDPConnection) Read() (result []byte, n int, clientAddr *net.UDPAddr, err error) {
	buffer := make([]byte, MaxPacketSize)
	n, clientAddr, err = c.conn.ReadFromUDP(buffer)

	result = make([]byte, n)
	copy(result, buffer)

	return
}

func (c *UDPConnection) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	return  c.conn.WriteToUDP(b, addr.(*net.UDPAddr))
}

func (c *UDPConnection) Close() error {
	return c.conn.Close()
}