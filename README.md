# Compliance Webhook

## Preamble
A kubernetes validating webhook server which helps to maintain the audit history of mutating operations like create,update and delete
on the cluster with respect to kubernetes objects like deployemnts,replicasets,statefulsets,pods or demonsets.

This way we can track the manual changes made to the cluster using kubectl.

## High Level Design
![resources/architecture-diagram.png](resources/architecture-diagram.png)

The validation webhook server (compliance-webhook) will only process [CREATE,UPDATE,DELETE] requests for deployemnts,replicasets,statefulsets,pods or demonsets.
it will check if the corresponding service now request is available else it will create a new service now request.

As described in the below sequence diagram.

![resources/webhook-seq-diagram.png](resources/webhook-seq-diagram.png)

### webhook-server flow

![resources/webhook-sever-flow.png](resources/webhook-server-flow.png)

