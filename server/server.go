package server

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"io"
)

type Server interface {

}

type server struct {
	domain string
	aws    *awsstore
	gcp    *gcpstore
}

type awsstore struct {
	client          *ec2.EC2
	regions         []string
}

type gcpstore struct {
}

func NewServer(configMap map[interface{}]interface{}) (Server, error) {


}
