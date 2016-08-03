// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package nginx

import (
	"bytes"

	"time"

	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NGINX", func() {

	var (
		gen        Generator
		writer     *bytes.Buffer
		db         database.Tenant
		lastUpdate *time.Time
		entry      resources.TenantEntry
		id         string
	)

	Context("NGINX", func() {

		BeforeEach(func() {
			id = "abcdef"
			writer = bytes.NewBuffer([]byte{})
			db = database.NewTenant(database.NewMemoryCloudantDB())
			entry = resources.TenantEntry{
				ProxyConfig: resources.ProxyConfig{
					Filters: resources.Filters{
						Rules:    []resources.Rule{},
						Versions: []resources.Version{},
					},
					LoadBalance: "round_robin",
				},
				BasicEntry: resources.BasicEntry{
					ID: id,
				},
				ServiceCatalog: resources.ServiceCatalog{
					Services:   []resources.Service{},
					LastUpdate: time.Now(),
				},
			}

			var err error
			gen, err = NewGenerator(Config{
				Database: db,
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("generates base config", func() {

			Expect(db.Create(entry)).ToNot(HaveOccurred())
			templateConf, err := gen.Generate(id, lastUpdate)
			Expect(err).ToNot(HaveOccurred())

			Expect(templateConf).ToNot(BeNil())
		})

		It("generates a valid NGINX conf", func() {
			// Setup input data
			services := []resources.Service{
				resources.Service{
					Name: "ServiceA",
					Endpoints: []resources.Endpoint{
						resources.Endpoint{
							Type:  "http",
							Value: "127.0.0.1:1234",
							Metadata: resources.MetaData{
								Version: "v2",
							},
						},
						resources.Endpoint{
							Type:  "http",
							Value: "127.0.0.5:1234",
							Metadata: resources.MetaData{
								Version: "v1",
							},
						},
					},
				},
				resources.Service{
					Name: "ServiceC",
					Endpoints: []resources.Endpoint{
						resources.Endpoint{
							Type:  "http",
							Value: "127.0.0.5:1234",
						},
					},
				},
			}
			entry.ServiceCatalog.Services = services

			entry.ProxyConfig.Filters.Rules = []resources.Rule{
				resources.Rule{
					Source:           "source",
					Destination:      "ServiceA",
					Delay:            0.3,
					DelayProbability: 0.9,
					ReturnCode:       501,
					AbortProbability: 0.1,
					Pattern:          "header_value",
					Header:           "header_name",
				},
			}
			entry.ProxyConfig.Filters.Versions = []resources.Version{
				resources.Version{
					Selectors: "{v2={weight=0.25}}",
					Service:   "ServiceA",
					Default:   "v1",
				},
			}

			// FIXME need to set aborts/delay to something and test it!
			//manager.GetVal.Filters

			Expect(db.Create(entry)).ToNot(HaveOccurred())

			// Generate the NGINX conf
			_, err := gen.Generate(id, lastUpdate)
			Expect(err).ToNot(HaveOccurred())

			// Verify the result...
			//for _, templ := range confTemplate.Proxies {
			//	if templ.ServiceName == "ServiceA" {
			//		Expect(templ.VersionDefault).To(Equal("v1"))
			//		Expect(templ.VersionSelectors).To(Equal("{v2={weight=0.25}}"))
			//		for _, version := range templ.Versions {
			//			Expect(version.UpstreamName).To(Or(Equal("ServiceA_v2"), Equal("ServiceA_v1")))
			//			Expect(len(version.Upstreams)).To(Equal(1))
			//			Expect(version.Upstreams[0]).To(Or(Equal("127.0.0.5:1234"), Equal("127.0.0.1:1234")))
			//		}
			//		Expect(len(templ.Rules)).To(Equal(1))
			//		Expect(templ.Rules[0].Delay).To(Equal(0.3))
			//		Expect(templ.Rules[0].DelayProbability).To(Equal(0.9))
			//		Expect(templ.Rules[0].ReturnCode).To(Equal(501))
			//		Expect(templ.Rules[0].AbortProbability).To(Equal(0.1))
			//	}
			//	if templ.ServiceName == "ServiceC" {
			//		Expect(templ.VersionDefault).To(Equal("UNVERSIONED"))
			//		//Expect(templ.VersionSelectors).To(HaveLen(0))
			//		Expect(templ.Rules).To(HaveLen(0))
			//		Expect(templ.Versions[0].UpstreamName).To(Equal("ServiceC_UNVERSIONED"))
			//		Expect(templ.Versions[0].Upstreams).To(HaveLen(1))
			//		Expect(len(templ.Rules)).To(Equal(0))
			//	}
			//}
			//Expect(confTemplate.Proxies[0].).To(ContainSubstring("127.0.0.1:1234")) // HTTP
			//
			//Expect(confTemplate.Proxies).To(ContainSubstring("127.0.0.1:1234"))
			//
			//
			//// Ensure proxy was generated for ServiceA and ServiceC
			//Expect(confTemplate).To(ContainSubstring("location /ServiceA/"))
			//Expect(confTemplate).To(ContainSubstring("upstream ServiceA_v2"))
			//Expect(confTemplate).To(ContainSubstring("upstream ServiceA_v1"))
			//Expect(confTemplate).To(ContainSubstring("ngx.var.target = splib.get_target(\"ServiceA\", \"v1\", {v2={weight=0.25}})"))
			//Expect(confTemplate).To(ContainSubstring("ngx.sleep(0.3)"))
			//Expect(confTemplate).To(ContainSubstring("ngx.exit(501)"))
			//
			//Expect(confTemplate).To(ContainSubstring("location /ServiceC/"))
			//Expect(confTemplate).To(ContainSubstring("upstream ServiceC_UNVERSIONED"))
			//Expect(confTemplate).To(ContainSubstring("ngx.var.target = splib.get_target(\"ServiceC\", \"UNVERSIONED\", nil)"))
		})
	})
})
