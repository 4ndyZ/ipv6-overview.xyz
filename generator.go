package main

import (
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    log "github.com/sirupsen/logrus"
    "sync"
    "strings"
    "html/template"
    "os"
    "time"
    "sort"

	"github.com/miekg/dns"
    yaml "gopkg.in/yaml.v2"
    "github.com/tdewolff/minify"
    "github.com/tdewolff/minify/html"
)

type IPv6_Support int

const (
    IPv6_Full_Support int = 1
    IPv6_Partial_Support int = 0
    IPv6_No_Support = -1

    RESOLVER_WORKER_GOROUTINE_COUNT int = 15
)

type YAMLConfig struct {
    Resolvers map[string][]string `yaml:"resolvers"`
    Categories []string `yaml:"categories"`
    Websites map[string]struct {
        Href string
        Hosts []string
        Icon string
        Twitter string
        Categories []string
    } `yaml:"websites"`
    WebsiteTitle string `yaml:"website_title"`
    GithubRepo string `yaml:"github_repo"`
}

type Resolver struct {
    Address string
    ResolverProvider *ResolverProvider
}

type ResolverProvider struct {
    Name string
    Resolvers []*Resolver
}

type Category struct {
    Name string
    Websites []*Website

    CountIPv6FullSupport uint
    CountIPv6PartialSupport uint
    CountIPv6NoSupport uint
}

func (category *Category) GetHTMLAnchor() string {
    return HTMLAnchorify(category.Name)
}

func (category *Category) DoTheCounting() {
    for _, website := range category.Websites {
        switch website.IPv6SupportStatus {
        case IPv6_Full_Support:
            category.CountIPv6FullSupport++

        case IPv6_Partial_Support:
            category.CountIPv6PartialSupport++

        default:
            category.CountIPv6NoSupport++
        }
    }
}

type Website struct {
    Name string
    URL string
    Domains []*Domain
    Icon string
    Twitter string
    Categories []string

    IPv6SupportStatus int
    CheckDurationInSeconds float64
}

func (website *Website) GetHTMLAnchor() string {
    return HTMLAnchorify(website.Name)
}

func (website *Website) GetCSSBackgroundColor() string {
    switch website.IPv6SupportStatus {
    case IPv6_Full_Support:
        return "test-result-full-ipv6"
    case IPv6_Partial_Support:
        return "test-result-partial-ipv6"
    default:
        return "test-result-no-ipv6"
    }
}

func (website *Website) IsFontAwesomeIcon() bool {
    return strings.HasPrefix(website.Icon, "fa-")
}

func (website *Website) GetSupportMessage() string {
    switch website.IPv6SupportStatus {
    case IPv6_Full_Support:
        return "Yay! Full IPv6 Support!"
    case IPv6_Partial_Support:
        return "You can do better!"
    default:
        return "Shame on you!"
    }
}

func (website *Website) GetBorderColor() string {
    switch website.IPv6SupportStatus {
    case IPv6_Full_Support:
        return "border-success"
    case IPv6_Partial_Support:
        return "border-warning"
    default:
        return "border-danger"
    }
}

func (website *Website) GetTwitterMessage() string {
    message := ""

    switch website.IPv6SupportStatus {
    case IPv6_Full_Support:
        message = "Thanks for serving your website over IPv6!"
    case IPv6_Partial_Support:
        message = "Can you please improve your IPv6 support?"
    default:
        message = "Isn't it about time to provide IPv6 on your website?"
    }

    return fmt.Sprintf(".%s %s #ipv6 #whyipv6", website.Twitter, message)
}

func (website *Website) FigureOutIPv6SupportStatus() {
    countIPv6Found := 0
    countIPv6NotFOund := 0

    for _, domain := range website.Domains {
        for _, results := range domain.ResolverResults {
            for _, result := range results.ResolverResults {
                if result.QuadAFound {
                    countIPv6Found++
                } else {
                    countIPv6NotFOund++
                }
            }
        }
    }

    if countIPv6Found > 0 && countIPv6NotFOund == 0 {
        website.IPv6SupportStatus = IPv6_Full_Support
    } else if countIPv6Found == 0 && countIPv6NotFOund > 0 {
        website.IPv6SupportStatus = IPv6_No_Support
    } else {
        website.IPv6SupportStatus = IPv6_Partial_Support
    }
}

type Domain struct {
    Domain string
    ResolverResults []DomainResolverResults
}

type DomainResolverResults struct {
    ResolverProvider *ResolverProvider
    ResolverResults []DomainResolverResult
}

type DomainResolverResult struct {
    Resolver *Resolver
    QuadAFound bool
    // Result string
}

type WebsiteTemplate struct {
    Categories []*Category
    Year int
    CreationTime string
    Title string
    GithubRepo string
}

func HTMLAnchorify(toAnchor string) string {
    replaceChars := []string { " ", ".", ",", "ä", "ö", "ü" }

    for _, char := range replaceChars {
        toAnchor = strings.Replace(toAnchor, char, "", -1)
    }

    return strings.ToLower(toAnchor)
}

