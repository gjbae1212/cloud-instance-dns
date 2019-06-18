package server

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net"

	goip "github.com/gjbae1212/go-module/ip"
	"github.com/logrusorgru/aurora"
	"github.com/miekg/dns"
)

const (
	defaultNameServer = "localhost."
)

type Server interface {
	Lookup(search string) ([]*Record, error)
	Start()
}

type server struct {
	domain     string
	port       string
	nameserver string
	publicIP   string
	store      *Store
}

func (s *server) Start() {
	udpServer := &dns.Server{Addr: ":" + s.port, Net: "udp"}
	go udpServer.ListenAndServe()
	tcpServer := &dns.Server{Addr: ":" + s.port, Net: "tcp"}
	log.Printf("[start] listen(%s) nameserver(%s) publicIP(%s)\n", aurora.Green(fmt.Sprintf("%s:%s", s.domain, s.port)),
		aurora.Yellow(s.nameserver), aurora.Cyan(s.publicIP))
	tcpServer.ListenAndServe()
}

func (s *server) Lookup(search string) ([]*Record, error) {
	return nil, nil
}

func (s *server) dnsRequest(w dns.ResponseWriter, r *dns.Msg) {
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

	domain, port, awsconfig, gcpconfig, err := ParseConfig(config)
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
	s := &server{domain: domain, port: port, nameserver: nameserver, publicIP: publicIP, store: store}

	// register handler
	dns.HandleFunc(s.domain, s.dnsRequest)
	return Server(s), nil
}
