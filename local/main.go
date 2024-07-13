package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/compliance-webhook/pkg/controller"
)

func main() {
	snowController, err := controller.NewSnowResource("compliance.complaince.org", "v1", "snows", false)
	if err != nil {
		return
	}
	ctx := context.Background()
	name := "nginx-app"
	kind := ""
	namespace := "default"
	operation := "create"

	snowOperation := fmt.Sprintf("%s-%s-%s-%s", name,
		operation,
		namespace,
		kind)
	// changeStr := fmt.Sprintf("%s-%s-%s-%s", name, operation, namespace, kind)
	changeID := md5.Sum([]byte(snowOperation))
	changeIDStr := hex.EncodeToString(changeID[:])
	get, err := snowController.Get(ctx, changeIDStr, "", "create")
	if err != nil {
		return
	}
	fmt.Println(get)
}
