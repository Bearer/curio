package url

import (
	"context"
	"errors"
	"net"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bearer/curio/pkg/report"
	"github.com/bearer/curio/pkg/report/detectors"
	"github.com/weppos/publicsuffix-go/publicsuffix"
)

const prefixPattern = "(?P<match>\\A(?:[^:]+://)?(?:[^/]+\\.)?"
const suffixPattern = "(?:/|\\z)"

// recipe url matching regexp
var regexpReplaceMatcher = regexp.MustCompile(`<\w+>`)
var regexpVariableMatcher = regexp.MustCompile(`\A[*\/.-:]+\z`)

// for domain resolution - find anything between * and .
var regexpDomainSplitMatcher = regexp.MustCompile(`\*\s*(.*?)\s*\.`)

// url validation regexp
var regexpDependencyFileMatcher = regexp.MustCompile(`Gemfile\.lock|package\.json|yarn\.lock|maven\-dependencies\.json|gemnasium\-maven\-plugin\.json|gradle\-dependencies\.json|Pipfile\.lock|package\-lock\.json|npm\-shrinkwrap\.json|packages\.lock\.json|project\.json|packages\.config|paket\.dependencies|ivy\-report\.xml|composer\.lock|composer\.json|pipdeptree\.json|go\.sum|requirements\.txt|pyproject\.toml|poetry\.lock|pom\.xml|build\.gradle`)
var regexpInvalidFilenameMatcher = regexp.MustCompile(`(trad|/translations?/|locales?|dockerfile|i18n)`)
var regexpValidPathMatcher = regexp.MustCompile(`\A[\w\-.*/?=&\[\]]+\z`)
var regexpInvalidExtensionsInPathMatcher = regexp.MustCompile(` /\.md|\.zip|\.css|\.csv|\.xls|\.sh|\.jpg|\.jpeg|\.png|\.pdf|\.htm|\.html|\.xhtml|\.txt|\.dtd|\.sql|\.xsd|\.gif|\.ico/i`)

var excludedDomains = map[string]struct{}{
	"k8s.io":     {},
	"jquery.com": {},
	"github.com": {},
}

var blocklistTLDs = map[string]struct{}{
	"id":       {},
	"name":     {},
	"country":  {},
	"email":    {},
	"ping":     {},
	"phone":    {},
	"link":     {},
	"menu":     {},
	"zip":      {},
	"domains":  {},
	"search":   {},
	"js":       {},
	"tf":       {},
	"host":     {},
	"dev":      {},
	"local":    {},
	"md":       {},
	"show":     {},
	"video":    {},
	"global":   {},
	"now":      {},
	"py":       {},
	"tab":      {},
	"so":       {},
	"cc":       {},
	"off":      {},
	"services": {},
	"money":    {},
	"org":      {},
}

var subdomainNotAllowed = map[string]struct{}{
	"media":      {},
	"community":  {},
	"example":    {},
	"schemas":    {},
	"static":     {},
	"wiki":       {},
	"dist":       {},
	"support":    {},
	"dl":         {},
	"download":   {},
	"docs":       {},
	"cloudfront": {},
	"img":        {},
	"packages":   {},
	"fonts":      {},
	"releases":   {},
	"msdn":       {},
	"downloads":  {},
	"images":     {},
	"updates":    {},
	"upload":     {},
	"mobile":     {},
	"demo":       {},
	"forum":      {},
	"video":      {},
	"doc":        {},
	"tools":      {},
	"www2":       {},
	"groups":     {},
	"shemas":     {},
	"widgets":    {},
	"feeds":      {},
	"modules":    {},
	"package":    {},
	"blogs":      {},
	"news":       {},
	"www":        {},
	"faq":        {},
	"cdn":        {},
}

var invalidLanguageTypes = map[string]struct{}{
	"markup": {},
}

var allowedFilenameExtensions = map[string]struct{}{
	"twig": {},
	"tpl":  {},
	"ejs":  {},
}

var restrictedDetectorTypes = map[string]struct{}{
	"simple": {},
}

