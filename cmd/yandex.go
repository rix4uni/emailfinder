package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/net/html"
)

// fetchYandexEmails fetches emails from Yandex search results for the given domain
func fetchYandexEmails(domain string) []string {
	searchQuery := fmt.Sprintf("email \"@%s\"", domain)

	// Encode the search query
	encodedQuery := url.QueryEscape(searchQuery)

	// Construct the Yandex search URL
	searchURL := fmt.Sprintf("https://yandex.ru/search/?text=%s&ia=web&count=50&first=51", encodedQuery)

	// Create an HTTP client and set the User-Agent header
	client := &http.Client{}
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		fmt.Println("Error creating the request:", err)
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.131 Safari/537.36")

	// Make the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error fetching the URL:", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: status code", resp.StatusCode)
		return nil
	}

	// Parse the HTML
	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return nil
	}

	// Extract emails
	emails := extractyandexEmails(doc)

	// Process emails: split, lowercase, and filter
	return processyandexEmails(emails, domain)
}

// extractyandexEmails traverses the HTML nodes to find and collect email addresses
func extractyandexEmails(n *html.Node) []string {
	var emails []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			// Use regex to find email addresses in the text
			re := regexp.MustCompile(`([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`)
			matches := re.FindAllString(n.Data, -1)
			emails = append(emails, matches...)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	// Remove duplicates
	emailSet := make(map[string]struct{})
	for _, email := range emails {
		emailSet[email] = struct{}{}
	}

	var uniqueEmails []string
	for email := range emailSet {
		uniqueEmails = append(uniqueEmails, email)
	}

	return uniqueEmails
}

// processyandexEmails processes the extracted emails according to the specified steps
func processyandexEmails(emails []string, domain string) []string {
	var processed []string
	domainLower := strings.ToLower(domain) // To handle case insensitivity

	for _, email := range emails {
		emailLower := strings.ToLower(email) // Convert email to lowercase for comparison

		// Check if email ends with the specified domain
		if exactyandexMatch {
			// Match only if the email ends exactly with @domain
			if strings.HasSuffix(emailLower, "@"+domainLower) {
				processed = append(processed, emailLower)
			}
		} else {
			// Match if the email ends with the specified domain
			if strings.HasSuffix(emailLower, domainLower) {
				processed = append(processed, emailLower)
			}
		}
	}

	return processed
}

// yandexCmd represents the Yandex command
var yandexCmd = &cobra.Command{
	Use:   "yandex",
	Short: "Fetch emails from Yandex search results for a given domain.",
	Long: `This command allows you to fetch email addresses associated with a specified domain from Yandex search results.

Examples:
echo "domain.com" | emailfinder yandex
cat domains.txt | emailfinder yandex
cat domains.txt | emailfinder yandex -e`,
	Run: func(cmd *cobra.Command, args []string) {
		// Use a scanner to read domains from standard input
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			domain := strings.TrimSpace(scanner.Text())
			if domain == "" {
				continue
			}

			fmt.Printf("Fetching emails for domain: %s\n", domain)

			// Perform the email extraction for the domain
			emails := fetchYandexEmails(domain)

			if len(emails) == 0 {
				fmt.Printf("No emails found for domain: %s\n", domain)
			} else {
				for _, email := range emails {
					fmt.Println(email)
				}
				fmt.Printf("Found %d emails for: %s\n", len(emails), domain)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
		}
	},
}

var exactyandexMatch bool // Declare exactyandexMatch as a global variable

func init() {
	rootCmd.AddCommand(yandexCmd)

	yandexCmd.Flags().BoolVarP(&exactyandexMatch, "exact-match", "e", false, "Match emails exactly with the domain (e.g., @domain)")
}
