package asm

// implements api for AWS Secrets Manager

import (
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
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

func ParseKV(secretName string, v interface{}) error {
	ss, err := GetSecretString(secretName)
	if err != nil || ss == "" {
		return err
	}
	return json.Unmarshal([]byte(ss), v)
}

func GetSecretString(secretName string) (string, error) {
	if !asmEnabled {
		return "", nil
	}

	svc, err := newSecretsManager()
	if err != nil {
		return "", err
	}

	result, err := svc.GetSecretValue(createSecretValueInput(secretName))
	if err != nil {
		return "", err
	}
	if result.SecretString == nil {
		return "", nil
	}
	return *result.SecretString, nil
}

func GetSecretStrings(secretNames ...string) (map[string]string, error) {
	if !asmEnabled || len(secretNames) == 0 {
		return nil, nil
	}

	svc, err := newSecretsManager()
	if err != nil {
		return nil, err
	}

	out := map[string]string{}
	for _, v := range secretNames {
		result, err := svc.GetSecretValue(createSecretValueInput(v))
		if err != nil {
			return nil, err
		}
		if result.SecretString == nil {
			out[v] = ""
			continue
		}
		out[v] = *result.SecretString
	}

	return out, nil
}

func newSecretsManager() (*secretsmanager.SecretsManager, error) {
	// go-aws-sdk procita sve iz enva osim regije
	// opcija je da u env za svaki servis stavim AWS_REGION=eu-central-1
	// radije zasad hardkodiram regiju
	region := "eu-central-1"
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	return secretsmanager.New(sess,
		aws.NewConfig().WithRegion(region)), nil
}

func createSecretValueInput(secretName string) *secretsmanager.GetSecretValueInput {
	return &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}
}
