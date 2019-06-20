package server

import (
	"fmt"
	"hash/crc32"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/logrusorgru/aurora"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type CloudVendor string

type LookupTable map[string][]*Record

const (
	TTL = 300 * time.Second
)

const (
	CacheName             = "CLOUD-NAME-SERVER"
	UNKNOWN   CloudVendor = "UNKNOWN"
	AWS       CloudVendor = "AWS"
	GCP       CloudVendor = "GCP"
)

type Store struct {
	awsconf        *AwsConfig
	gcpconf        *GcpConfig
	cache          *sync.Map
	cacheUpdatedAt time.Time
}

type Record struct {
	Vendor       CloudVendor
	ZoneOrRegion string
	PublicIP     net.IP
	PrivateIP    net.IP
	ExpiredAt    time.Time
}

func (s *Store) Lookup(key string) ([]*Record, error) {
	m, ok := s.cache.Load(CacheName)
	if !ok {
		return nil, fmt.Errorf("[err] unknown error<not found store table>")
	}
	records, ok := m.(LookupTable)[key]
	if !ok {
		return []*Record{}, nil
	}
	return records, nil
}

func (s *Store) renewal() error {
	now := time.Now()

	table := make(LookupTable)
	count := 0
	// aws renewal
	if s.awsconf != nil {
		// get running instances
		input := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{Name: aws.String("instance-state-name"), Values: []*string{aws.String("running")}},
			},
		}

		for region, client := range s.awsconf.clients {
			output, err := client.DescribeInstances(input)
			if err != nil {
				return err
			} else {
				for _, rv := range output.Reservations {
					for _, inst := range rv.Instances {
						record := &Record{Vendor: AWS, ExpiredAt: now.Add(TTL), ZoneOrRegion: region}

						// insert public ip
						if inst.PublicIpAddress != nil {
							if value := net.ParseIP(*inst.PublicIpAddress); value != nil {
								record.PublicIP = value
							}
						}

						// insert private ip
						if inst.PrivateIpAddress != nil {
							if value := net.ParseIP(*inst.PrivateIpAddress); value != nil {
								record.PrivateIP = value
							}
						}

						count += 1
						// register instance-id
						table[strings.ToLower(*inst.InstanceId)] = append(table[strings.ToLower(*inst.InstanceId)], record)

						// register name
						for _, tag := range inst.Tags {
							if *tag.Key == "Name" {
								table[strings.ToLower(*tag.Value)] = append(table[strings.ToLower(*tag.Value)], record)
							}
						}
					}
				}
			}
		}
	}

	// gcp renewal
	if s.gcpconf != nil {
		for _, zone := range s.gcpconf.zones {
			gcpListCall := s.gcpconf.client.Instances.List(s.gcpconf.projectId, zone)
			gcpListCall.Filter("status = RUNNING")
			instances, err := gcpListCall.Do()
			if err != nil {
				return err
			}
			for _, instance := range instances.Items {
				if len(instance.NetworkInterfaces) > 0 {
					record := &Record{Vendor: GCP, ExpiredAt: now.Add(TTL), ZoneOrRegion: zone}
					// insert public ip
					if len(instance.NetworkInterfaces[0].AccessConfigs) > 0 {
						if value := net.ParseIP(instance.NetworkInterfaces[0].AccessConfigs[0].NatIP); value != nil {
							record.PublicIP = value
						}
					}
					// insert private ip
					if value := net.ParseIP(instance.NetworkInterfaces[0].NetworkIP); value != nil {
						record.PrivateIP = value
					}

					count += 1
					// register instance-id
					table[strconv.FormatInt(int64(instance.Id), 10)] = append(table[strconv.FormatInt(int64(instance.Id), 10)], record)

					// register name
					table[strings.ToLower(instance.Name)] = append(table[strings.ToLower(instance.Name)], record)
				}
			}
		}
	}

	// if array is not changed, array of order same as prev order array is returned
	for _, v := range table {
		sort.Slice(v, func(i, j int) bool {
			return crc32.ChecksumIEEE(v[i].PublicIP) < crc32.ChecksumIEEE(v[j].PublicIP)
		})
	}

	s.cache.Store(CacheName, table)
	s.cacheUpdatedAt = time.Now()
	log.Printf("%s[%d] cache table %s\n", aurora.Yellow("[update]"), count, time.Now().String())
	return nil
}

func (r *Record) TTL() time.Duration {
	now := time.Now()
	duration := r.ExpiredAt.Sub(now)
	if duration < 0 {
		return 5 * time.Second
	}
	return duration
}

func NewStore(awsconf *AwsConfig, gcpconf *GcpConfig) (*Store, error) {
	store := &Store{}
	store.cache = &sync.Map{}
	store.awsconf = awsconf
	store.gcpconf = gcpconf

	// a first renewal must be success
	if err := store.renewal(); err != nil {
		return nil, err
	}

	// periodic renewal
	go func() {
		tick := time.NewTicker(1 * time.Minute)
		for {
			select {
			case <-tick.C:
				if err := store.renewal(); err != nil {
					log.Printf("[err] renewal %+v\n", err)
				}
			}
		}
	}()
	return store, nil
}
