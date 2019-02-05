package pre_test

import (
	"fmt"

	"github.com/concourse/go-concourse/concourse"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const DisabledVersionsIdentifier = "disabled-versions"

var _ = Describe("Disabled Versions", func() {
	var (
		guid *uuid.UUID
	)

	BeforeEach(func() {
		var err error
		guid, err = uuid.NewV4()
		Expect(err).ToNot(HaveOccurred())

		By("setting up two team")
		fly.Run("set-team", "-n", "team1-"+DisabledVersionsIdentifier, "--non-interactive", "--local-user="+parsedEnv.Username)
		fly.Run("set-team", "-n", "team2-"+DisabledVersionsIdentifier, "--non-interactive", "--local-user="+parsedEnv.Username)

		By("setting team 1s pipeline and creating versions")
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team1-"+DisabledVersionsIdentifier)
		fly.Run("sp", "-p", "pipeline1-"+DisabledVersionsIdentifier,
			"-c", "../pipelines/release-resource.yml",
			"-y", "trigger=false",
			"-v", "hash="+guid.String(),
			"-n",
		)
		fly.Run("up", "-p", "pipeline1-"+DisabledVersionsIdentifier)
		fly.Run("cr", "-r", fmt.Sprintf("pipeline1-%s/some-resource", DisabledVersionsIdentifier), "-f", "version:4.0.0")

		By("setting up team 2s pipeline and creating versions")
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team2-"+DisabledVersionsIdentifier)
		fly.Run("sp", "-p", "pipeline2-"+DisabledVersionsIdentifier,
			"-c", "../pipelines/release-resource.yml",
			"-y", "trigger=false",
			"-v", "hash="+guid.String(),
			"-n",
		)
		fly.Run("up", "-p", "pipeline2-"+DisabledVersionsIdentifier)
		fly.Run("cr", "-r", fmt.Sprintf("pipeline2-%s/some-resource", DisabledVersionsIdentifier), "-f", "version:4.0.0")
	})

	It("Disables some versions", func() {
		By("listing the resource versions for team 1")
		ccClient := login(parsedEnv.Endpoint, parsedEnv.Username, parsedEnv.Password)
		versions, _, found, err := ccClient.Team("team1-"+DisabledVersionsIdentifier).ResourceVersions("pipeline1-"+DisabledVersionsIdentifier, "some-resource", concourse.Page{})
		Expect(err).ToNot(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(versions).ToNot(HaveLen(0))
		By("disabling the first version")
		var versionID int
		for _, v := range versions {
			if v.Version["version"] == "4.0.0" {
				versionID = v.ID
			}
		}
		By("disabling the 4.0.0 version for team 1")
		disabled, err := ccClient.Team("team1-"+DisabledVersionsIdentifier).DisableResourceVersion("pipeline1-"+DisabledVersionsIdentifier, "some-resource", versionID)
		Expect(disabled).To(BeTrue())
		Expect(err).ToNot(HaveOccurred())
		By("verifying the version is disabled")
		versions, _, found, err = ccClient.Team("team1-"+DisabledVersionsIdentifier).ResourceVersions("pipeline1-"+DisabledVersionsIdentifier, "some-resource", concourse.Page{})
		Expect(err).ToNot(HaveOccurred())
		Expect(found).To(BeTrue())
		for _, v := range versions {
			if v.Version["version"] == "4.0.0" {
				Expect(v.Enabled).To(BeFalse())
			}
		}

		versionID = 0

		By("listing the resource versions for team 2")
		ccClient = login(parsedEnv.Endpoint, parsedEnv.Username, parsedEnv.Password)
		versions, _, found, err = ccClient.Team("team2-"+DisabledVersionsIdentifier).ResourceVersions("pipeline2-"+DisabledVersionsIdentifier, "some-resource", concourse.Page{})
		Expect(err).ToNot(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(versions).ToNot(HaveLen(0))
		By("disabling the second version")
		for _, v := range versions {
			if v.Version["version"] == "4.1.0" {
				versionID = v.ID
			}
		}
		By("disabling the 4.1.0 version for team 2")
		disabled, err = ccClient.Team("team2-"+DisabledVersionsIdentifier).DisableResourceVersion("pipeline2-"+DisabledVersionsIdentifier, "some-resource", versionID)
		Expect(disabled).To(BeTrue())
		Expect(err).ToNot(HaveOccurred())
		By("verifying the version is disabled")
		versions, _, found, err = ccClient.Team("team2-"+DisabledVersionsIdentifier).ResourceVersions("pipeline2-"+DisabledVersionsIdentifier, "some-resource", concourse.Page{})
		Expect(err).ToNot(HaveOccurred())
		Expect(found).To(BeTrue())
		for _, v := range versions {
			if v.Version["version"] == "4.1.0" {
				Expect(v.Enabled).To(BeFalse())
			}
		}
	})
})
