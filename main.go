package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"github.com/jomei/notionapi"
	"gopkg.in/yaml.v3"
)

// Configuration for the application
type Config struct {
	NotionAPIToken        string
	NotionBlogDatabaseID  string
	NotionDiaryDatabaseID string
	OutputDir             string
	DatabaseType          string // "blog" or "diary"
}

// Frontmatter for Astro templates
type Frontmatter struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description,omitempty"`
	PublishedAt string   `yaml:"publishedAt,omitempty"`
	UpdatedAt   string   `yaml:"updatedAt,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
	Draft       bool     `yaml:"draft,omitempty"`
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// extractRichText extracts plain text from rich text
func extractRichText(richText []notionapi.RichText) string {
	var text strings.Builder
	for _, rt := range richText {
		text.WriteString(rt.PlainText)
	}
	return text.String()
}

// generateFrontmatterYAML generates YAML frontmatter
func generateFrontmatterYAML(frontmatter Frontmatter) (string, error) {
	yamlBytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter to YAML: %w", err)
	}
	return string(yamlBytes), nil
}

// generateFilename generates a filename for the article
func generateFilename(page notionapi.Page) string {
	title := ""

	// Try to get title from properties
	if titleProp, ok := page.Properties["title"]; ok {
		if tp, ok := titleProp.(*notionapi.TitleProperty); ok && len(tp.Title) > 0 {
			title = tp.Title[0].PlainText
		}
	} else if titleProp, ok := page.Properties["Title"]; ok {
		if tp, ok := titleProp.(*notionapi.TitleProperty); ok && len(tp.Title) > 0 {
			title = tp.Title[0].PlainText
		}
	} else if titleProp, ok := page.Properties["Name"]; ok {
		if tp, ok := titleProp.(*notionapi.TitleProperty); ok && len(tp.Title) > 0 {
			title = tp.Title[0].PlainText
		}
	}

	// If no title found, use page ID
	if title == "" {
		title = page.ID.String()
	}

	// Sanitize title for filename
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	sanitized := reg.ReplaceAllString(strings.ToLower(title), "-")
	sanitized = strings.Trim(sanitized, "-")

	return sanitized + ".md"
}

func main() {
	// Define command-line flags
	dbType := flag.String("type", "blog", "Database type to process: 'blog' or 'diary'")
	flag.Parse()

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	} else {
		log.Println("Loaded environment variables from .env file")
	}

	// Get configuration from environment variables
	config := Config{
		NotionAPIToken:        getEnv("NOTION_API_TOKEN", ""),
		NotionBlogDatabaseID:  getEnv("NOTION_BLOG_DATABASE_ID", ""),
		NotionDiaryDatabaseID: getEnv("NOTION_DIARY_DATABASE_ID", ""),
		OutputDir:             getEnv("OUTPUT_DIR", "./content"),
		DatabaseType:          *dbType,
	}

	// Validate configuration
	if config.NotionAPIToken == "" {
		log.Fatal("NOTION_API_TOKEN environment variable is required")
	}

	// Validate database ID based on the selected type
	if config.DatabaseType == "blog" {
		if config.NotionBlogDatabaseID == "" {
			log.Fatal("NOTION_BLOG_DATABASE_ID environment variable is required for blog database")
		}
	} else if config.DatabaseType == "diary" {
		if config.NotionDiaryDatabaseID == "" {
			log.Fatal("NOTION_DIARY_DATABASE_ID environment variable is required for diary database")
		}
	} else {
		log.Fatalf("Invalid database type: %s. Must be 'blog' or 'diary'", config.DatabaseType)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Initialize Notion client
	client := notionapi.NewClient(notionapi.Token(config.NotionAPIToken))

	// Determine which database ID to use
	var databaseID string
	if config.DatabaseType == "blog" {
		databaseID = config.NotionBlogDatabaseID
		fmt.Println("Processing blog database...")
	} else {
		databaseID = config.NotionDiaryDatabaseID
		fmt.Println("Processing diary database...")
	}

	// Fetch database
	database, err := client.Database.Get(context.Background(), notionapi.DatabaseID(databaseID))
	if err != nil {
		log.Fatalf("Failed to get database: %v", err)
	}

	fmt.Printf("Found database: %s\n", database.Title[0].PlainText)

	// Query database for pages
	query := &notionapi.DatabaseQueryRequest{
		PageSize: 100,
	}

	resp, err := client.Database.Query(context.Background(), notionapi.DatabaseID(databaseID), query)
	if err != nil {
		log.Fatalf("Failed to query database: %v", err)
	}

	fmt.Printf("Found %d articles in Notion database\n", len(resp.Results))

	// Process each article
	for _, page := range resp.Results {
		// Extract title
		title := ""
		if titleProp, ok := page.Properties["title"]; ok {
			if tp, ok := titleProp.(*notionapi.TitleProperty); ok && len(tp.Title) > 0 {
				title = tp.Title[0].PlainText
			}
		} else if titleProp, ok := page.Properties["Title"]; ok {
			if tp, ok := titleProp.(*notionapi.TitleProperty); ok && len(tp.Title) > 0 {
				title = tp.Title[0].PlainText
			}
		} else if titleProp, ok := page.Properties["Name"]; ok {
			if tp, ok := titleProp.(*notionapi.TitleProperty); ok && len(tp.Title) > 0 {
				title = tp.Title[0].PlainText
			}
		}

		if title == "" {
			log.Printf("Skipping page %s: no title found", page.ID)
			continue
		}

		// Create frontmatter
		frontmatter := Frontmatter{
			Title:     title,
			UpdatedAt: page.LastEditedTime.Format("2006-01-02"),
		}

		// Generate frontmatter YAML
		frontmatterYAML, err := generateFrontmatterYAML(frontmatter)
		if err != nil {
			log.Printf("Failed to generate frontmatter for page %s: %v", page.ID, err)
			continue
		}

		// Create content with frontmatter
		content := fmt.Sprintf("---\n%s---\n\n# %s\n\nThis content was imported from Notion.", frontmatterYAML, title)

		// Save to file
		filename := generateFilename(page)
		outputPath := filepath.Join(config.OutputDir, filename)
		if err := ioutil.WriteFile(outputPath, []byte(content), 0644); err != nil {
			log.Printf("Failed to write article to file %s: %v", outputPath, err)
			continue
		}

		fmt.Printf("Successfully converted article: %s\n", outputPath)
	}

	fmt.Println("Conversion completed!")
}
