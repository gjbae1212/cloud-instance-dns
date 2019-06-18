package server

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	goip "github.com/gjbae1212/go-module/ip"
	"github.com/logrusorgru/aurora"
	"github.com/miekg/dns"
)

const (
	defaultNameServer = "localhost."
)

type Server interface {
	Start()
}

type server struct {
	domain     string
	port       string
	rname      string
	nameserver string
	publicIP   string
	store      *Store
}

func (s *server) Start() {
	udpServer := &dns.Server{Addr: ":" + s.port, Net: "udp"}
	go udpServer.ListenAndServe()
	tcpServer := &dns.Server{Addr: ":" + s.port, Net: "tcp"}
	log.Printf("%s listen(%s) nameserver(%s) domain(%s)\n", aurora.Green("[start]"), aurora.Blue(fmt.Sprintf("%s:%s", s.publicIP, s.port)),
		aurora.Yellow(s.nameserver), aurora.Cyan(s.domain))
	tcpServer.ListenAndServe()
}

func (s *server) Lookup(search string) ([]*Record, error) {
	search = strings.ToLower(search)
	seps := strings.Split(search, ".")
	if len(seps) == 1 {
		return s.store.Lookup(seps[0])
	}

	prefix := seps[0]
	suffix := seps[len(seps)-1]
	num := 0
	checkVendor := UNKNOWN
	if ix, err := strconv.Atoi(prefix); err == nil {
		num = ix
	}
	if suffix == "aws" {
		checkVendor = AWS
	} else if suffix == "gcp" {
		checkVendor = GCP
	}

	var query string
	if checkVendor != UNKNOWN && num > 0 { // if prefix is number and suffix is aws or gcp
		if len(seps) == 2 {
			query = strings.Join(seps[0:(len(seps)-1)], ".")
		} else {
			query = strings.Join(seps[1:(len(seps)-1)], ".")
		}
	} else if checkVendor != UNKNOWN && num == 0 { // if prefix is string and suffix is aws or gcp
		query = strings.Join(seps[0:(len(seps)-1)], ".")
	} else if checkVendor == UNKNOWN && num > 0 { // if prefix is number and suffix is not aws or gcp
		query = strings.Join(seps[1:(len(seps))], ".")
	} else { // etc
		query = strings.Join(seps[0:(len(seps))], ".")
	}

	allRecords, err := s.store.Lookup(query)
	if err != nil {
		return nil, err
	}

	var filter []*Record
	if checkVendor != UNKNOWN {
		for _, record := range allRecords {
			if record.Vendor == checkVendor {
				filter = append(filter, record)
			}
		}
	} else {
		filter = allRecords
	}

	var records []*Record
	if num > 0 {
		if len(filter) >= num {
			records = append(records, filter[num-1])
		}
	} else {
		records = filter
	}
	return records, nil
}

func (s *server) dnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false
	m.Authoritative = true

	for _, msg := range m.Question {
		switch msg.Qtype {
		case dns.TypeNS: // dns nameserver
			if msg.Name == s.domain {
				m.Answer = append(m.Answer, s.ns())
			}
		case dns.TypeSOA: // dns info
			if msg.Name == s.domain {
				m.Answer = append(m.Answer, s.soa())
			}
		case dns.TypeA: // ipv4
			if strings.HasSuffix(msg.Name, s.domain) {
				prefix := strings.TrimSpace(strings.TrimSuffix(msg.Name, "."+s.domain))
				records, err := s.Lookup(prefix)
				if err != nil {
					log.Printf("[err] lookup %+v\n", err)
				} else {
					for _, record := range records {
						m.Answer = append(m.Answer, &dns.A{
							Hdr: dns.RR_Header{
								Name:   msg.Name,
								Rrtype: dns.TypeA,
								Class:  dns.ClassINET,
								Ttl:    uint32(record.TTL() / time.Second),
							},
							A: record.PublicIP,
						})
					}
				}
			}
		}
	}

	// if response is not exist.
	if len(m.Answer) == 0 {
		m.Ns = append(m.Ns, s.soa())
	}

	w.WriteMsg(m)
}

func (s *server) ns() *dns.NS {
	return &dns.NS{
		Hdr: dns.RR_Header{Name: s.domain, Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: uint32(TTL / time.Second)},
		Ns:  s.nameserver,
	}
}

func (s *server) soa() *dns.SOA {
	return &dns.SOA{
		Hdr:     dns.RR_Header{Name: s.domain, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: uint32(TTL / time.Second)},
		Ns:      s.nameserver,
		Mbox:    s.rname,
		Serial:  uint32(s.store.cacheUpdatedAt.Unix()), // cache updatedAt
		Refresh: uint32((6 * time.Hour) / time.Second),
		Retry:   uint32((30 * time.Minute) / time.Second),
		Expire:  uint32((24 * time.Hour) / time.Second),
		Minttl:  uint32((2 * time.Minute) / time.Second),
	}
}

func NewServer(yamlPath string) (Server, error) {
	if yamlPath == "" {
		return nil, fmt.Errorf("[err] empty params")
	}

	config := make(map[interface{}]interface{})
	bys, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(bys, &config); err != nil {
		return nil, err
	}

	domain, port, rname, awsconfig, gcpconfig, err := ParseConfig(config)
	if err != nil {
		return nil, err
	}

	if awsconfig == nil && gcpconfig == nil {
		return nil, fmt.Errorf("[err] the aws or the gcp must be useful at least one")
	}

	publicIP, err := goip.GetPublicIPV4()
	if err != nil {
		log.Printf("%s not found machine public ip \n", aurora.Red("[fail]"))
	}

	// check NS Record
	nameserver := defaultNameServer
	nsrecords, err := net.LookupNS(domain)
	if err != nil {
		log.Printf("%s %s not found NS Record\n", aurora.Red("[fail]"), aurora.Magenta(domain))
	} else {
		for _, ns := range nsrecords {
			ips, err := net.LookupIP(ns.Host)
			if err != nil {
				log.Printf("%s %s not found NS Domain IP\n", aurora.Red("[fail]"), aurora.Magenta(ns.Host))
			} else {
				check := false
				for _, ip := range ips {
					if ip.String() == publicIP {
						check = true
					}
				}
				if check {
					nameserver = ns.Host
					log.Printf("%s %s matched %s \n", aurora.Green("[success]"), aurora.Magenta(domain), aurora.Magenta(publicIP))
				} else {
					log.Printf("%s %s not matched %s \n", aurora.Red("[fail]"), aurora.Magenta(domain), aurora.Magenta(publicIP))
				}
			}
		}
	}

	store, err := NewStore(awsconfig, gcpconfig)
	if err != nil {
		return nil, err
	}
	s := &server{domain: domain, port: port, nameserver: nameserver, rname: rname, publicIP: publicIP, store: store}

	// register handler
	dns.HandleFunc(s.domain, s.dnsRequest)
	return Server(s), nil
}
