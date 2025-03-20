package tunnel

import (
	"fmt"
	"math/rand/v2"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync/atomic"

	"github.com/andrebq/mixtape/generics"
)

type (
	DNS2Socket struct {
		domains generics.SyncMap[domainKey, []sockInfo]

		count uint64
	}

	sockInfo struct {
		d           *DNS2Socket
		lst         net.Listener
		fp          string
		virtualAddr net.Addr
		dk          domainKey
	}

	domainKey struct {
		host string
		port uint32
	}
)

var (
	// very simplistic dns name validation,
	// basically we jsut need to know if the given name can be used as a
	// valid unix socket name
	validHostName = regexp.MustCompile(`^[a-z0-9]+[a-z0-9]*(.[a-z0-9]+[a-z0-9]*)*$`)
)

func (d *DNS2Socket) CreateNew(host string, port uint32) (net.Listener, error) {
	if !validHostName.MatchString(host) {
		return nil, fmt.Errorf("invalid host, must be a domain name")
	}
	dk := domainKey{host: host, port: port}
	var err error
	var si sockInfo
	id := atomic.AddUint64(&d.count, 1)
	d.domains.Update(dk, func(v []sockInfo, _ bool) ([]sockInfo, bool) {
		sockname := d.computeSocketName(host, port, id)
		err = os.MkdirAll(filepath.Dir(sockname), 0755)
		if err != nil {
			return v, true
		}
		var sock net.Listener
		sock, err = net.Listen("unix", sockname)
		vaddr, _ := net.ResolveTCPAddr("tcp", net.JoinHostPort(host, strconv.FormatUint(uint64(port), 10)))
		si = sockInfo{lst: sock, fp: sockname, d: d, dk: dk, virtualAddr: vaddr}
		for i, old := range v {
			if old.fp == "" {
				v[i] = si
				return v, true
			}
		}
		v = append(v, si)
		return v, true
	})
	return si, err
}

func (d *DNS2Socket) Dial(host string, port uint32) (net.Conn, error) {
	dk := domainKey{host: host, port: port}
	var selected string
	d.domains.Use(dk, func(v []sockInfo) {
		for i := 0; i < 100; i++ {
			idx := rand.IntN(len(v))
			selected = v[idx].fp
			if selected != "" {
				return
			}
		}
	})
	if selected == "" {
		return nil, fmt.Errorf("unable to find socket for given host/port pair")
	}
	return net.Dial("unix", selected)
}

func (d *DNS2Socket) removeSocket(si sockInfo) error {
	d.domains.Update(si.dk, func(v []sockInfo, present bool) (newval []sockInfo, keep bool) {
		for i, s := range v {
			if s.fp == si.fp {
				v[i] = sockInfo{}
			}
		}
		return v, keep
	})
	return si.lst.Close()
}

func (d *DNS2Socket) computeSocketName(host string, port uint32, id uint64) string {
	return filepath.Join("var", "run", host, strconv.FormatUint(uint64(port), 36), fmt.Sprintf("%v.sock", id))
}

func (s sockInfo) Accept() (net.Conn, error) { return s.lst.Accept() }
func (s sockInfo) Addr() net.Addr            { return s.virtualAddr }
func (si sockInfo) Close() error             { return si.d.removeSocket(si) }