var invalidFilenameExtensions = map[string]struct{}{
	"tf":          {},
	"sql":         {},
	"ipynb":       {},
	"sh":          {},
	"io":          {},
	"j2":          {},
	"feature":     {},
	"xsl":         {},
	"ps1":         {},
	"bzl":         {},
	"cake":        {},
	"dockerfile":  {},
	"hcl":         {},
	"yar":         {},
	"xslt":        {},
	"tfvars":      {},
	"sbt":         {},
	"1":           {},
	"mk":          {},
	"psm1":        {},
	"bat":         {},
	"linq":        {},
	"ftl":         {},
	"rockspec":    {},
	"fs":          {},
	"nomad":       {},
	"es":          {},
	"snippet":     {},
	"pb":          {},
	"bash":        {},
	"m":           {},
	"com":         {},
	"ascs":        {},
	"rtf":         {},
	"genrule_cmd": {},
	"csx":         {},
	"old":         {},
	"tmp":         {},
	"notused":     {},
	"pp":          {},
	"cmd":         {},
	"bundle":      {},
	"purs":        {},
}

var potentialDetectors = map[string]struct{}{
	"env_file":    {},
	"yaml_config": {},
}

var cloudDetectors = map[string]struct{}{
	"sql-readonly": {},
	"sql-advanced": {},
}

type ValidationState int

const (
	Valid ValidationState = iota + 1
	Invalid
	Potential
)

type ValidationResult struct {
	State  ValidationState
	Reason string
}

// given a valid URL, is it reachable?
func IsReachable(myURL string, timeout time.Duration) bool {
	resolverContext, cancel := context.WithTimeout(context.TODO(), timeout)
	defer cancel()

	staticDomain := getStaticDomain(myURL)
	if myURL == staticDomain {
		return domainResolves(myURL, resolverContext)
	}

	if isNameserver(myURL, staticDomain, resolverContext) {
		return true
	}

	return domainResolves(myURL, resolverContext)
}

func Match(url string, matcher *regexp.Regexp) (string, error) {
	match := matcher.FindStringSubmatch(url)
	if match != nil {
		for i, name := range matcher.SubexpNames() {
			if name == "match" {
				return match[i], nil
			}
		}
	}
	return "", nil
}

func PrepareRegexpMatcher(myURL string) (*regexp.Regexp, error) {
	parsedURL, err := url.Parse(myURL)
	if err != nil {
		return nil, err
	}

	parsedDomain, err := publicsuffix.ParseFromListWithOptions(
		publicsuffix.DefaultList,
		parsedURL.Host,
		&publicsuffix.FindOptions{IgnorePrivate: true, DefaultRule: nil},
	)
	if err != nil {
		return nil, err
	}

	return regexp.Compile(prefixPattern + domainPattern(parsedDomain) + pathPattern(parsedURL) + ")" + suffixPattern)
}

func PrepareURLValue(myURL string) (string, error) {
	if regexpVariableMatcher.MatchString(myURL) {
		return "", errors.New("URL is only made of variables")
	}

	var preparedURL string
	// replace placeholders with wildcard *
	preparedURL = strings.ReplaceAll(myURL, `%d`, "*")
	preparedURL = strings.ReplaceAll(preparedURL, `%s`, "*")
	preparedURL = regexpReplaceMatcher.ReplaceAllString(preparedURL, "*")

	// ensure scheme is present
	if strings.HasPrefix(preparedURL, "http://") || strings.HasPrefix(preparedURL, "https://") {
		return preparedURL, nil
	}
	if strings.Contains(preparedURL, ".") {
		preparedURL = "https://" + preparedURL
	}

	return preparedURL, nil
}

