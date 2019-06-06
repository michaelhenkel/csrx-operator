# Instructions
## Network preparation
### Create Network Attachments
```
cat << EOF > networkattachment.yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: net1
  namespace: default
  annotations:
    "opencontrail.org/cidr" : "1.0.0.0/24"
    "opencontrail.org/ip_fabric_snat" : "true"
    "opencontrail.org/ip_fabric_forwarding" : "true"
spec:
  config: '{
    “cniVersion”: “0.3.0”,
    "type": "contrail-k8s-cni"
}'
---
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: net2
  namespace: default
  annotations:
    "opencontrail.org/cidr" : "2.0.0.0/24"
    "opencontrail.org/ip_fabric_snat" : "true"
    "opencontrail.org/ip_fabric_forwarding" : "true"
spec:
  config: '{
    “cniVersion”: “0.3.0”,
    "type": "contrail-k8s-cni"
}'
EOF
```
### Apply Network Attachments
```
kubectl apply -f networkattachment.yaml
```
## cSRX
### Create a secret for juniper docker repository
```
kubectl create secret docker-registry csrx \
  --docker-server=hub.juniper.net/security \
  --docker-username=$YOUR_USERNAME \
  --docker-password=$YOUR_PASSWORD
```
### Create cSRX Operator
```
kubectl apply -f \
  https://raw.githubusercontent.com/michaelhenkel/csrx-operator/master/deploy/create-csrx-operator.yaml
```
### create a cSrx resource
```
cat << EOF > csrx_cr.yaml
apiVersion: common.contrail.com/v1alpha1
kind: Csrx
metadata:
  name: csrx-1
spec:
  initImage: docker.io/michaelhenkel/csrx-init:latest
  csrxImage: hub.juniper.net/security/csrx:18.1R1.9
  imagePullSecrets:
  - csrx
  networks:
    - name: net1
    - name: net2
EOF
```
### Apply cSrx resource
```
kubectl apply -f csrx_cr.yaml
```
### some checks
```
kubectl get csrx
NAME     AGE
csrx-1   8s

[root@kvm1 ~]# kubectl get csrx csrx-1 -oyaml
apiVersion: common.contrail.com/v1alpha1
kind: Csrx
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"common.contrail.com/v1alpha1","kind":"Csrx","metadata":{"annotations":{},"name":"csrx-1","namespace":"default"},"spec":{"csrxImage":"hub.juniper.net/security/csrx:18.1R1.9","imagePullSecrets":["csrx"],"initImage":"docker.io/michaelhenkel/csrx-init:latest","networks":[{"name":"net1"},{"name":"net2"}]}}
  creationTimestamp: "2019-06-06T10:57:28Z"
  generation: 1
  name: csrx-1
  namespace: default
  resourceVersion: "4335664"
  selfLink: /apis/common.contrail.com/v1alpha1/namespaces/default/csrxes/csrx-1
  uid: dfcadd20-8849-11e9-ab46-525400d14c09
spec:
  csrxImage: hub.juniper.net/security/csrx:18.1R1.9
  imagePullSecrets:
  - csrx
  initImage: docker.io/michaelhenkel/csrx-init:latest
  networks:
  - name: net1
  - name: net2
status:
  nodes:
  - csrx-1-pod
  prefix: ""

[root@kvm1 ~]# kubectl exec -it csrx-1-pod bash
root@csrx-1-pod> show configuration interfaces | display set
set interfaces ge-0/0/0 unit 0 family inet address 2.0.0.252/32
set interfaces ge-0/0/1 unit 0 family inet address 1.0.0.252/32
```
```

