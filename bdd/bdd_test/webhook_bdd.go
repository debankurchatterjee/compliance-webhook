package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"os"
	"os/exec"
)

var CrNamespace = "snow-compliance"

func querySnowCR(name, operation, kind, namespace, label string) error {
	changeStr := fmt.Sprintf("%s-%s-%s-%s", name, operation, namespace, kind)
	changeID := md5.Sum([]byte(changeStr)) //nolint
	changeIDStr := hex.EncodeToString(changeID[:])
	res, err := runKubectlCommand("get", "snow", "-l", fmt.Sprintf("%s=%s", label, changeIDStr), "-n", CrNamespace)
	if res == "No resources found in snow-compliance namespace." || err != nil {
		return fmt.Errorf("No resources found in %s namespace. %v", CrNamespace, err)
	}
	return nil
}

func runKubectlCommand(args ...string) (string, error) {
	// Construct the command
	cmd := exec.Command("kubectl", args...)
	// Run the command and capture the output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error executing kubectl command: %v, output: %s", err, output)
	}

	return string(output), nil
}

func aValidDeploymentDefinition() error {
	_, err := runKubectlCommand("apply", "-f", "examples/deployments/test_deployment.yml", "--dry-run=client")
	if err != nil {
		return err
	}
	return nil
}

func correspondingCreateSnowCRShouldBeCreatedWithChangeID() error {
	name := "nginx-app"
	operation := "create"
	kind := "Deployment"
	namespace := "test"
	return querySnowCR(name, operation, kind, namespace, "snow.controller/changeID")
}

func iApplyTheDeploymentDefinition() error {
	_, err := runKubectlCommand("apply", "-f", "examples/deployments/test_deployment.yml", "-n", "test")
	if err != nil {
		return err
	}
	return nil
}

func theDeploymentShouldBeCreatedSuccessfully() error {
	_, err := runKubectlCommand("get", "deployment", "nginx-app", "-n", "test")
	if err != nil {
		return err
	}
	return nil
}

func iDeleteTheDeploymentDefinition() error {
	_, err := runKubectlCommand("delete", "deployment", "nginx-app", "-n", "test")
	if err != nil {
		return err
	}
	return nil
}

func iApplyTheUpdateDeploymentDefinition() error {
	_, err := runKubectlCommand("set", "image", "deployment/nginx-app", "nginx=1.26.1", "-n", "test")
	if err != nil {
		return err
	}
	return nil
}

func correspondingUpdateSnowCRShouldBeCreatedWithParentID() error {
	name := "nginx-app"
	operation := "update"
	kind := "Deployment"
	namespace := "test"
	return querySnowCR(name, operation, kind, namespace, "snow.controller/parentChangeID")
}

func theDeploymentShouldBeUpdatedSuccessfully() error {
	return godog.ErrPending
}

func correspondingDeleteSnowShouldBeCreatedWithChangeID() error {
	name := "nginx-app"
	operation := "delete"
	kind := "Deployment"
	namespace := "test"
	return querySnowCR(name, operation, kind, namespace, "snow.controller/changeID")
}

func theDeploymentShouldBeDeleteSuccessfully() error {
	_, err := runKubectlCommand("get", "deployment", "nginx-app", "-n", "test")
	if err != nil {
		return nil
	}
	return fmt.Errorf("deployment was not deleted")
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^a valid Deployment definition$`, aValidDeploymentDefinition)
	ctx.Step(`^I apply the Deployment definition$`, iApplyTheDeploymentDefinition)
	ctx.Step(`^corresponding create snow CR should be created with Change ID$`, correspondingCreateSnowCRShouldBeCreatedWithChangeID)
	ctx.Step(`^the Deployment should be created successfully$`, theDeploymentShouldBeCreatedSuccessfully)
	ctx.Step(`^corresponding update snow CR should be created with parent ID$`, correspondingUpdateSnowCRShouldBeCreatedWithParentID)
	ctx.Step(`^I apply the update Deployment definition$`, iApplyTheUpdateDeploymentDefinition)
	ctx.Step(`^the Deployment should be updated successfully$`, theDeploymentShouldBeUpdatedSuccessfully)
	ctx.Step(`^I delete the Deployment definition$`, iDeleteTheDeploymentDefinition)
	ctx.Step(`^corresponding delete snow CR should be created with Change ID$`, correspondingDeleteSnowShouldBeCreatedWithChangeID)
	ctx.Step(`^the Deployment should be deleted successfully$`, theDeploymentShouldBeDeleteSuccessfully)
}

func main() {
	opts := godog.Options{
		Output: colors.Colored(os.Stdout),
		Paths:  []string{"bdd/bdd_test/compliance-webhook.feature"},
		Format: "pretty",
	}
	status := godog.TestSuite{
		Name:                "godogs",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	os.Exit(status)
}
