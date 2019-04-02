## AuditSink Prototype

#### What you have to do

 - `go get -u github.com/masudur-rahman/auditsink-prototype`
 - `cd /home/$USER/go/src/github.com/masudur-rahman/auditsink-prototype`
 - `go run main.go`
 
#### In another window

 - `./minikube-start-dynamic-backend.sh`
 - `kubectl apply -f audit-policy.yaml`
 
##### Now your are set. Your every ResponseComplete will be written to `audit.log` file.