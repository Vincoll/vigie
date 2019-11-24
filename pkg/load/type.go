package load

const defaultTestSuitePath = "test/testsuite"
const defaultVariablePath = "test/variable"

type ConfGit struct {
	Clone bool `toml:"clone"`
	// Final Branch to checkout
	Branch string `toml:"branch"`
	// URL Repo Git
	Repo string `toml:"repo"`
	// Destination Path to clone
	Path string `toml:"path"`
	// Allow x509: certificate signed by unknown authority
	Insecure bool `toml:"insecure"`
}

type ConfTestfiles struct {
	Included []string `toml:"included"`
	Excluded []string `toml:"excluded"`
}
type ConfVariables struct {
	Variables []string `toml:"variables"`
	Included  []string `toml:"included"`
	Excluded  []string `toml:"excluded"`
	Fromenv   bool     `toml:"fromenv"`
}
