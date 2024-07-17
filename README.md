# Compliance Webhook

A kubernetes validating webhook server which helps to maintain the audit history of mutating operations like create,update and delete
on the cluster with respect to kubernetes objects like deployemnts,replicasets,statefulsets,pods or demonsets.

This way we can track the manual changes made to the cluster using kubectl.