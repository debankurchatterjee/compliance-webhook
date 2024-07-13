echo "Creating certificates"
mkdir -p certs
openssl genrsa -out certs/tls.key 2048
openssl req -new -key certs/tls.key -out certs/tls.csr -subj "/CN=webhook-server.kube-system.svc"
openssl x509 -req -extfile <(printf "subjectAltName=DNS:webhook-server.kube-system.svc") -in certs/tls.csr -signkey certs/tls.key -out certs/tls.crt

echo "Creating Webhook Server TLS Secret"
kubectl create secret tls webhook-server-tls \
    --cert "certs/tls.crt" \
    --key "certs/tls.key" -n kube-system

make docker-build
echo "Creating Webhook Server Deployment"
kubectl create -f manifests/webhook_server.yaml -n kube-system
echo "Creating K8s Webhooks"
ENCODED_CA=$(cat certs/tls.crt | base64 | tr -d '\n')
sed -e 's@${ENCODED_CA}@'"$ENCODED_CA"'@g' <"manifests/webhooks.yaml" | kubectl create -f -