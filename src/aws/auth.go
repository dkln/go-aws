package aws

import (
  "errors"
  "os"
)

type Auth struct {
	AccessKey string
  SecretKey string
  Token string
}

type credentials struct {
	Code            string
	LastUpdated     string
	Type            string
	AccessKeyId     string
	SecretAccessKey string
	Token           string
	Expiration      string
}

/** 
 * GetAuth creates an Auth based on either passed in credentials,
 * environment information or instance based role credentials.
 */
func GetAuth(accessKey string, secretKey string) (Auth, error) {
	// First try passed in credentials
	if accessKey != "" && secretKey != "" {
		return Auth{accessKey, secretKey, ""}, nil
	}

	// Next try to get auth from the environment
  auth, error := EnvAuth()

	if error == nil {
		// Found auth, return
		return auth, nil
	}

	// Next try getting auth from the instance role
	credentials, error := getInstanceCredentials()

	if error == nil {
		// Found auth, return
		auth.AccessKey = credentials.AccessKeyId
		auth.SecretKey = credentials.SecretAccessKey
		auth.Token = credentials.Token

		return auth, nil

	} else {
    return auth, errors.New("No valid AWS authentication found")

  }
}

/** 
 * EnvAuth creates an Auth based on environment information.
 * The AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment
 * variables are used.
 */
func EnvAuth() (auth Auth, err error) {
	auth.AccessKey = os.Getenv("AWS_ACCESS_KEY_ID")

	if auth.AccessKey == "" {
		auth.AccessKey = os.Getenv("AWS_ACCESS_KEY")
	}

	auth.SecretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

	if auth.SecretKey == "" {
		auth.SecretKey = os.Getenv("AWS_SECRET_KEY")
	}

	if auth.AccessKey == "" {
		err = errors.New("AWS_ACCESS_KEY_ID or AWS_ACCESS_KEY not found in environment")
	}

	if auth.SecretKey == "" {
		err = errors.New("AWS_SECRET_ACCESS_KEY or AWS_SECRET_KEY not found in environment")
	}

	return
}
