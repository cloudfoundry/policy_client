package policy_client_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPolicyClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PolicyClient Suite")
}