func Validate(myURL string, isInternal bool, data *report.Detection) (*ValidationResult, error) {
	ValidationResult := ValidationResult{
		State:  Invalid,
		Reason: "uncertain", // default
	}

	// todo: to be shared for all classifications
	if isVendored(data.Source.Filename) {
		ValidationResult.Reason = "included_in_vendor_folder"
		return &ValidationResult, nil
	}

	if isPotentialDetector(data.DetectorType) {
		ValidationResult.State = Potential
		ValidationResult.Reason = "potential_detectors"
		return &ValidationResult, nil
	}

	if myURL == "" {
		ValidationResult.Reason = "blank_url"
		return &ValidationResult, nil
	}

	parsedURL, err := url.Parse(myURL)
	if err != nil {
		return nil, err
	}

	if parsedURL.Host == "" {
		ValidationResult.Reason = "no_host_error"
		return &ValidationResult, nil
	}

	if net.ParseIP(parsedURL.Host) != nil {
		ValidationResult.Reason = "ip_address_error"
		return &ValidationResult, nil
	}

	if regexpDependencyFileMatcher.MatchString(data.Source.Filename) {
		ValidationResult.Reason = "dependency_file_error"
		return &ValidationResult, nil
	}

	filenameExtension := strings.TrimPrefix(strings.TrimSpace(filepath.Ext(data.Source.Filename)), ".")
	if invalidLanguageType(filenameExtension, data) {
		ValidationResult.Reason = "language_type_error"
		return &ValidationResult, nil
	}

	if invalidLanguage(filenameExtension) {
		ValidationResult.Reason = "invalid_language_error"
		return &ValidationResult, nil
	}

	if regexpInvalidFilenameMatcher.MatchString(data.Source.Filename) {
		ValidationResult.Reason = "filename_error"
		return &ValidationResult, nil
	}

	if parsedURL.Host == "" || parsedURL.Host[0:1] == "." {
		ValidationResult.Reason = "tld_error"
		return &ValidationResult, nil
	}

	parsedDomain, err := publicsuffix.ParseFromListWithOptions(
		publicsuffix.DefaultList,
		parsedURL.Host,
		&publicsuffix.FindOptions{IgnorePrivate: true, DefaultRule: nil},
	)
	if err != nil {
		return nil, err
	}

	if !isInternal {
		if parsedDomain.TLD == "" || isBlocklisted(parsedDomain.TLD) {
			ValidationResult.Reason = "tld_error"
			return &ValidationResult, nil
		}

		// todo: :invalid, :domain_not_reachable

		if domainIsExcluded(parsedURL.Host) {
			ValidationResult.Reason = "excluded_domains_error"
			return &ValidationResult, nil
		}
	}

	if subdomainIsNotAllowed(parsedDomain.TRD) {
		if isInternal {
			ValidationResult.Reason = "internal_domain_subdomain_error"
		} else {
			ValidationResult.Reason = "subdomain_error"
		}
		return &ValidationResult, nil
	}

	if pathError(parsedURL.Path) {
		if isInternal {
			ValidationResult.Reason = "internal_domain_errors_in_path"
		} else {
			ValidationResult.Reason = "errors_in_path"
		}
		return &ValidationResult, nil
	}

	if pathContainsAPIorAuth(parsedURL.Path) {
		ValidationResult.State = Valid
		if isInternal {
			ValidationResult.Reason = "internal_domain_path_contains_api_or_auth"
		} else {
			ValidationResult.Reason = "path_contains_api_or_auth"
		}
		return &ValidationResult, nil
	}

	if isInternal {
		if parsedDomain.TRD == "" {
			ValidationResult.Reason = "internal_domain_but_no_subdomain"
			return &ValidationResult, nil
		}

		if strings.Contains(parsedURL.Host, "*") {
			ValidationResult.Reason = "internal_domain_and_subdomain_with_wildcard"
			ValidationResult.State = Potential
			return &ValidationResult, nil
		}

		ValidationResult.Reason = "internal_domain_and_subdomain"
		ValidationResult.State = Valid
		return &ValidationResult, nil
	} else {
		if parsedDomain.TRD == "" {
			ValidationResult.State = Potential
			ValidationResult.Reason = "no_subdomain"
			return &ValidationResult, nil
		}

		if strings.Contains(parsedDomain.TRD, "api") {
			if strings.Contains(parsedURL.Host, "*") {
				ValidationResult.State = Potential
				ValidationResult.Reason = "subdomain_contains_api_with_wildcard"
				return &ValidationResult, nil
			}

			ValidationResult.State = Valid
			ValidationResult.Reason = "subdomain_contains_api"
			return &ValidationResult, nil
		}
	}

	ValidationResult.State = Potential // uncertain
	return &ValidationResult, nil
}

