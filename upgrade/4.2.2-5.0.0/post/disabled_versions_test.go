package post_test

import (
	"fmt"

	"github.com/concourse/concourse/go-concourse/concourse"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const DisabledVersionsIdentifier = "disabled-versions"

var _ = Describe("Disabled Versions", func() {
	Context("when two teams had the same resources configured but with different versions disabled", func() {
		It("keeps the same versions disabled for each resource after the upgrade", func() {
			By("listing the resource versions for team 1")
			fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team1-"+DisabledVersionsIdentifier)
			fly.Run("cr", "-r", fmt.Sprintf("pipeline1-%s/some-resource", DisabledVersionsIdentifier), "-f", "version:4.0.0")
			ccClient := login(parsedEnv.Endpoint, parsedEnv.Username, parsedEnv.Password)
			versions, _, found, err := ccClient.Team("team1-"+DisabledVersionsIdentifier).ResourceVersions("pipeline1-"+DisabledVersionsIdentifier, "some-resource", concourse.Page{})
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(versions).ToNot(HaveLen(0))
			for _, v := range versions {
				if v.Version["version"] == "4.0.0" {
					Expect(v.Enabled).To(BeFalse())
				} else {
					Expect(v.Enabled).To(BeTrue())
				}
			}

			By("listing the resource versions for team 2")
			fly.Login(parsedEnv.Username, parsedEnv.Password, parsedEnv.Endpoint, "-n", "team2-"+DisabledVersionsIdentifier)
			fly.Run("cr", "-r", fmt.Sprintf("pipeline2-%s/some-resource", DisabledVersionsIdentifier), "-f", "version:4.0.0")
			versions, _, found, err = ccClient.Team("team2-"+DisabledVersionsIdentifier).ResourceVersions("pipeline2-"+DisabledVersionsIdentifier, "some-resource", concourse.Page{})
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(versions).ToNot(HaveLen(0))
			for _, v := range versions {
				if v.Version["version"] == "4.1.0" {
					Expect(v.Enabled).To(BeFalse())
				} else {
					Expect(v.Enabled).To(BeTrue())
				}
			}
		})
	})
})
