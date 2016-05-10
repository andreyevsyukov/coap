package coap

import (
	//"errors"
	"net"
	"time"
)

// ----------------------------------------------------------------

// Creates a new default CanopousConnect
func NewUDPConnection(c *net.UDPConn) Connection {
	return &UDPConnection{
		conn: c,
	}
}

// Creates a new Connect given an existing UDP
// connection and address
func NewUDPConnectionWithAddr(c *net.UDPConn, a net.Addr) Connection {
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

func (c *UDPConnection) Read() (buf []byte, n int, err error) {
	buf = make([]byte, MaxPacketSize)
	n, _, err = c.conn.ReadFromUDP(buf)

	return
}

func (c *UDPConnection) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	n, err = c.conn.WriteToUDP(b, addr.(*net.UDPAddr))

	return
}