func domainPattern(parsedDomain *publicsuffix.DomainName) string {
	var domainPatterns []string
	if parsedDomain.TRD != "" {
		domainPatterns = append(domainPatterns, wildcardPattern(parsedDomain.TRD, "."))
	}
	if parsedDomain.SLD != "" {
		domainPatterns = append(domainPatterns, regexp.QuoteMeta(parsedDomain.SLD))
	}
	if parsedDomain.TLD != "" {
		domainPatterns = append(domainPatterns, regexp.QuoteMeta(parsedDomain.TLD))
	}

	return strings.Join(domainPatterns, "\\.")
}

func pathPattern(parsedURL *url.URL) string {
	return wildcardPattern(strings.TrimSuffix(parsedURL.Path, "/"), "/")
}

func wildcardPattern(myString string, separator string) string {
	var parts []string
	for _, part := range strings.Split(myString, separator) {
		parts = append(parts, "("+strings.ReplaceAll(regexp.QuoteMeta(part), "\\*", ".+")+"|\\*)")
	}
	return strings.Join(parts, regexp.QuoteMeta(separator))
}

func domainIsExcluded(host string) bool {
	_, ok := excludedDomains[host]
	return ok
}

func isBlocklisted(tld string) bool {
	_, ok := blocklistTLDs[tld]
	return ok
}

func subdomainIsNotAllowed(trd string) bool {
	_, ok := subdomainNotAllowed[trd]
	return ok
}

func invalidLanguageType(filenameExtension string, data *report.Detection) bool {
	_, invalidLanguageType := invalidLanguageTypes[data.Source.LanguageType]
	_, validFilenameExtension := allowedFilenameExtensions[filenameExtension]
	_, restrictedDetectorType := restrictedDetectorTypes[string(data.DetectorType)]

	return invalidLanguageType && !validFilenameExtension && restrictedDetectorType
}

func invalidLanguage(filenameExtension string) bool {
	_, ok := invalidFilenameExtensions[filenameExtension]
	return ok
}

func pathError(path string) bool {
	if path == "" {
		return false
	}

	return !regexpValidPathMatcher.MatchString(path) || regexpInvalidExtensionsInPathMatcher.MatchString(path)
}

func pathContainsAPIorAuth(path string) bool {
	if path == "" {
		return false
	}

	return strings.Contains(path, "api") || strings.Contains(path, "auth")
}

func isVendored(filename string) bool {
	_, ok := cloudDetectors[filename]
	if ok {
		return false
	}

	return strings.Contains(filename, "vendor/")
}

func isPotentialDetector(detectorType detectors.Type) bool {
	_, ok := potentialDetectors[string(detectorType)]
	return ok
}

func getStaticDomain(myURL string) string {
	domainSplit := regexpDomainSplitMatcher.Split(myURL, -1)
	lenDomainSplit := len(domainSplit)
	if lenDomainSplit <= 1 {
		// single part, no static domain
		return myURL
	}

	return domainSplit[lenDomainSplit-1]
}

func isNameserver(myURL string, staticDomain string, resolverContext context.Context) bool {
	_, err := publicsuffix.ParseFromListWithOptions(
		publicsuffix.DefaultList,
		staticDomain,
		&publicsuffix.FindOptions{IgnorePrivate: true, DefaultRule: nil},
	)
	if err != nil {
		// invalid domain
		return false
	}

	nameserver, err := resolverLookupNS(resolverContext, staticDomain)
	if err != nil {
		// return false even for transient errors
		return false
	}

	return len(nameserver) > 0
}

func domainResolves(myURL string, resolverContext context.Context) bool {
	// handle any special characters
	sanitizedURL, err := publicsuffix.ToASCII(myURL)
	if err != nil {
		return false
	}

	names, err := resolverLookUpAddr(resolverContext, sanitizedURL)
	if err != nil {
		// return false even for transient errors
		return false
	}

	return len(names) > 0
}

// for unit test mocking
var resolver = net.Resolver{PreferGo: true}
var LookUpAddr = resolver.LookupAddr
var LookUpNS = resolver.LookupNS

func resolverLookUpAddr(ctx context.Context, addr string) ([]string, error) {
	return LookUpAddr(ctx, addr)
}
func resolverLookupNS(ctx context.Context, name string) ([]*net.NS, error) {
	return LookUpNS(ctx, name)
}
