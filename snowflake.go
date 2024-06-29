package snowflake

import (
	"errors"
	"net"
	"sync"
	"time"
)

const (
	Epoch           = 1288834974657
	BitLenTime      = 41
	BitLenWorkerId  = 10
	BitLenSequence  = 12
	BitMaskWorkerId = (1 << BitLenWorkerId) - 1
	BitMaskSequence = (1 << BitLenSequence) - 1
)

var (
	ErrWorkerInvalid = errors.New("worker id invalid")
	ErrNoPrivateIP   = errors.New("no private ip address")
)

func GenerateWorkerId() (int16, error) {
	as, err := net.InterfaceAddrs()
	if err != nil {
		return 0, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			low := int16(ip[2])<<8 + int16(ip[3])
			return low & BitMaskWorkerId, nil
		}
	}

	return 0, ErrNoPrivateIP
}

type Snowflake struct {
	mutex    *sync.Mutex
	lastMs   int64
	seq      int16
	workerId int16
}

func NewSnowflake(workerId int16) (*Snowflake, error) {
	if workerIdOk := isWorkerIdOk(workerId); workerIdOk == false {
		return nil, ErrWorkerInvalid
	}

	return &Snowflake{
		mutex:    new(sync.Mutex),
		seq:      BitMaskSequence,
		workerId: workerId,
	}, nil
}

func (s *Snowflake) NextId() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.seq += 1
	s.seq %= BitMaskSequence

	ts := timestamp()
	for s.seq == 0 && ts == s.lastMs {
		ts = timestamp()
	}
	s.lastMs = ts

	return (s.lastMs << (BitLenWorkerId + BitLenSequence)) + int64(s.workerId<<BitLenSequence) + int64(s.seq)
}

type Id struct {
	Id        int64 `json:"id"`
	Timestamp int64 `json:"ts"`
	WorkerId  int16 `json:"worker"`
	Sequence  int16 `json:"seq"`
}

func ParseId(id int64) Id {
	return Id{
		Id:        id,
		Timestamp: id >> (BitLenWorkerId + BitLenSequence),
		WorkerId:  int16((id >> BitLenSequence) & BitMaskWorkerId),
		Sequence:  int16(id & BitMaskSequence),
	}
}

func timestamp() int64 {
	return time.Now().UTC().UnixMilli() - Epoch
}

func isWorkerIdOk(workerId int16) bool {
	return workerId > 0 && workerId <= 1023
}

func isPrivateIPv4(ip net.IP) bool {
	// Allow private IP addresses (RFC1918) and link-local addresses (RFC3927)
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168 || ip[0] == 169 && ip[1] == 254)
}
