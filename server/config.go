package server

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	compute "google.golang.org/api/compute/v1"
)

const (
	defaultPort       = "53"
	defaultRName      = "gjbae1212.gmail.com."
	defaultNameServer = "localhost."
)

type CommonConfig struct {
	domain     string
	port       string
	rname      string
	nameserver string
	private    bool
}

type AwsConfig struct {
	clients map[string]*ec2.EC2 // map[region]client
}

type GcpConfig struct {
	projectId string
	zones     []string
	client    *compute.Service
}

//domain, nameserver, port, rname string, private bool,
func ParseConfig(config map[interface{}]interface{}) (commonConfig *CommonConfig, awsConfig *AwsConfig, gcpConfig *GcpConfig, err error) {
	if config == nil {
		err = fmt.Errorf("[err] ParseConfig empty params")
		return
	}

	commonConfig = &CommonConfig{}

	// get domain
	if v, ok := config["domain"]; !ok {
		commonConfig = nil
		err = fmt.Errorf("[err] ParseConfig empty domain")
		return
	} else {
		rawDomain := strings.TrimSpace(v.(string))
		if !strings.HasSuffix(rawDomain, ".") {
			commonConfig.domain = rawDomain + "."
		} else {
			commonConfig.domain = rawDomain
		}
	}

	// get nameserver
	if v, ok := config["nameserver"]; !ok {
		commonConfig.nameserver = defaultNameServer
	} else {
		rawNameserver := strings.TrimSpace(v.(string))
		if !strings.HasSuffix(rawNameserver, ".") {
			commonConfig.nameserver = rawNameserver + "."
		} else {
			commonConfig.nameserver = rawNameserver
		}
	}

	// get port
	if v, ok := config["port"]; !ok {
		commonConfig.port = defaultPort
	} else {
		switch v.(type) {
		case int, int64:
			commonConfig.port = fmt.Sprintf("%d", v)
		case string:
			commonConfig.port = strings.TrimSpace(v.(string))
		}
	}

	// get email
	if v, ok := config["email"]; !ok {
		commonConfig.rname = defaultRName
	} else {
		commonConfig.rname = strings.Replace(v.(string), "@", ".", -1) + "."
	}

	// private
	if v, ok := config["private"]; !ok {
		commonConfig.private = false
	} else {
		switch v.(type) {
		case bool:
			commonConfig.private = v.(bool)
		case string:
			e, suberr := strconv.ParseBool(strings.TrimSpace(v.(string)))
			if suberr != nil {
				commonConfig = nil
				err = fmt.Errorf("[err] private field is invalid.")
				return
			}
			commonConfig.private = e
		}
	}

	for name, v := range config {
		switch name.(string) {
		case "aws":
			data, ok := v.(map[interface{}]interface{})["enable"]
			if ok {
				enable := false
				switch data.(type) {
				case bool:
					enable = data.(bool)
				case string:
					e, suberr := strconv.ParseBool(strings.TrimSpace(data.(string)))
					if suberr != nil {
						err = suberr
						return
					}
					enable = e
				}
				if enable {
					ak, ok := v.(map[interface{}]interface{})["access_key"]
					if !ok || strings.TrimSpace(ak.(string)) == "" {
						continue
					}
					sak, ok := v.(map[interface{}]interface{})["secret_access_key"]
					if !ok || strings.TrimSpace(sak.(string)) == "" {
						continue
					}
					regions, ok := v.(map[interface{}]interface{})["regions"]
					if !ok {
						continue
					}

					accessKey := strings.TrimSpace(ak.(string))
					secretAccessKey := strings.TrimSpace(sak.(string))
					awsConfig = &AwsConfig{
						clients: make(map[string]*ec2.EC2),
					}
					for _, r := range regions.([]interface{}) {
						region := strings.TrimSpace(r.(string))
						sess, suberr := session.NewSession(&aws.Config{
							Region:      aws.String(region),
							Credentials: credentials.NewStaticCredentials(accessKey, secretAccessKey, ""),
						})
						if suberr != nil {
							err = suberr
							return
						}
						awsservice := ec2.New(sess)
						awsConfig.clients[region] = awsservice
					}
					// if valid client is not exist.
					if len(awsConfig.clients) == 0 {
						gcpConfig = nil
					}
				}
			}
		case "gcp":
			data, ok := v.(map[interface{}]interface{})["enable"]
			if ok {
				enable := false
				switch data.(type) {
				case bool:
					enable = data.(bool)
				case string:
					e, suberr := strconv.ParseBool(strings.TrimSpace(data.(string)))
					if suberr != nil {
						err = suberr
						return
					}
					enable = e
				}
				if enable {
					projectId, ok := v.(map[interface{}]interface{})["project_id"]
					if !ok || strings.TrimSpace(projectId.(string)) == "" {
						continue
					}
					jwt, ok := v.(map[interface{}]interface{})["jwt"]
					if !ok {
						continue
					}
					zones, ok := v.(map[interface{}]interface{})["zones"]
					if !ok {
						continue
					}

					gcpConfig = &GcpConfig{
						projectId: strings.TrimSpace(projectId.(string)),
					}
					for _, zone := range zones.([]interface{}) {
						gcpConfig.zones = append(gcpConfig.zones, strings.TrimSpace(zone.(string)))
					}
					jwtConfig, suberr := google.JWTConfigFromJSON([]byte(jwt.(string)), compute.ComputeScope)
					if suberr != nil {
						err = suberr
						return
					}
					gcpservice, suberr := compute.NewService(context.Background(), option.WithTokenSource(jwtConfig.TokenSource(context.Background())))
					if suberr != nil {
						err = suberr
						return
					}
					gcpConfig.client = gcpservice
					// if zones is not exist.
					if len(gcpConfig.zones) == 0 {
						gcpConfig = nil
					}
				}
			}
		}
	}
	return
}