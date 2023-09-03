package pii

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDetect(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DetectPII Suite")
}

var _ = Describe("DetectPII", func() {
	When("obfuscate is true", func() {
		It("Should detect IPv4 and obfuscate", func() {
			text := []byte("My IP is 192.168.1.1.")
			Expect(DetectPII(text, true)).To(BeTrue())
			Expect(string(text)).To(Equal("My IP is ***********."))
		})

		It("Should detect email and obfuscate", func() {
			text := []byte("My email is example@email.com.")
			Expect(DetectPII(text, true)).To(BeTrue())
			Expect(string(text)).To(Equal("My email is *****************."))
		})
	})

	When("obfuscate is false", func() {
		It("Should detect IPv4 but not obfuscate", func() {
			text := []byte("My IP is 192.168.1.1.")
			Expect(DetectPII(text, false)).To(BeTrue())
			Expect(string(text)).To(Equal("My IP is 192.168.1.1."))
		})

		It("Should detect email but not obfuscate", func() {
			text := []byte("My email is example@email.com.")
			Expect(DetectPII(text, false)).To(BeTrue())
			Expect(string(text)).To(Equal("My email is example@email.com."))
		})
	})

	When("no PII is present", func() {
		It("Should return false and not change the text", func() {
			text := []byte("No secrets here.")
			Expect(DetectPII(text, true)).To(BeFalse())
			Expect(string(text)).To(Equal("No secrets here."))
		})
	})

	When("multiple types of PII are present", func() {
		It("Should detect all and obfuscate if true", func() {
			text := []byte("IP 192.168.1.1 and email example@email.com.")
			Expect(DetectPII(text, true)).To(BeTrue())
			Expect(string(text)).To(Equal("IP *********** and email *****************."))
		})

		It("Should detect all but not obfuscate if false", func() {
			text := []byte("IP 192.168.1.1 and email example@email.com.")
			Expect(DetectPII(text, false)).To(BeTrue())
			Expect(string(text)).To(Equal("IP 192.168.1.1 and email example@email.com."))
		})
	})

	When("edge case with text that looks like PII but isn't", func() {
		It("Should return false for false PII", func() {
			text := []byte("This is not.an.email and IP 299.299.299.")
			Expect(DetectPII(text, false)).To(BeFalse())
			Expect(string(text)).To(Equal("This is not.an.email and IP 299.299.299."))
		})
	})

	When("detecting AWS Env Var", func() {
		It("should detect and obfuscate", func() {
			text := []byte("AWS_SECRET_ACCESS_KEY=myawssecretkeyvalue")
			Expect(DetectPII(text, true)).To(BeTrue())
			Expect(string(text)).To(Equal("AWS_SECRET_ACCESS_KEY=*******************"))
		})
	})

	When("detecting GCP Env Var", func() {
		It("should detect and obfuscate", func() {
			text := []byte("GOOGLE_APPLICATION_CREDENTIALS=mygcpsecretkeyvalue")
			Expect(DetectPII(text, true)).To(BeTrue())
			Expect(string(text)).To(Equal("GOOGLE_APPLICATION_CREDENTIALS=*******************"))
		})
	})

	When("detecting Azure Env Var", func() {
		It("should detect and obfuscate", func() {
			text := []byte("AZURE_CLIENT_SECRET=myazuresecretkeyvalue")
			Expect(DetectPII(text, true)).To(BeTrue())
			Expect(string(text)).To(Equal("AZURE_CLIENT_SECRET=*********************"))
		})
	})

	When("detecting Database Password in URL", func() {
		It("should detect and obfuscate", func() {
			text := []byte("DATABASE_URL=postgres://user:password@host")
			Expect(DetectPII(text, true)).To(BeTrue())
			Expect(string(text)).To(Equal("DATABASE_URL=*****************************"))
		})
	})
})
