# nv2-opa

## install OPA
- find docs for installing opa: https://www.openpolicyagent.org/docs/latest/kubernetes-tutorial/
- this is `NOT` a secure opa setup!


```bash
cd deploy 

# create ns
kubectl create namespace opa


# change context to ns opa


# generate certs
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -days 100000 -out ca.crt -subj "/CN=admission_ca"
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -config server.conf
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 100000 -extensions v3_req -extfile server.conf

# create secret
kubectl create secret tls opa-server --cert=server.crt --key=server.key

# create deployment
kubectl apply -f admission-controller.yaml

# label namespaces to exclude them
kubectl label ns kube-system openpolicyagent.org/webhook=ignore
kubectl label ns opa openpolicyagent.org/webhook=ignore

# create webhook-configuration.yaml with certificate
cat > webhook-configuration.yaml <<EOF
kind: ValidatingWebhookConfiguration
apiVersion: admissionregistration.k8s.io/v1beta1
metadata:
  name: opa-validating-webhook
webhooks:
  - name: validating-webhook.openpolicyagent.org
    namespaceSelector:
      matchExpressions:
      - key: openpolicyagent.org/webhook
        operator: NotIn
        values:
        - ignore
    rules:
      - operations: ["CREATE", "UPDATE"]
        apiGroups: ["*"]
        apiVersions: ["*"]
        resources: ["*"]
    clientConfig:
      caBundle: $(cat ca.crt | base64 | tr -d '\n')
      service:
        namespace: opa
        name: opa
EOF

# create webhook
kubectl apply -f webhook-configuration.yaml

```


## allowed deployment
```bash
# convert wabit-networks certificate to string
cert=$(cat wabbit-networks.crt)
echo "${cert//$'\n'/\\n}"


# add cert string to trusted_certificates to the key `registry.wabbit-networks.io` in file rego/nv2.rego


# create configmap
kubectl create configmap nv2 --from-file=rego/nv2.rego -n opa

# create pod and/or deployment
# for the deployment to succeed, OPA needs to be able to reach the registry
# change the registry in pod.yaml and deployment.yaml to your registrys IP/FQDN
kubectl apply -f app/pod.yaml
kubectl apply -f app/deployment.yaml
```


## denied deployment
```bash

# set comment on correct certificate
sed -e '/"registry": "registry.wabbit-networks.io"/ s/^#*/# /' -i rego/nv2.rego

# delete configmap
kubectl delete cm nv2 -n opa

# create configmap
kubectl create configmap nv2 --from-file=rego/nv2.rego -n opa

# create pod and/or deployment
kubectl apply -f app/pod.yaml
kubectl apply -f app/deployment.yaml
```
