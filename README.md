## AuditSink to RBAC

#### Prerequisite
 - `go get -u github.com/searchlight/auditsink-to-rbac`

### Start Minikube in one window
 - `cd /home/$USER/go/src/github.com/searchlight/auditsink-to-rabac`
 - `./minikube-start-dynamic-backend.sh`
 - `kubectl apply -f audit-policy.yaml`
 
### In Second window
 - `cd /home/$USER/go/src/github.com/searchlight/auditsink-to-rabac`
 - `./nats-streaming-server --cluster_id=auditsink-cluster --store file --dir ./data --max_msgs 0 --max_bytes 0`
 

 ### In third window
  - `cd /home/$USER/go/src/github.com/searchlight/auditsink-to-rbac`
  - `go run main.go`
 
 
##### Now your are set. Your every ResponseComplete will be written to `audit.log` file and `Role`s and `RoleBinding`s will be saved to `<username>-roles.yaml`