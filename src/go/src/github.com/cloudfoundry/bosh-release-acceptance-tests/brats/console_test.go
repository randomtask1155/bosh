package brats_test

import (
	"os/exec"
	"syscall"
	"time"

	"github.com/kr/pty"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("director console", func() {
	BeforeEach(func() {
		startInnerBosh()
	})

	It("allows a user to launch the director console", func() {
		ptyF, ttyF, err := pty.Open()
		Expect(err).ShouldNot(HaveOccurred())
		defer ptyF.Close()

		consoleCmd := exec.Command(outerBoshBinaryPath, "-d", "bosh", "ssh", "bosh")
		consoleCmd.Stdin = ttyF
		consoleCmd.Stdout = ttyF
		consoleCmd.Stderr = ttyF
		consoleCmd.SysProcAttr = &syscall.SysProcAttr{Setctty: true, Setsid: true}

		session, err := gexec.Start(consoleCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(ttyF.Close()).NotTo(HaveOccurred())

		Eventually(session.Out, 30*time.Second).Should(gbytes.Say(`bosh/[0-9a-f\-]{36}:~\$ `))

		_, err = ptyF.Write([]byte("sudo /var/vcap/jobs/director/bin/console\n"))
		Expect(err).NotTo(HaveOccurred())

		Eventually(session.Out, 30*time.Second).Should(gbytes.Say(`irb`))

		_, err = ptyF.Write([]byte("Bosh::Director::VERSION\n"))
		Expect(err).NotTo(HaveOccurred())

		Eventually(session.Out).Should(gbytes.Say(`"0.0.0"`))

		_, err = ptyF.Write([]byte("quit\n"))
		Expect(err).NotTo(HaveOccurred())

		_, err = ptyF.Write([]byte("exit\n"))
		Expect(err).NotTo(HaveOccurred())

		Eventually(session, time.Minute).Should(gexec.Exit(0))
	})
})