func main() {
    minifyPage := flag.Bool("minify", false, "Minfiy page")
    flag.Parse()

    log.SetLevel(log.InfoLevel)

    yamlConfig := LoadYAML()

    resolverProviders := ParseResolverProviders(yamlConfig)
    categories := ParseCategories(yamlConfig)
    websites := ParseWebsites(yamlConfig)

    TestEveryWebsite(websites, resolverProviders)

    SortEveryWebsiteIntoCategory(websites, categories)
    GenerateCategoryCounters(categories)
    SortWebsitesInsideCategories(categories)

    renderedPage := RenderPage(yamlConfig, categories)

    if *minifyPage {
        renderedPage = MinifyPage(renderedPage)
    }

    WritePageToDisk(renderedPage)

    log.Info("Finished")
}

func WritePageToDisk(page string) {
    log.Info("Writing page to disk")

    file, fileerr := os.Create("dist/index.html")

    if fileerr != nil {
        log.WithField("ErrorMessage", fileerr.Error()).Fatal("Was not able to create/open dist/index.html")
    } else {
        io.WriteString(file, page)
        file.Close()

        log.Info("Wrote page to index.html")
    }
}

func MinifyPage(unminified string) string {
    log.Info("Minifying page")

    m := minify.New()
    m.AddFunc("text/html", html.Minify)

    minified, minifyerr := m.String("text/html", unminified)

    if minifyerr != nil {
        log.WithField("ErrorMessage", minifyerr.Error()).Fatal("Failed to minify page")
        return ""
    } else {
        log.Info("Minifying done")
        return minified
    }
}

func RenderPage(yamlConfig *YAMLConfig, categories []*Category) string {
    log.Info("Rendering page")

    websiteTemplate := &WebsiteTemplate{}
    websiteTemplate.Categories = categories
    t := time.Now()
    websiteTemplate.Year = t.Year()
    websiteTemplate.CreationTime = t.UTC().Format(time.RFC822)
    websiteTemplate.Title = yamlConfig.WebsiteTitle
    websiteTemplate.GithubRepo = yamlConfig.GithubRepo

    // funcMap := template.FuncMap{
	// 	"add": func (a,b int) int {
    //         return a + b
    //     },
    // }

    htmlTemplate := template.New("index.template")
    // htmlTemplate.Funcs(funcMap)
    if _, error := htmlTemplate.ParseFiles("index.template"); error != nil {
        log.WithField("ErrorMessage", error.Error()).Fatal("Failed to parse template")
        return ""
    } else {
        renderedPage := &strings.Builder{}

        if executeError := htmlTemplate.Execute(renderedPage, websiteTemplate); executeError != nil {
            log.WithField("ErrorMessage", executeError.Error()).Fatal("Failed to render template")
            return ""
        } else {
            log.Info("Rendering done")

            return renderedPage.String()
        }
    }
}

func SortEveryWebsiteIntoCategory(websites[]*Website, categories []*Category) {
    log.Info("Sorting website into categories")

    for _, website := range websites {
        wasSorted := false

        for _, websiteCategory := range website.Categories {

            for _, category := range categories {
                if websiteCategory == category.Name {
                    category.Websites = append(category.Websites, website)
                    wasSorted = true

                    log.WithFields(log.Fields{
                       "Website": website.Name,
                       "Category": websiteCategory }).Debug("Sorted website into category")

                    break
                }
            }
        }

        if ! wasSorted {
            log.WithField("Website", website.Name).Error("Website was not sorted into a category. Check the spelling!")
        }
    }

    log.Info("Finished sorting website into categories")
}

func SortCategories(categories []*Category) {
    log.Info("Sorting categories")

    sort.SliceStable(categories, func(i, j int) bool {
        if strings.Compare(categories[i].Name, categories[j].Name) == -1 {
            return true
        } else {
            return false
        }
    })

    log.Info("Sorting categories finished")
}

func SortWebsitesInsideCategories(categories []*Category) {
    log.Info("Sorting websites inside categories")

    for _, category := range categories {
        sort.SliceStable(category.Websites, func(i, j int) bool {
            if strings.Compare(category.Websites[i].Name, category.Websites[j].Name) == -1 {
                return true
            } else {
                return false
            }
        })
    }

    log.Info("Sorting websites inside categories finished")
}

func GenerateCategoryCounters(categories []*Category) {
    log.Info("Generating category overall counters")

    for _, category := range categories {
        category.DoTheCounting()
    }

    log.Info("Finished generating category overall counters")
}

func TestEveryWebsite(websites []*Website, resolverProviders []*ResolverProvider) {
    log.Info("Testing websites")

    channel := make(chan *Website, 2 * RESOLVER_WORKER_GOROUTINE_COUNT)

    waitForResolve := &sync.WaitGroup{}

    for i := 0; i < RESOLVER_WORKER_GOROUTINE_COUNT; i++ {
        waitForResolve.Add(1)
        go ResolverWorker(channel, resolverProviders, waitForResolve)
        log.WithField("WorkerID", i).Info("Started worker goroutine")
    }

    for _, website := range websites {
        channel <- website
    }

    close(channel)

    log.Info("Finished queuing up websites")

    waitForResolve.Wait()

    log.Info("Workers finished their tasks")
}

