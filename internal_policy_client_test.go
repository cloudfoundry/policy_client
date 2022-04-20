package policy_client_test

import (
	"encoding/json"
	"errors"

	hfakes "code.cloudfoundry.org/cf-networking-helpers/fakes"
	"code.cloudfoundry.org/policy_client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var policyData = `
{
	"policies": [
		{
			"source": { "id": "some-app-guid", "tag": "BEEF" },
			"destination": { "id": "some-other-app-guid", "protocol": "tcp", "port": 8090, "ports": { "start": 8090, "end": 8090 } } 
		}
	],
	"egress_policies": [
		{
			"source": { "id": "some-other-app-guid" },
			"destination": { "protocol": "tcp", "ips": [{ "start": "1.2.3.4", "end": "1.2.3.5" }] }
		},
		{
			"source": { "id": "some-other-app-guid" },
			"destination": { "protocol": "icmp", "icmp_type": 8, "icmp_code": 4, "ips": [{ "start": "1.2.3.4", "end": "1.2.3.5" }] }
		}
	]
}`

var asgData1 = `
{
  "next": 2,
  "security_groups": [
    {
	  "guid": "public-asg-guid",
      "name": "public_networks",
      "rules": "[{\"protocol\":\"all\",\"destination\":\"0.0.0.0-9.255.255.255\",\"ports\":\"\",\"type\":0,\"code\":0,\"description\":\"\",\"log\":false},{\"protocol\":\"all\",\"destination\":\"11.0.0.0-169.253.255.255\",\"ports\":\"\",\"type\":0,\"code\":0,\"description\":\"\",\"log\":false}]",
      "staging_default": true,
      "running_default": true,
      "staging_space_guids": [],
      "running_space_guids": []
    },
    {
      "guid": "sg-1-guid",
      "name": "security-group-1",
      "rules": "[{\"protocol\":\"icmp\",\"destination\":\"0.0.0.0/0\",\"ports\":\"\",\"type\":0,\"code\":0,\"description\":\"\",\"log\":false}]",
      "staging_default": false,
      "running_default": false,
      "staging_space_guids": [],
      "running_space_guids": [
        "some-space-guid"
      ]
    }
  ]
}
`

var asgData2 = `
{
  "next": 0,
  "security_groups": [
    {
      "guid": "sg-2-guid",
      "name": "security-group-2",
      "rules": "[{\"protocol\":\"tcp\",\"destination\":\"10.0.11.0/24\",\"ports\":\"80,443\",\"type\":0,\"code\":0,\"description\":\"Allow http and https traffic to ZoneA\",\"log\":true}]",
      "staging_default": false,
      "running_default": false,
      "staging_space_guids": [
        "some-other-space-guid"
      ],
      "running_space_guids": []
    }
  ]
}
`

var _ = Describe("InternalClient", func() {
	var (
		client     *policy_client.InternalClient
		jsonClient *hfakes.JSONClient
	)

	BeforeEach(func() {
		jsonClient = &hfakes.JSONClient{}
		client = &policy_client.InternalClient{
			JsonClient: jsonClient,
			Config:     policy_client.Config{PerPageSecurityGroups: 2},
		}
	})

	Describe("GetPolicies", func() {
		BeforeEach(func() {
			jsonClient.DoStub = func(method, route string, reqData, respData interface{}, token string) error {
				respBytes := []byte(policyData)
				json.Unmarshal(respBytes, respData)
				return nil
			}
		})

		It("does the right json http client request", func() {
			policies, err := client.GetPolicies()
			Expect(err).NotTo(HaveOccurred())

			Expect(jsonClient.DoCallCount()).To(Equal(1))
			method, route, reqData, _, token := jsonClient.DoArgsForCall(0)
			Expect(method).To(Equal("GET"))
			Expect(route).To(Equal("/networking/v1/internal/policies"))
			Expect(reqData).To(BeNil())

			Expect(policies).To(Equal([]*policy_client.Policy{
				{
					Source: policy_client.Source{
						ID:  "some-app-guid",
						Tag: "BEEF",
					},
					Destination: policy_client.Destination{
						ID: "some-other-app-guid",
						Ports: policy_client.Ports{
							Start: 8090,
							End:   8090,
						},
						Protocol: "tcp",
					},
				},
			},
			))
			Expect(token).To(BeEmpty())
		})

		Context("when the json client fails", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(errors.New("banana"))
			})
			It("returns the error", func() {
				_, err := client.GetPolicies()
				Expect(err).To(MatchError("banana"))
			})
		})
	})

	Describe("GetPoliciesByID", func() {
		BeforeEach(func() {
			jsonClient.DoStub = func(method, route string, reqData, respData interface{}, token string) error {
				respBytes := []byte(policyData)
				json.Unmarshal(respBytes, respData)
				return nil
			}
		})

		It("does the right json http client request", func() {
			policies, err := client.GetPoliciesByID("some-app-guid", "some-other-app-guid")
			Expect(err).NotTo(HaveOccurred())

			Expect(jsonClient.DoCallCount()).To(Equal(1))
			method, route, reqData, _, token := jsonClient.DoArgsForCall(0)
			Expect(method).To(Equal("GET"))
			Expect(route).To(Equal("/networking/v1/internal/policies?id=some-app-guid,some-other-app-guid"))
			Expect(reqData).To(BeNil())

			Expect(policies).To(Equal([]policy_client.Policy{
				{
					Source: policy_client.Source{
						ID:  "some-app-guid",
						Tag: "BEEF",
					},
					Destination: policy_client.Destination{
						ID: "some-other-app-guid",
						Ports: policy_client.Ports{
							Start: 8090,
							End:   8090,
						},
						Protocol: "tcp",
					},
				},
			},
			))

			Expect(token).To(BeEmpty())
		})

		Context("when the json client fails", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(errors.New("banana"))
			})
			It("returns the error", func() {
				_, err := client.GetPoliciesByID("foo")
				Expect(err).To(MatchError("banana"))
			})
		})

		Context("when ids is empty", func() {
			BeforeEach(func() {})
			It("returns an error and does not call the json http client", func() {
				policies, err := client.GetPoliciesByID()
				Expect(err).To(MatchError("ids cannot be empty"))
				Expect(policies).To(BeNil())
				Expect(jsonClient.DoCallCount()).To(Equal(0))
			})
		})
	})

	Describe("GetSecurityGroupsForSpace", func() {
		BeforeEach(func() {
			callCount := 0
			jsonClient.DoStub = func(method, route string, reqData, respData interface{}, token string) error {
				var respBytes []byte
				if callCount == 0 {
					respBytes = []byte(asgData1)
				} else {
					respBytes = []byte(asgData2)
				}
				err := json.Unmarshal(respBytes, respData)
				Expect(err).NotTo(HaveOccurred())
				callCount++
				return nil
			}
		})

		It("does the right json http client request", func() {
			securityGroups, err := client.GetSecurityGroupsForSpace("some-space-guid", "some-other-space-guid")
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonClient.DoCallCount()).To(Equal(2))
			method, route, reqData, _, token := jsonClient.DoArgsForCall(0)
			Expect(method).To(Equal("GET"))
			Expect(route).To(Equal("/networking/v1/internal/security_groups?per_page=2&space_guids=some-space-guid,some-other-space-guid"))
			Expect(reqData).To(BeNil())
			Expect(token).To(BeEmpty())

			method, route, reqData, _, token = jsonClient.DoArgsForCall(1)
			Expect(method).To(Equal("GET"))
			Expect(route).To(Equal("/networking/v1/internal/security_groups?per_page=2&space_guids=some-space-guid,some-other-space-guid&from=2"))
			Expect(reqData).To(BeNil())
			Expect(token).To(BeEmpty())

			Expect(securityGroups).To(ContainElements(policy_client.SecurityGroup{
				Guid: "public-asg-guid",
				Name: "public_networks",
				Rules: []policy_client.SecurityGroupRule{
					{
						Protocol:    "all",
						Destination: "0.0.0.0-9.255.255.255",
					},
					{
						Protocol:    "all",
						Destination: "11.0.0.0-169.253.255.255",
					},
				},
				StagingDefault:    true,
				RunningDefault:    true,
				StagingSpaceGuids: []string{},
				RunningSpaceGuids: []string{},
			}, policy_client.SecurityGroup{
				Guid: "sg-1-guid",
				Name: "security-group-1",
				Rules: []policy_client.SecurityGroupRule{
					{
						Protocol:    "icmp",
						Destination: "0.0.0.0/0",
					},
				},
				StagingDefault:    false,
				RunningDefault:    false,
				StagingSpaceGuids: []string{},
				RunningSpaceGuids: []string{"some-space-guid"},
			}, policy_client.SecurityGroup{
				Guid: "sg-2-guid",
				Name: "security-group-2",
				Rules: []policy_client.SecurityGroupRule{
					{
						Protocol:    "tcp",
						Destination: "10.0.11.0/24",
						Ports:       "80,443",
						Description: "Allow http and https traffic to ZoneA",
						Log:         true,
					},
				},
				StagingDefault:    false,
				RunningDefault:    false,
				StagingSpaceGuids: []string{"some-other-space-guid"},
				RunningSpaceGuids: []string{},
			}))
		})

		Context("when the json client fails", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(errors.New("banana"))
			})
			It("returns the error", func() {
				_, err := client.GetPoliciesByID("foo")
				Expect(err).To(MatchError("banana"))
			})
		})

		Context("when space_guids is empty", func() {
			var globalAsgData = `
			{
			  "next": 0,
			  "security_groups": [
				{
				  "guid": "public-asg-guid",
				  "name": "public_networks",
				  "rules": "[{\"protocol\":\"all\",\"destination\":\"0.0.0.0-9.255.255.255\",\"ports\":\"\",\"type\":0,\"code\":0,\"description\":\"\",\"log\":false},{\"protocol\":\"all\",\"destination\":\"11.0.0.0-169.253.255.255\",\"ports\":\"\",\"type\":0,\"code\":0,\"description\":\"\",\"log\":false}]",
				  "staging_default": true,
				  "running_default": true,
				  "staging_space_guids": [],
				  "running_space_guids": []
				}
			  ]
			}
			`
			BeforeEach(func() {
				jsonClient.DoStub = func(method, route string, reqData, respData interface{}, token string) error {
					respBytes := []byte(globalAsgData)
					err := json.Unmarshal(respBytes, respData)
					Expect(err).NotTo(HaveOccurred())
					return nil
				}
			})

			It("returns global security groups", func() {
				//		policies, egressPolicies, err := client.GetPoliciesByID()
				securityGroups, err := client.GetSecurityGroupsForSpace()
				Expect(err).NotTo(HaveOccurred())
				Expect(jsonClient.DoCallCount()).To(Equal(1))
				method, route, reqData, _, token := jsonClient.DoArgsForCall(0)
				Expect(method).To(Equal("GET"))
				Expect(route).To(Equal("/networking/v1/internal/security_groups?per_page=2"))
				Expect(reqData).To(BeNil())
				Expect(token).To(BeEmpty())
				Expect(securityGroups).To(ContainElements(policy_client.SecurityGroup{
					Guid: "public-asg-guid",
					Name: "public_networks",
					Rules: []policy_client.SecurityGroupRule{
						{
							Protocol:    "all",
							Destination: "0.0.0.0-9.255.255.255",
						},
						{
							Protocol:    "all",
							Destination: "11.0.0.0-169.253.255.255",
						},
					},
					StagingDefault:    true,
					RunningDefault:    true,
					StagingSpaceGuids: []string{},
					RunningSpaceGuids: []string{},
				}))
			})
		})
	})

	Describe("CreateOrGetTag", func() {
		BeforeEach(func() {
			jsonClient.DoStub = func(method, route string, reqData, respData interface{}, token string) error {
				respBytes := []byte(`{ "id": "SOME_ID", "type": "some_type", "tag": "1234" }`)
				json.Unmarshal(respBytes, respData)
				return nil
			}
		})
		It("returns a tag", func() {
			tag, err := client.CreateOrGetTag("SOME_ID", "some_type")
			Expect(err).NotTo(HaveOccurred())
			Expect(tag).To(Equal("1234"))

			Expect(jsonClient.DoCallCount()).To(Equal(1))
			method, route, reqData, _, token := jsonClient.DoArgsForCall(0)
			Expect(method).To(Equal("PUT"))
			Expect(route).To(Equal("/networking/v1/internal/tags"))
			Expect(reqData).To(Equal(policy_client.TagRequest{
				ID:   "SOME_ID",
				Type: "some_type",
			}))
			Expect(token).To(BeEmpty())
		})

		Context("when the json client fails", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(errors.New("banana"))
			})
			It("returns the error", func() {
				_, err := client.CreateOrGetTag("", "")
				Expect(err).To(MatchError("banana"))
			})
		})
	})

	Describe("HealthCheck", func() {
		BeforeEach(func() {
			jsonClient.DoStub = func(method, route string, reqData, respData interface{}, token string) error {
				respBytes := []byte(`{ "healthcheck": true }`)
				json.Unmarshal(respBytes, respData)
				return nil
			}
		})

		It("Returns if the server is up", func() {
			health, err := client.HealthCheck()
			Expect(err).NotTo(HaveOccurred())
			Expect(health).To(Equal(true))

			Expect(jsonClient.DoCallCount()).To(Equal(1))
			method, route, reqData, _, token := jsonClient.DoArgsForCall(0)
			Expect(method).To(Equal("GET"))
			Expect(route).To(Equal("/networking/v1/internal/healthcheck"))
			Expect(reqData).To(BeNil())
			Expect(token).To(BeEmpty())
		})

		Context("when the json client fails", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(errors.New("banana"))
			})
			It("returns the error", func() {
				_, err := client.HealthCheck()
				Expect(err).To(MatchError("banana"))
			})
		})
	})
})
