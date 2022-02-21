package asm

// implements api for AWS Secrets Manager

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"os"
)

const (
	EnvAsmEnabled = "SVCKIT_ASM_ENABLED"
)

var (
	asmEnabled = false
)

func init() {
	env, ok := os.LookupEnv(EnvAsmEnabled)
	if !ok || (env == "0") || (env == "false") || (env == "") {
		return
	}
	asmEnabled = true
}

func GetKV(secretName string) (map[string]string, error) {
	if !asmEnabled {
		return nil, nil
	}
	// go-aws-sdk procita sve iz enva osim regije
	// opcija je da u env za svaki servis stavim AWS_REGION=eu-central-1
	// radije zasad hardkodiram regiju
	region := "eu-central-1"
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	svc := secretsmanager.New(sess,
		aws.NewConfig().WithRegion(region))
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}
	result, err := svc.GetSecretValue(input)
	if err != nil {
		return nil, err
	}
	if result.SecretString == nil {
		return nil, nil
	}
	res := map[string]string{}
	err = json.Unmarshal([]byte(*result.SecretString), &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
