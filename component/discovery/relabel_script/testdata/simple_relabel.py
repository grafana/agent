def relabel_targets(targets):
	for t in targets:
		namespace, pod = t["job"].split("/")
		t["namespace"] = namespace
		t["pod"] = pod
	return targets
