package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleHealthCheckTypeOverride", func() {
	var (
		originalManifest    pushmanifestparser.Manifest
		transformedManifest pushmanifestparser.Manifest
		overrides           FlagOverrides
		executeErr          error
	)

	BeforeEach(func() {
		originalManifest = pushmanifestparser.Manifest{}
		overrides = FlagOverrides{}
	})

	JustBeforeEach(func() {
		transformedManifest, executeErr = HandleHealthCheckTypeOverride(originalManifest, overrides)
	})

	When("manifest web process does not specify health check type", func() {
		BeforeEach(func() {
			originalManifest.Applications = []pushmanifestparser.Application{
				{
					Processes: []pushmanifestparser.Process{
						{Type: "web"},
					},
				},
			}
		})

		When("health check type is not set on the flag overrides", func() {
			It("does not change the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest).To(Equal(originalManifest))
			})
		})

		When("health check type set on the flag overrides", func() {
			BeforeEach(func() {
				overrides.HealthCheckType = constant.HTTP
			})

			It("changes the health check type of the web process in the manifest", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(transformedManifest.Applications).To(ConsistOf(
					pushmanifestparser.Application{
						Processes: []pushmanifestparser.Process{
							{Type: "web", HealthCheckType: constant.HTTP},
						},
					},
				))
			})
		})
	})

	When("health check type flag is set, and manifest app has non-web processes", func() {
		BeforeEach(func() {
			overrides.HealthCheckType = constant.HTTP

			originalManifest.Applications = []pushmanifestparser.Application{
				{
					Processes: []pushmanifestparser.Process{
						{Type: "worker", HealthCheckType: constant.Port},
					},
				},
			}
		})

		It("changes the health check type in the app level only", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				pushmanifestparser.Application{
					HealthCheckType: constant.HTTP,
					Processes: []pushmanifestparser.Process{
						{Type: "worker", HealthCheckType: constant.Port},
					},
				},
			))
		})
	})

	When("health check type flag is set, and manifest app has web and non-web processes", func() {
		BeforeEach(func() {
			overrides.HealthCheckType = constant.HTTP

			originalManifest.Applications = []pushmanifestparser.Application{
				{
					Processes: []pushmanifestparser.Process{
						{Type: "worker", HealthCheckType: constant.Port},
						{Type: "web", HealthCheckType: constant.Process},
					},
					HealthCheckType: constant.Port,
				},
			}
		})

		It("changes the health check type of the web process in the manifest", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(transformedManifest.Applications).To(ConsistOf(
				pushmanifestparser.Application{
					Processes: []pushmanifestparser.Process{
						{Type: "worker", HealthCheckType: constant.Port},
						{Type: "web", HealthCheckType: constant.HTTP},
					},
					HealthCheckType: constant.Port,
				},
			))
		})
	})

	When("health check type flag is set and there are multiple apps in the manifest", func() {
		BeforeEach(func() {
			overrides.HealthCheckType = constant.HTTP

			originalManifest.Applications = []pushmanifestparser.Application{
				{},
				{},
			}
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.CommandLineArgsWithMultipleAppsError{}))
		})
	})
})
