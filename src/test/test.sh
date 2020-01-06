python signature.py --region=cn-east-1 \
       --service=ncs \
       --access-key=c9afd3f2c1af4f03a3134c3ba58189ac \
       --access-secret=eb1997ae8d304516888f80d80129387a \
       -X POST \
       -H 'Content-Type: application/json' \
       --data '{
  "Placement": {
   "ZoneId": ["cn-east-1a"]
  },
  "SpecType": "ncs.n1.medium2",
  "VirtualPrivateCloud": {
   "VpcId": "3c931077-1551-4495-8afd-ff58c9dfd847",
   "SubnetId": "None"
  },
  "SecurityGroupIds": ["f80d92ef-e184-43c3-b72d-96da10d98b29"],
  "ContainerType": "Standard",
  "NamespaceId": 246722,
  "Name": "test",
  "Replicas": 3,
  "MinReadySeconds": 10,
  "Containers": [{
   "Name": "container",
   "Image": "hub.c.163.com/hellwen/startup:latest",
   "Args": ["xxx", "xxx"],
   "Envs": [{
    "Name": "HOSTNAME",
    "Value": "test"
   }],
   "ResourceRequirements": {
    "Limits": {
     "Cpu": 1000,
     "Memory": 1024
    },
    "Requests": {
     "Cpu": 1000,
     "Memory": 1024
    }
   },
  }]
}' \
       "https://open.cn-east-1.163yun.com/ncs?Action=CreateDeployment&Version=2017-11-16"

