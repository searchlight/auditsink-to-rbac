package system

var (
	SystemRequest string
	UserList      []string
)

func init() {
	UserList = append(UserList,
		"system:apiserver",
		"system:kube-scheduler",
		"system:kube-controller-manager",
		"system:serviceaccount:kube-system:resourcequota-controller",
		"system:serviceaccount:kube-system:generic-garbage-collector",
		"system:serviceaccount:kube-system:cronjob-controller",
		"system:serviceaccount:kube-system:default",
		"system:serviceaccount:kube-system:coredns",
		"system:serviceaccount:kube-system:deployment-controller",
		"system:serviceaccount:kube-system:endpoint-controller",
		"system:serviceaccount:kube-system:kube-proxy",
		"system:serviceaccount:kube-system:pod-garbage-collector",
		"system:serviceaccount:kube-system:replicaset-controller",
		"system:serviceaccount:kube-system:storage-provisioner",
		"system:serviceaccount:kube-system:attachdetach-controller",

		"system:node:minikube",
		"minikube",
	)
	SystemRequest = "/healthz"
}
