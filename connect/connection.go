package connect

import (
	"dusnet/logger"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type IConnection interface {
	Read([]byte) (int, error)  // 读取
	Write([]byte) (int, error) // 写入
	Close() error              // 关闭
	Alive() bool               // 是否存活
	SetAlive(alive bool)       // 设置存活
	GetID() uint64             // 获取连接id
	SetID(uint64)              // 设置连接id
	GetLocalHost() string      // 获取本地host
	GetLocalPort() int         // 获取本地端口
	GetRemoteHost() string     // 获取远程host
	GetRemotePort() int        // 获取远程端口
}

type IConnectionMgr interface {
	GetConnByID(id uint64) IConnection          // 根据连接id获取连接
	All() []IConnection                         // 获取所有连接
	GetConnBySrcHost(host string) []IConnection // 根据源host获取连接

	RemoveConnByID(id uint64) error        // 根据连接id移除连接
	RemoveConnBySrcHost(host string) error // 根据源host移除连接

	AddConn(conn IConnection) // 新增连接
	GenConnID() uint64        // 生成连接id
}

type mConnection struct {
	id         uint64
	conn       *net.TCPConn
	alive      bool
	connTime   time.Time
	updateTime time.Time
}

type connectionMgr struct {
	GlobalConnID uint64   // 全局连接id计数
	pool         sync.Map // 连接池
}

func (c *connectionMgr) GenConnID() uint64 {
	c.GlobalConnID++
	return c.GlobalConnID
}

func DefaultConnMgr() IConnectionMgr {
	mgr := &connectionMgr{
		pool:         sync.Map{},
		GlobalConnID: 0,
	}
	return mgr
}

func (m *mConnection) SetAlive(alive bool) {
	m.alive = alive
}

func (m *mConnection) GetRemotePort() int {
	port, _ := strconv.Atoi(strings.Split(m.conn.RemoteAddr().String(), ":")[1])
	return port
}

func (m *mConnection) GetLocalHost() string {
	return strings.Split(m.conn.LocalAddr().String(), ":")[0]
}

func (m *mConnection) GetLocalPort() int {
	port, _ := strconv.Atoi(strings.Split(m.conn.LocalAddr().String(), ":")[1])
	return port
}

func (m *mConnection) GetID() uint64 {
	return m.id
}

func (m *mConnection) Read(bytes []byte) (int, error) {
	i, err := m.conn.Read(bytes)
	return i, err
}

func (m *mConnection) Write(bytes []byte) (int, error) {
	i, err := m.conn.Write(bytes)
	return i, err
}

func (m *mConnection) GetRemoteHost() string {
	// host:port 仅截取host
	return strings.Split(m.conn.RemoteAddr().String(), ":")[0]
}

func New(l *net.TCPListener, mgr IConnectionMgr) IConnection {
	conn, err := l.AcceptTCP()
	if err != nil {
		logger.Error("listener.AcceptTCP error,error:%+v", err)
		return nil
	}
	var idLock sync.RWMutex
	idLock.RLock()
	defer idLock.RUnlock()
	c := &mConnection{
		id:    mgr.GenConnID(),
		conn:  conn,
		alive: true,
	}
	mgr.AddConn(c)
	return c
}

func (m *mConnection) Close() error {
	return m.conn.Close()
}

func (m *mConnection) Alive() bool {
	return m.alive
}

func (m *mConnection) SetID(id uint64) {
	m.id = id
}

func (c *connectionMgr) AddConn(conn IConnection) {
	if conn == nil {
		logger.Warn("connection is nil and will not be added.")
		return
	}
	if conn.Alive() {
		c.pool.Store(conn.GetID(), conn)
		logger.Warn("One connection[id=%d,laddr:%s:%d,raddr:%s:%d] established",
			conn.GetID(), conn.GetLocalHost(), conn.GetLocalPort(), conn.GetRemoteHost(), conn.GetRemotePort())
	} else {
		logger.Warn("connection not alive and will not be added.")
	}
}

func (c *connectionMgr) GetConnByID(id uint64) IConnection {
	if value, ok := c.pool.Load(id); ok {
		conn := (value).(IConnection)
		return conn
	} else {
		return nil
	}
}

func (c *connectionMgr) All() []IConnection {
	var conns []IConnection
	c.pool.Range(func(key, value any) bool {
		conns = append(conns, (value).(IConnection))
		return true
	})
	return conns
}

func (c *connectionMgr) GetConnBySrcHost(host string) []IConnection {
	var conns []IConnection
	if host == "" {
		return conns
	}
	c.pool.Range(func(key, value any) bool {
		conn := (value).(IConnection)
		if conn.GetRemoteHost() == host {
			conns = append(conns, conn)
		}
		return true
	})
	return conns
}

func (c *connectionMgr) RemoveConnByID(id uint64) error {
	value, ok := c.pool.Load(id)
	if !ok {
		logger.Warn("connection with srcHost:%d not exist", id)
		return nil
	}
	conn := (value).(IConnection)
	return doRemoveConn(c, conn)
}

func doRemoveConn(c *connectionMgr, conn IConnection) error {
	if conn != nil && conn.Alive() {
		conn.SetAlive(false)
		if err := conn.Close(); err != nil {
			// close err
			logger.Error("conn.Close error,error:%+v", err)
			return err
		}
	}
	c.pool.Delete(conn.GetID())
	return nil
}

func (c *connectionMgr) RemoveConnBySrcHost(host string) error {
	var conn IConnection

	c.pool.Range(func(key, value any) bool {
		conn0 := (value).(IConnection)
		if conn0.GetRemoteHost() == host {
			conn = conn0
		}
		return true
	})
	if conn == nil {
		logger.Warn("connection with srcHost:%s not exist", host)
		return nil
	}
	return doRemoveConn(c, conn)
}
