# cloud-instance-dns
 
<p align="left">
<a href="https://circleci.com/gh/gjbae1212/cloud-instance-dns"><img src="https://circleci.com/gh/gjbae1212/cloud-instance-dns.svg?style=svg"></a>
<a href="https://hits.seeyoufarm.com"/><img src="https://hits.seeyoufarm.com/api/count/incr/badge.svg?url=https%3A%2F%2Fgithub.com%2Fgjbae1212%2Fcloud-instance-dns"/></a>
<a href="/LICENSE"><img src="https://img.shields.io/badge/license-MIT-GREEN.svg" alt="license" /></a>
<a href="https://goreportcard.com/report/github.com/gjbae1212/cloud-instance-dns"><img src="https://goreportcard.com/badge/github.com/gjbae1212/cloud-instance-dns" alt="Go Report Card" /></a> 
</p>

## OVERVIEW
**cloud-instance-dns** is DNS server that will look up public or private ip on AWS ec2 or GCP compute-engine.  
**cloud-instance-dns** is supporting to search multi regions(zones) instances on clouds(AWS,GCP).   
In addition it could be searching private or public ip. 

## Getting Started
### quickstart
```bash
$ bash local.sh build
$ cloud-instance-dns -c your-config.yaml
```

### config(yaml) 

------

#### spec
```yaml
domain: DNS domain # EX) localhost, dns.example.com, hello.example.com ...  
nameserver: public domain for server running on`cloud-instance-dns` # server public domain(never ip) running `cloud-instance-dns`
port: port number
email: your email
prviate: false or true # if you'd like to answer private-ip -> true, but public-ip -> false
aws:
  enable: true or false # if your'd use to aws -> true, but not -> false
  access_key: your-aws-access-key
  secret_access_key: your-aws-secret-access-key
  regions:
    - your-aws-region-1
    - your-aws-region-2
gcp:
  enable: true or false # if your'd use to gcp -> true, but not -> false
  project_id: your-gcp-project-id
  zones:
    - your-gcp-zone-1
    - your-gcp-zone-2
  jwt: your-gcp-jwt-string
```
------

#### config example

- using AWS and GCP
```yaml
domain: hello.example.com  
nameserver: ec2-1.1.1.1.region.compute.amazonaws.com
port: 53
email: gjbae1212@gmail.com
prviate: true
aws:
  enable: true
  access_key: blahblah
  secret_access_key: blahblah
  regions:
    - ap-northeast-1
    - ap-northeast-2    
gcp:
  enable: true
  project_id: gcp-project-id
  zones:    
    - asia-northeast1-a
    - asia-northeast1-b    
  jwt: '{
          blahblah
        }'
```
- only AWS
```
domain: hello.example.com  
nameserver: ec2-1.1.1.1.region.compute.amazonaws.com
port: 53
email: gjbae1212@gmail.com
prviate: true
aws:
  enable: true
  access_key: blahblah
  secret_access_key: blahblah
  regions:
    - ap-northeast-1
    - ap-northeast-2    
gcp:
  enable: false
```
-  only GCP
```
domain: hello.example.com  
nameserver: ec2-1.1.1.1.region.compute.amazonaws.com
port: 53
email: gjbae1212@gmail.com
prviate: true
aws:
  enable: false  
gcp:
  enable: true
  project_id: gcp-project-id
  zones:    
    - asia-northeast1-a
    - asia-northeast1-b    
  jwt: '{
          blahblah
        }'
```
------

### usage
You will search to dns records following rule patterns below, Assume having `hello.example.com` dns

- `(name or instacne-id).hello.example.com` will return instances matching name regardless cloud infra.  
- `(num).(name or instacne-id).hello.example.com` will return a instance matching name and number.
- `(name or instacne-id).aws.hello.example.com` will return instances matching name at aws.
- `(num).(name or instacne-id).aws.hello.example.com` will return a instance matching name and number at aws.
- `(name or instacne-id).gcp.hello.example.com` will return instances matching name at gcp.
- `(num).(name or instacne-id).gcp.hello.example.com` will return a instance matching name and number at gcp.
- `(name or instacne-id).rr.hello.example.com` will return instances matching name with dns round robin.

### install
```bash
# your-machine(mac ... and so on)
$ bash local.sh build

# linux
$ bash local.sh linux_build

# download
go to `https://github.com/gjbae1212/cloud-instance-dns/releases`

# homebrew
$ brew tap gjbae1212/cloud-instance-dns
$ brew install cloud-instance-dns
```

## Explanation
If you would be setup to **cloud-instance-dns**, Be several attention. 
 
### AWS 
- aws.enable of config.yaml should be true when you'd like to use.
- a aws_key must have permission to access ec2(ec2:DescribeInstances).
- ingress port running **cloud-instance-dns** must open(port of config.yaml).

### GCP
- gcp.enable of config.yaml should be true when you'd like to use.
- a gcp-jwt must have permission to access compute-engine(Compute Viewer).
- ingress port running **cloud-instance-dns** must open(port of config.yaml).

### Configure NS Record
If your **cloud-instance-dns** will register global DNS, you must input NS record from your domain.   
Assume having `example.com` domain and you are running **cloud-instance-dns** on instance(assume public domain `ec2-1.1.1.1.region.compute.amazonaws.com`<must not be a IP>).  
And then you will make `hello.example.com.` DNS.
```bash
# DNS Domain(domain in your yaml)                      #TTL          # public hostname(nameserver in your yaml)    
hello.example.com.                                     300   IN  NS  ec2-1.1.1.1.region.compute.amazonaws.com 
``` 
NS record value must not be a IP. It is public domain or hostname<could dns resolve>. 

### Test
- dig (name).hello.example.com @localhost  -->  using localhost dns.
- dig (name).hello.example.com @ec2-1.1.1.1.region.compute.amazonaws.com --> check A record using your public dns. 
- dig NS hello.example.com   --> check NS record using your public dns.
