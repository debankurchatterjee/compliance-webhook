Feature: Kubectl operations
       As a Kubernetes user
       I want to run kubectl commands
       So that I can manage my Kubernetes cluster
       All the kubectl mutating operation should be recorded as snow CR

  Scenario: Create a Deployment
        Given a valid Deployment definition
        When I apply the Deployment definition
        Then the Deployment should be created successfully
        Then corresponding create snow CR should be created with Change ID

  Scenario: Update a Deployment
           Given a valid Deployment definition
           When I apply the update Deployment definition
           Then the Deployment should be updated successfully
           Then corresponding update snow CR should be created with parent ID

  Scenario: Delete a Deployment
         Given a valid Deployment definition
         When I delete the Deployment definition
         Then the Deployment should be deleted successfully
         Then corresponding delete snow CR should be created with Change ID
