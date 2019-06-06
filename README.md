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
kubectl create secret docker-registry csrx --docker-server=hub.juniper.net/security --docker-username=$YOUR_USERNAME --docker-password=$YOUR_PASSWORD
```
### Create cSRX Operator
```
kubectl apply -f https://raw.githubusercontent.com/michaelhenkel/csrx-operator/master/deploy/create-csrx-operator.yaml
```
### create a custom resource
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
### Apply custom resource
```
kubectl apply -f common_v1alpha1_csrx_cr.yaml
```
