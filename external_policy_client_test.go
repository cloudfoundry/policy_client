package policy_client_test

import (
	"encoding/json"
	"errors"
	"net/http"

	hfakes "code.cloudfoundry.org/cf-networking-helpers/fakes"
	"code.cloudfoundry.org/cf-networking-helpers/json_client"
	. "code.cloudfoundry.org/policy_client"
	"code.cloudfoundry.org/policy_client/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ExternalClient", func() {
	var (
		client      *ExternalClient
		fakeChunker *fakes.Chunker
		jsonClient  *hfakes.JSONClient
	)

	BeforeEach(func() {
		jsonClient = &hfakes.JSONClient{}
		fakeChunker = &fakes.Chunker{}
		fakeChunker.ChunkReturns([][]PolicyV0{
			[]PolicyV0{
				{
					Source: SourceV0{
						ID: "some-app-guid",
					},
					Destination: DestinationV0{
						ID:       "some-other-app-guid",
						Port:     8090,
						Protocol: "tcp",
					},
				},
				{
					Source: SourceV0{
						ID: "some-app-guid-2",
					},
					Destination: DestinationV0{
						ID:       "some-other-app-guid-2",
						Port:     8091,
						Protocol: "tcp",
					},
				},
			},
			[]PolicyV0{
				{
					Source: SourceV0{
						ID: "some-app-guid-3",
					},
					Destination: DestinationV0{
						ID:       "some-other-app-guid-3",
						Port:     8092,
						Protocol: "tcp",
					},
				},
			},
		})
		client = &ExternalClient{
			JsonClient: jsonClient,
			Chunker:    fakeChunker,
		}
	})

	Describe("GetPolicies", func() {
		BeforeEach(func() {
			jsonClient.DoStub = func(method, route string, reqData, respData interface{}, token string) error {
				respBytes := []byte(`{ "policies": [ {"source": { "id": "some-app-guid" }, "destination": { "id": "some-other-app-guid", "protocol": "tcp", "ports": { "start": 8090, "end": 8100 } } } ] }`)
				json.Unmarshal(respBytes, respData)
				return nil
			}
		})
		It("does the right json http client request", func() {
			policies, err := client.GetPolicies("some-token")
			Expect(err).NotTo(HaveOccurred())

			Expect(jsonClient.DoCallCount()).To(Equal(1))
			method, route, reqData, _, token := jsonClient.DoArgsForCall(0)
			Expect(method).To(Equal("GET"))
			Expect(route).To(Equal("/networking/v1/external/policies"))
			Expect(reqData).To(BeNil())

			Expect(policies).To(Equal([]Policy{
				{
					Source: Source{
						ID: "some-app-guid",
					},
					Destination: Destination{
						ID: "some-other-app-guid",
						Ports: Ports{
							Start: 8090,
							End:   8100,
						},
						Protocol: "tcp",
					},
				},
			},
			))
			Expect(token).To(Equal("some-token"))
		})
		Context("when the json client fails", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(errors.New("banana"))
			})
			It("returns the error", func() {
				_, err := client.GetPolicies("some-token")
				Expect(err).To(MatchError("banana"))
			})
		})
		Context("when the json client gets a bad status code", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(&json_client.HttpResponseCodeError{
					StatusCode: http.StatusTeapot,
					Message:    "some-error",
				})
			})
			It("parses out the error body", func() {
				_, err := client.GetPolicies("some-token")
				Expect(err).To(MatchError("418 I'm a teapot: some-error"))
			})
		})
	})

	Describe("GetPoliciesByID", func() {
		BeforeEach(func() {
			jsonClient.DoStub = func(method, route string, reqData, respData interface{}, token string) error {
				respBytes := []byte(`{ "policies": [ {"source": { "id": "some-app-guid" }, "destination": { "id": "some-other-app-guid", "protocol": "tcp", "ports": { "start": 8090, "end": 8100 } } } ] }`)
				json.Unmarshal(respBytes, respData)
				return nil
			}
		})
		It("does the right json http client request", func() {
			policies, err := client.GetPoliciesByID("some-token", "some-app-guid", "another-app-guid")
			Expect(err).NotTo(HaveOccurred())

			Expect(jsonClient.DoCallCount()).To(Equal(1))
			method, route, reqData, _, token := jsonClient.DoArgsForCall(0)
			Expect(method).To(Equal("GET"))
			Expect(route).To(Equal("/networking/v1/external/policies?id=some-app-guid,another-app-guid"))
			Expect(reqData).To(BeNil())
			Expect(policies).To(Equal([]Policy{
				{
					Source: Source{
						ID: "some-app-guid",
					},
					Destination: Destination{
						ID: "some-other-app-guid",
						Ports: Ports{
							Start: 8090,
							End:   8100,
						},
						Protocol: "tcp",
					},
				},
			},
			))
			Expect(token).To(Equal("some-token"))
		})
		Context("when the json client fails", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(errors.New("banana"))
			})
			It("returns the error", func() {
				_, err := client.GetPoliciesByID("some-token", "some-id")
				Expect(err).To(MatchError("banana"))
			})
		})
		Context("when the json client gets a bad status code", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(&json_client.HttpResponseCodeError{
					StatusCode: http.StatusTeapot,
					Message:    "some-error",
				})
			})
			It("parses out the error body", func() {
				_, err := client.GetPoliciesByID("some-token", "some-id")
				Expect(err).To(MatchError("418 I'm a teapot: some-error"))
			})
		})
	})

	Describe("GetPoliciesV0", func() {
		BeforeEach(func() {
			jsonClient.DoStub = func(method, route string, reqData, respData interface{}, token string) error {
				respBytes := []byte(`{ "policies": [ {"source": { "id": "some-app-guid" }, "destination": { "id": "some-other-app-guid", "protocol": "tcp", "port": 8090 } } ] }`)
				json.Unmarshal(respBytes, respData)
				return nil
			}
		})
		It("does the right json http client request", func() {
			policies, err := client.GetPoliciesV0("some-token")
			Expect(err).NotTo(HaveOccurred())

			Expect(jsonClient.DoCallCount()).To(Equal(1))
			method, route, reqData, _, token := jsonClient.DoArgsForCall(0)
			Expect(method).To(Equal("GET"))
			Expect(route).To(Equal("/networking/v0/external/policies"))
			Expect(reqData).To(BeNil())

			Expect(policies).To(Equal([]PolicyV0{
				{
					Source: SourceV0{
						ID: "some-app-guid",
					},
					Destination: DestinationV0{
						ID:       "some-other-app-guid",
						Port:     8090,
						Protocol: "tcp",
					},
				},
			},
			))
			Expect(token).To(Equal("some-token"))
		})
		Context("when the json client fails", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(errors.New("banana"))
			})
			It("returns the error", func() {
				_, err := client.GetPoliciesV0("some-token")
				Expect(err).To(MatchError("banana"))
			})
		})
		Context("when the json client gets a bad status code", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(&json_client.HttpResponseCodeError{
					StatusCode: http.StatusTeapot,
					Message:    "some-error",
				})
			})
			It("parses out the error body", func() {
				_, err := client.GetPoliciesV0("some-token")
				Expect(err).To(MatchError("418 I'm a teapot: some-error"))
			})
		})
	})

	Describe("GetPoliciesV0ByID", func() {
		BeforeEach(func() {
			jsonClient.DoStub = func(method, route string, reqData, respData interface{}, token string) error {
				respBytes := []byte(`{ "policies": [ {"source": { "id": "some-app-guid" }, "destination": { "id": "some-other-app-guid", "protocol": "tcp", "port": 8090 } } ] }`)
				json.Unmarshal(respBytes, respData)
				return nil
			}
		})
		It("does the right json http client request", func() {
			policies, err := client.GetPoliciesV0ByID("some-token", "some-app-guid", "another-app-guid")
			Expect(err).NotTo(HaveOccurred())

			Expect(jsonClient.DoCallCount()).To(Equal(1))
			method, route, reqData, _, token := jsonClient.DoArgsForCall(0)
			Expect(method).To(Equal("GET"))
			Expect(route).To(Equal("/networking/v0/external/policies?id=some-app-guid,another-app-guid"))
			Expect(reqData).To(BeNil())
			Expect(policies).To(Equal([]PolicyV0{
				{
					Source: SourceV0{
						ID: "some-app-guid",
					},
					Destination: DestinationV0{
						ID:       "some-other-app-guid",
						Port:     8090,
						Protocol: "tcp",
					},
				},
			},
			))
			Expect(token).To(Equal("some-token"))
		})
		Context("when the json client fails", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(errors.New("banana"))
			})
			It("returns the error", func() {
				_, err := client.GetPoliciesV0ByID("some-token", "some-id")
				Expect(err).To(MatchError("banana"))
			})
		})
		Context("when the json client gets a bad status code", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(&json_client.HttpResponseCodeError{
					StatusCode: http.StatusTeapot,
					Message:    "some-error",
				})
			})
			It("parses out the error body", func() {
				_, err := client.GetPoliciesV0ByID("some-token", "some-id")
				Expect(err).To(MatchError("418 I'm a teapot: some-error"))
			})
		})
	})

	Describe("AddPoliciesV0", func() {
		var policiesToAdd []PolicyV0
		BeforeEach(func() {
			jsonClient.DoStub = func(method, route string, reqData, respData interface{}, token string) error {
				respBytes := []byte(`{}`)
				json.Unmarshal(respBytes, respData)
				return nil
			}

			policiesToAdd = []PolicyV0{
				{
					Source: SourceV0{
						ID: "some-app-guid",
					},
					Destination: DestinationV0{
						ID:       "some-other-app-guid",
						Port:     8090,
						Protocol: "tcp",
					},
				},
			}
		})
		It("does the right json http client request and passes the authorization token", func() {
			err := client.AddPoliciesV0("some-token", policiesToAdd)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeChunker.ChunkCallCount()).To(Equal(1))
			Expect(fakeChunker.ChunkArgsForCall(0)).To(Equal(policiesToAdd))

			Expect(jsonClient.DoCallCount()).To(Equal(2))
			method, route, reqData, _, token := jsonClient.DoArgsForCall(0)
			Expect(method).To(Equal("POST"))
			Expect(route).To(Equal("/networking/v0/external/policies"))
			Expect(reqData).To(Equal(map[string][]PolicyV0{
				"policies": []PolicyV0{{
					Source: SourceV0{
						ID: "some-app-guid",
					},
					Destination: DestinationV0{
						ID:       "some-other-app-guid",
						Port:     8090,
						Protocol: "tcp",
					},
				},
					{
						Source: SourceV0{
							ID: "some-app-guid-2",
						},
						Destination: DestinationV0{
							ID:       "some-other-app-guid-2",
							Port:     8091,
							Protocol: "tcp",
						},
					},
				}},
			))
			Expect(token).To(Equal("some-token"))

			method, route, reqData, _, token = jsonClient.DoArgsForCall(1)
			Expect(method).To(Equal("POST"))
			Expect(route).To(Equal("/networking/v0/external/policies"))
			Expect(reqData).To(Equal(map[string][]PolicyV0{
				"policies": []PolicyV0{
					{
						Source: SourceV0{
							ID: "some-app-guid-3",
						},
						Destination: DestinationV0{
							ID:       "some-other-app-guid-3",
							Port:     8092,
							Protocol: "tcp",
						},
					},
				},
			},
			))
			Expect(token).To(Equal("some-token"))
		})
		Context("when the json client fails", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(errors.New("banana"))
			})
			It("returns the error", func() {
				err := client.AddPoliciesV0("some-token", policiesToAdd)
				Expect(err).To(MatchError("banana"))
			})
		})
		Context("when the json client gets a bad status code", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(&json_client.HttpResponseCodeError{
					StatusCode: http.StatusTeapot,
					Message:    "some-error",
				})
			})
			It("parses out the error body", func() {
				err := client.AddPoliciesV0("some-token", policiesToAdd)
				Expect(err).To(MatchError("418 I'm a teapot: some-error"))
			})
		})
	})

	Describe("AddPolicies", func() {
		var policiesToAdd []Policy
		BeforeEach(func() {
			jsonClient.DoStub = func(method, route string, reqData, respData interface{}, token string) error {
				respBytes := []byte(`{}`)
				json.Unmarshal(respBytes, respData)
				return nil
			}

			policiesToAdd = []Policy{{
				Source: Source{
					ID: "some-app-guid",
				},
				Destination: Destination{
					ID:       "some-other-app-guid",
					Ports:    Ports{Start: 8080, End: 8090},
					Protocol: "tcp",
				},
			},
				{
					Source: Source{
						ID: "some-app-guid-2",
					},
					Destination: Destination{
						ID:       "some-other-app-guid-2",
						Ports:    Ports{Start: 8091, End: 8100},
						Protocol: "tcp",
					},
				},
			}
		})
		It("does the right json http client request and passes the authorization token", func() {
			err := client.AddPolicies("some-token", policiesToAdd)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeChunker.ChunkCallCount()).To(Equal(0))

			Expect(jsonClient.DoCallCount()).To(Equal(1))
			method, route, reqData, _, token := jsonClient.DoArgsForCall(0)
			Expect(method).To(Equal("POST"))
			Expect(route).To(Equal("/networking/v1/external/policies"))
			Expect(reqData).To(Equal(map[string][]Policy{
				"policies": []Policy{{
					Source: Source{
						ID: "some-app-guid",
					},
					Destination: Destination{
						ID:       "some-other-app-guid",
						Ports:    Ports{Start: 8080, End: 8090},
						Protocol: "tcp",
					},
				},
					{
						Source: Source{
							ID: "some-app-guid-2",
						},
						Destination: Destination{
							ID:       "some-other-app-guid-2",
							Ports:    Ports{Start: 8091, End: 8100},
							Protocol: "tcp",
						},
					},
				}},
			))
			Expect(token).To(Equal("some-token"))
		})
		Context("when the json client fails", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(errors.New("banana"))
			})
			It("returns the error", func() {
				err := client.AddPolicies("some-token", policiesToAdd)
				Expect(err).To(MatchError("banana"))
			})
		})
		Context("when the json client gets a bad status code", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(&json_client.HttpResponseCodeError{
					StatusCode: http.StatusTeapot,
					Message:    "some-error",
				})
			})
			It("parses out the error body", func() {
				err := client.AddPolicies("some-token", policiesToAdd)
				Expect(err).To(MatchError("418 I'm a teapot: some-error"))
			})
		})
	})

	Describe("DeletePoliciesV0", func() {
		var policiesToDelete []PolicyV0

		BeforeEach(func() {
			jsonClient.DoStub = func(method, route string, reqData, respData interface{}, token string) error {
				respBytes := []byte(`{}`)
				json.Unmarshal(respBytes, respData)
				return nil
			}

			policiesToDelete = []PolicyV0{
				{
					Source: SourceV0{
						ID: "some-app-guid",
					},
					Destination: DestinationV0{
						ID:       "some-other-app-guid",
						Port:     8090,
						Protocol: "tcp",
					},
				},
			}
		})
		It("does the right json http client request", func() {
			err := client.DeletePoliciesV0("some-token", policiesToDelete)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeChunker.ChunkCallCount()).To(Equal(1))
			Expect(fakeChunker.ChunkArgsForCall(0)).To(Equal(policiesToDelete))

			Expect(jsonClient.DoCallCount()).To(Equal(2))
			method, route, reqData, _, token := jsonClient.DoArgsForCall(0)
			Expect(method).To(Equal("POST"))
			Expect(route).To(Equal("/networking/v0/external/policies/delete"))
			Expect(reqData).To(Equal(map[string][]PolicyV0{
				"policies": []PolicyV0{{
					Source: SourceV0{
						ID: "some-app-guid",
					},
					Destination: DestinationV0{
						ID:       "some-other-app-guid",
						Port:     8090,
						Protocol: "tcp",
					},
				},
					{
						Source: SourceV0{
							ID: "some-app-guid-2",
						},
						Destination: DestinationV0{
							ID:       "some-other-app-guid-2",
							Port:     8091,
							Protocol: "tcp",
						},
					},
				}},
			))
			Expect(token).To(Equal("some-token"))

			method, route, reqData, _, token = jsonClient.DoArgsForCall(1)
			Expect(method).To(Equal("POST"))
			Expect(route).To(Equal("/networking/v0/external/policies/delete"))
			Expect(reqData).To(Equal(map[string][]PolicyV0{
				"policies": []PolicyV0{
					{
						Source: SourceV0{
							ID: "some-app-guid-3",
						},
						Destination: DestinationV0{
							ID:       "some-other-app-guid-3",
							Port:     8092,
							Protocol: "tcp",
						},
					},
				},
			},
			))
			Expect(token).To(Equal("some-token"))
		})
		Context("when the json client fails", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(errors.New("banana"))
			})
			It("returns the error", func() {
				err := client.DeletePoliciesV0("some-token", policiesToDelete)
				Expect(err).To(MatchError("banana"))
			})
		})
		Context("when the json client gets a bad status code", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(&json_client.HttpResponseCodeError{
					StatusCode: http.StatusTeapot,
					Message:    "some-error",
				})
			})
			It("parses out the error body", func() {
				err := client.DeletePoliciesV0("some-token", policiesToDelete)
				Expect(err).To(MatchError("418 I'm a teapot: some-error"))
			})
		})
	})

	Describe("DeletePolicies", func() {
		var policiesToDelete []Policy

		BeforeEach(func() {
			jsonClient.DoStub = func(method, route string, reqData, respData interface{}, token string) error {
				respBytes := []byte(`{}`)
				json.Unmarshal(respBytes, respData)
				return nil
			}

			policiesToDelete = []Policy{
				{
					Source: Source{
						ID: "some-app-guid",
					},
					Destination: Destination{
						ID: "some-other-app-guid",
						Ports: Ports{
							Start: 1234,
							End:   2345,
						},
						Protocol: "tcp",
					},
				},
			}
		})
		It("does the right json http client request", func() {
			err := client.DeletePolicies("some-token", policiesToDelete)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeChunker.ChunkCallCount()).To(Equal(0))

			Expect(jsonClient.DoCallCount()).To(Equal(1))
			method, route, reqData, _, token := jsonClient.DoArgsForCall(0)
			Expect(method).To(Equal("POST"))
			Expect(route).To(Equal("/networking/v1/external/policies/delete"))
			Expect(reqData).To(Equal(map[string][]Policy{
				"policies": []Policy{
					{
						Source: Source{
							ID: "some-app-guid",
						},
						Destination: Destination{
							ID: "some-other-app-guid",
							Ports: Ports{
								Start: 1234,
								End:   2345,
							},
							Protocol: "tcp",
						},
					},
				},
			},
			))
			Expect(token).To(Equal("some-token"))
		})
		Context("when the json client fails", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(errors.New("banana"))
			})
			It("returns the error", func() {
				err := client.DeletePolicies("some-token", policiesToDelete)
				Expect(err).To(MatchError("banana"))
			})
		})
		Context("when the json client gets a bad status code", func() {
			BeforeEach(func() {
				jsonClient.DoReturns(&json_client.HttpResponseCodeError{
					StatusCode: http.StatusTeapot,
					Message:    "some-error",
				})
			})
			It("parses out the error body", func() {
				err := client.DeletePolicies("some-token", policiesToDelete)
				Expect(err).To(MatchError("418 I'm a teapot: some-error"))
			})
		})
	})
})