func ParseResolverProviders(yamlConfig *YAMLConfig) []*ResolverProvider {
    resolverProviders := make([]*ResolverProvider, 0)

    for name,addresses := range yamlConfig.Resolvers {
        resolverProvider := &ResolverProvider{}
        resolverProvider.Name = name
        resolverProvider.Resolvers = make([]*Resolver, 0)

        for _, address := range addresses {
            resolver := &Resolver{}

            if strings.Contains(address, ":") {
                resolver.Address = "[" + address + "]"
            } else {
                resolver.Address = address
            }

            resolver.ResolverProvider = resolverProvider

            resolverProvider.Resolvers = append(resolverProvider.Resolvers, resolver)
        }

        resolverProviders = append(resolverProviders, resolverProvider)
    }

    return resolverProviders
}

func ParseCategories(yamlConfig *YAMLConfig) []*Category {
    log.Info("Parsing categories from YAML config")

    categories := make([]*Category, 0)

    for _, categoryName := range yamlConfig.Categories {
        category := &Category{ Name: categoryName }
        categories = append(categories, category)

        log.WithField("CategoryName", categoryName).Debug("Parsed category")
    }

    log.WithField("count", len(categories)).Info("Finished parsing categories")

    SortCategories(categories)

    return categories
}

func ParseWebsites(yamlConfig *YAMLConfig) []*Website{
    log.Info("Parsing websites from YAML config")

    websites := make([]*Website, 0)

    for websiteName, websiteConfig := range yamlConfig.Websites {
        log.WithField("Website", websiteName).Debug("Found website")

        website := &Website{}
        website.Name = websiteName
        website.URL = websiteConfig.Href

        website.Domains = make([]*Domain, 0)

        for _, domain := range websiteConfig.Hosts {
            log.WithField("Domain", domain).WithField("Website", websiteName).Debug("Found Domain for website")

            domain := &Domain{ Domain: domain }
            website.Domains = append(website.Domains, domain)
        }

        website.Icon = websiteConfig.Icon
        website.Twitter = websiteConfig.Twitter
        website.Categories = websiteConfig.Categories

        if len(website.Categories) == 0 {
            website.Categories = append(website.Categories, "Uncategorized")
        }

        websites = append(websites, website)
    }

    log.WithField("count", len(websites)).Info("Finished parsing websites")

    return websites
}

func LoadYAML() (*YAMLConfig) {
    yamlContent, fileReadError := ioutil.ReadFile("config.yml")

    if fileReadError != nil {
        log.WithField("ErrorMessage", fileReadError.Error()).Fatal("Was not able to read config.yml")
        return nil
    }

    yamlconfig := &YAMLConfig{}

    if unmarshallError := yaml.Unmarshal([]byte(yamlContent), yamlconfig); unmarshallError != nil {
        log.WithField("ErrorMessage", unmarshallError.Error()).Fatal("Was not able to parse config.yaml")
        return nil
    } else {
        return yamlconfig
    }
}

func ResolverWorker(websites <-chan *Website, resolverProviders []*ResolverProvider, waitGroup *sync.WaitGroup) {

    client := new(dns.Client)

    for website := range websites {
        startTime := time.Now()

        for _, domain := range website.Domains {
            domain.ResolverResults = make([]DomainResolverResults, 0)

            for _, resolverProvider := range resolverProviders {
                domainResolverResults := DomainResolverResults{}
                domainResolverResults.ResolverProvider = resolverProvider
                domainResolverResults.ResolverResults = make([]DomainResolverResult, 0)

                for _, resolver := range resolverProvider.Resolvers {

                    domainResolverResult := DomainResolverResult{}
                    domainResolverResult.Resolver = resolver
                    domainResolverResult.QuadAFound = false

                    message := new(dns.Msg)
                	message.RecursionDesired = true
                	message.SetQuestion(domain.Domain + ".", dns.TypeAAAA)

                    answer, _, err := client.Exchange(message, resolver.Address + ":53")

                    if err == nil && answer.Rcode == dns.RcodeSuccess {
                        for _, record := range answer.Answer {
                            if _, ok := record.(*dns.AAAA); ok {
                                domainResolverResult.QuadAFound = true
                                break // one is enough
                            }
                        }
                	}

                    domainResolverResults.ResolverResults = append(domainResolverResults.ResolverResults,
                        domainResolverResult)

                    log.WithFields(log.Fields{
                        "Website": website.Name,
                        "Domain": domain.Domain,
                        "ResolverProvider": resolverProvider.Name,
                        "Resolver": resolver.Address,
                        "QuadAFound": domainResolverResult.QuadAFound }).Debug("Resolve result received")
                }

                domain.ResolverResults = append(domain.ResolverResults, domainResolverResults)
            }
        }

        website.CheckDurationInSeconds = time.Now().Sub(startTime).Seconds()

        website.FigureOutIPv6SupportStatus()
    }

    waitGroup.Done()
}
