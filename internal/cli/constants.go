package cli

const (
	cronjobs               = "cronjobs"
	daemonsets             = "daemonsets"
	deployments            = "deployments"
	jobs                   = "jobs"
	pods                   = "pods"
	replicasets            = "replicasets"
	replicationcontrollers = "replicationcontrollers"
	statefulsets           = "statefulsets"
)

var all_resources = []string{
	cronjobs,
	daemonsets,
	deployments,
	jobs,
	pods,
	replicasets,
	replicationcontrollers,
	statefulsets,
}
