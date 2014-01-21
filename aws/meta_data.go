package aws

import (
  "fmt"
  "io/ioutil"
  "encoding/json"
)

/**
 * GetMetaData retrieves instance metadata about the current machine.
 * See http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AESDG-chapter-instancedata.html for more details.
 */
func GetMetaData(path string) ([]byte, error) {
	url := "http://169.254.169.254/latest/meta-data/" + path

	response, error := RetryingClient.Get(url)

	if error != nil {
		return nil, error
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Code %d returned for url %s", response.StatusCode, url)
	}

	body, error := ioutil.ReadAll(response.Body)

	if error != nil {
		return nil, error
	}

	return []byte(body), nil
}

/**
 *
 */
func getInstanceCredentials() (cred credentials, err error) {
	credentialPath := "iam/security-credentials/"

	// Get the instance role
	role, err := GetMetaData(credentialPath)
	if err != nil {
		return
	}

	// Get the instance role credentials
	credentialJSON, err := GetMetaData(credentialPath + string(role))
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(credentialJSON), &cred)
	return
}
