package brats_test

import (
	"io/ioutil"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"testing"

	"os/exec"
	"time"

	"path/filepath"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

const BLOBSTORE_ACCESS_LOG = "/var/vcap/sys/log/blobstore/blobstore_access.log"

func TestBrats(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Brats Suite")
}

var (
	outerBoshBinaryPath,
	boshBinaryPath,
	innerBoshPath,
	innerBoshJumpboxPrivateKeyPath,
	innerDirectorIP,
	boshRelease,
	directorBackupName,
	innerDirectorUser,
	deploymentName,
	boshDirectorReleasePath,
	candidateWardenLinuxStemcellPath,
	stemcellOS,
	dnsReleasePath string
)

var _ = BeforeSuite(func() {
	outerBoshBinaryPath = assertEnvExists("BOSH_BINARY_PATH")

	deploymentName = "dns-with-templates"
	directorBackupName = "director-backup"
	innerDirectorUser = "jumpbox"
	innerBoshPath = "/tmp/inner-bosh/director/"
	boshBinaryPath = filepath.Join(innerBoshPath, "bosh")
	innerBoshJumpboxPrivateKeyPath = filepath.Join(innerBoshPath, "jumpbox_private_key.pem")
	boshRelease = assertEnvExists("BOSH_RELEASE")
	innerDirectorIP = "10.245.0.34"
	dnsReleasePath = assertEnvExists("DNS_RELEASE_PATH")
	boshDirectorReleasePath = assertEnvExists("BOSH_DIRECTOR_RELEASE_PATH")
	candidateWardenLinuxStemcellPath = assertEnvExists("CANDIDATE_STEMCELL_TARBALL_PATH")
	stemcellOS = assertEnvExists("STEMCELL_OS")

	assertEnvExists("BOSH_ENVIRONMENT")
	assertEnvExists("BOSH_DEPLOYMENT_PATH")
})

var _ = AfterSuite(func() {
	session, err := gexec.Start(exec.Command(outerBoshBinaryPath, "-n", "clean-up", "--all"), GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Eventually(session, time.Minute).Should(gexec.Exit(0))
})

var _ = AfterEach(func() {
	By("cleanin up deployments")
	session := bosh("deployments", "--column=name")
	Eventually(session, 1*time.Minute).Should(gexec.Exit())
	deployments := strings.Fields(string(session.Out.Contents()))

	for _, deploymentName := range deployments {
		By(fmt.Sprintf("deleting deployment %v", deploymentName))
		if deploymentName == "" {
			continue
		}
		session := bosh("delete-deployment", "-n", "-d", deploymentName)
		Eventually(session, 5*time.Minute).Should(gexec.Exit())
	}
})

func assertEnvExists(envName string) string {
	val, found := os.LookupEnv(envName)
	if !found {
		Fail(fmt.Sprintf("Expected %s", envName))
	}
	return val
}

func startInnerBosh(args ...string) {
	startInnerBoshWithExpectation(false, "", args...)
}

func startInnerBoshWithExpectation(expectedFailure bool, expectedErrorToMatch string, args ...string) {
	effectiveArgs := args
	if stemcellOS == "ubuntu-xenial" {
		effectiveArgs = append(args, "-o", assetPath("inner-bosh-xenial-ops.yml"))
	}

	cmd := exec.Command(fmt.Sprintf("../../../../../../../ci/docker/main-bosh-docker/start-inner-bosh.sh"), effectiveArgs...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("bosh_release_path=%s", boshDirectorReleasePath))

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	if expectedFailure {
		Eventually(session, 25*time.Minute).Should(gbytes.Say(expectedErrorToMatch))
		Eventually(session, 25*time.Minute).Should(gexec.Exit(1))
	} else {
		Eventually(session, 25*time.Minute).Should(gexec.Exit(0))
	}
}

func stopInnerBosh() {
	session, err := gexec.Start(exec.Command("../../../../../../../ci/docker/main-bosh-docker/destroy-inner-bosh.sh"), GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Eventually(session, 15*time.Minute).Should(gexec.Exit(0))
}

func assetPath(filename string) string {
	path, err := filepath.Abs("../assets/" + filename)
	Expect(err).ToNot(HaveOccurred())

	return path
}

func boshDeploymentAssetPath(assetPath string) string {
	return filepath.Join(os.Getenv("BOSH_DEPLOYMENT_PATH"), assetPath)
}

func execCommand(binaryPath string, args ...string) *gexec.Session {
	session, err := gexec.Start(
		exec.Command(binaryPath, args...),
		GinkgoWriter,
		GinkgoWriter,
	)

	Expect(err).ToNot(HaveOccurred())

	return session
}

func outerBosh(args ...string) *gexec.Session {
	return execCommand(outerBoshBinaryPath, args...)
}

func bosh(args ...string) *gexec.Session {
	return execCommand(boshBinaryPath, args...)
}

func uploadStemcell(stemcellUrl string) {
	session := bosh("-n", "upload-stemcell", stemcellUrl)
	Eventually(session, 5*time.Minute).Should(gexec.Exit(0))
}

func uploadRelease(releaseUrl string) {
	session := bosh("-n", "upload-release", releaseUrl)
	Eventually(session, 2*time.Minute).Should(gexec.Exit(0))
}

type ExternalDBConfig struct {
	Type     string
	Host     string
	User     string
	Password string
	DBName   string

	CACertPath     string
	ClientCertPath string
	ClientKeyPath  string

	ConnectionVarFile     string
	ConnectionOptionsFile string
}

func loadExternalDBConfig(DBaaS string, mutualTLSEnabled bool, tmpCertDir string) ExternalDBConfig {
	var databaseType string
	if strings.HasSuffix(DBaaS, "mysql") {
		databaseType = "mysql"
	} else {
		databaseType = "postgres"
	}

	config := ExternalDBConfig{
		Type:                  databaseType,
		Host:                  assertEnvExists(fmt.Sprintf("%s_EXTERNAL_DB_HOST", strings.ToUpper(DBaaS))),
		User:                  assertEnvExists(fmt.Sprintf("%s_EXTERNAL_DB_USER", strings.ToUpper(DBaaS))),
		Password:              assertEnvExists(fmt.Sprintf("%s_EXTERNAL_DB_PASSWORD", strings.ToUpper(DBaaS))),
		DBName:                assertEnvExists(fmt.Sprintf("%s_EXTERNAL_DB_NAME", strings.ToUpper(DBaaS))),
		ConnectionVarFile:     fmt.Sprintf("external_db/%s.yml", DBaaS),
		ConnectionOptionsFile: fmt.Sprintf("external_db/%s_connection_options.yml", DBaaS),
	}

	caContents, err := exec.Command(outerBoshBinaryPath, "int", assetPath(config.ConnectionVarFile), "--path", "/db_ca").Output()
	Expect(err).ToNot(HaveOccurred())
	caFile, err := ioutil.TempFile(tmpCertDir, "db_ca")
	Expect(err).ToNot(HaveOccurred())

	defer caFile.Close()
	_, err = caFile.Write(caContents)
	Expect(err).ToNot(HaveOccurred())

	config.CACertPath = caFile.Name()

	if mutualTLSEnabled {
		clientCertContents := assertEnvExists(fmt.Sprintf("%s_EXTERNAL_DB_CLIENT_CERTIFICATE", strings.ToUpper(DBaaS)))
		clientKeyContents := assertEnvExists(fmt.Sprintf("%s_EXTERNAL_DB_CLIENT_PRIVATE_KEY", strings.ToUpper(DBaaS)))

		clientCertFile, err := ioutil.TempFile(tmpCertDir, "client_cert")
		Expect(err).ToNot(HaveOccurred())

		defer clientCertFile.Close()
		_, err = clientCertFile.Write([]byte(clientCertContents))
		Expect(err).ToNot(HaveOccurred())

		clientKeyFile, err := ioutil.TempFile(tmpCertDir, "client_key")
		Expect(err).ToNot(HaveOccurred())

		defer clientKeyFile.Close()
		_, err = clientKeyFile.Write([]byte(clientKeyContents))
		Expect(err).ToNot(HaveOccurred())

		config.ClientCertPath = clientCertFile.Name()
		config.ClientKeyPath = clientKeyFile.Name()
	}

	return config
}

func innerBoshWithExternalDBOptions(dbConfig ExternalDBConfig) []string {
	options := []string{
		"-o", boshDeploymentAssetPath("misc/external-db.yml"),
		"-o", boshDeploymentAssetPath("experimental/db-enable-tls.yml"),
		"-o", assetPath(dbConfig.ConnectionOptionsFile),
		"--vars-file", assetPath(dbConfig.ConnectionVarFile),
		"-v", fmt.Sprintf("external_db_host=%s", dbConfig.Host),
		"-v", fmt.Sprintf("external_db_user=%s", dbConfig.User),
		"-v", fmt.Sprintf("external_db_password=%s", dbConfig.Password),
		"-v", fmt.Sprintf("external_db_name=%s", dbConfig.DBName),
	}

	if dbConfig.ClientCertPath != "" || dbConfig.ClientKeyPath != "" {
		options = append(options,
			fmt.Sprintf("-o %s", boshDeploymentAssetPath("experimental/db-enable-mutual-tls.yml")),
			fmt.Sprintf("-o %s", assetPath("tls-skip-host-verify.yml")),
			fmt.Sprintf("--var-file=db_client_certificate=%s", dbConfig.ClientCertPath),
			fmt.Sprintf("--var-file=db_client_private_key=%s", dbConfig.ClientKeyPath),
		)
	}

	return options
}
