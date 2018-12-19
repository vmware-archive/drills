package pre_test

import (
	"fmt"

	"github.com/concourse/go-concourse/concourse"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Disabled Versions", func() {
	var (
		guid *uuid.UUID
	)

	BeforeEach(func() {
		var err error
		guid, err = uuid.NewV4()
		Expect(err).ToNot(HaveOccurred())

		By("setting up two team")
		fly.Run("set-team", "-n", "team1-"+guid.String(), "--non-interactive", "--local-user="+parsedEnv.Username)
		fly.Run("set-team", "-n", "team2-"+guid.String(), "--non-interactive", "--local-user="+parsedEnv.Username)

		By("setting team 1s pipeline and creating two versions")
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team1-"+guid.String())
		fly.Run("sp", "-p", "resource1-"+guid.String(),
			"-c", "pipelines/release-resource.yml",
			"-y", "trigger=false",
			"-v", "hash="+guid.String(),
			"-n",
		)
		fly.Run("up", "-p", "resource1-"+guid.String())
		fly.Run("cr", "-r", fmt.Sprintf("resource1-%s/some-resource", guid.String()), "-f", "version:4.0.0")

		By("setting up team 2s pipeline and creating two versions")
		fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team2-"+guid.String())
		fly.Run("sp", "-p", "resource2-"+guid.String(),
			"-c", "pipelines/release-resource.yml",
			"-y", "trigger=false",
			"-v", "hash="+guid.String(),
			"-n",
		)
		fly.Run("up", "-p", "resource2-"+guid.String())
		fly.Run("cr", "-r", fmt.Sprintf("resource2-%s/some-resource", guid.String()), "-f", "version:4.0.0")
	})

	It("Disables some versions", func() {
		By("listing the resource versions for team 1")
		ccClient := login(parsedEnv.Endpoint, parsedEnv.Username, parsedEnv.Password)
		versions, _, found, err := ccClient.Team("team1-"+guid.String()).ResourceVersions("resource1-"+guid.String(), "some-resource", concourse.Page{})
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
		disabled, err := ccClient.Team("team1-"+guid.String()).DisableResourceVersion("resource1-"+guid.String(), "some-resource", versionID)
		Expect(disabled).To(BeTrue())
		Expect(err).ToNot(HaveOccurred())
		By("verifying the version is disabled")
		versions, _, found, err = ccClient.Team("team1-"+guid.String()).ResourceVersions("resource1-"+guid.String(), "some-resource", concourse.Page{})
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
		versions, _, found, err = ccClient.Team("team2-"+guid.String()).ResourceVersions("resource2-"+guid.String(), "some-resource", concourse.Page{})
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
		disabled, err = ccClient.Team("team2-"+guid.String()).DisableResourceVersion("resource2-"+guid.String(), "some-resource", versionID)
		Expect(disabled).To(BeTrue())
		Expect(err).ToNot(HaveOccurred())
		By("verifying the version is disabled")
		versions, _, found, err = ccClient.Team("team2-"+guid.String()).ResourceVersions("resource2-"+guid.String(), "some-resource", concourse.Page{})
		Expect(err).ToNot(HaveOccurred())
		Expect(found).To(BeTrue())
		for _, v := range versions {
			if v.Version["version"] == "4.1.0" {
				Expect(v.Enabled).To(BeFalse())
			}
		}
	})
})
