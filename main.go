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
	ID          string   `yaml:"id,omitempty"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description,omitempty"`
	PublishedAt string   `yaml:"publishedAt,omitempty"`
	UpdatedAt   string   `yaml:"updatedAt,omitempty"`
	Date        string   `yaml:"date,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
	Draft       bool     `yaml:"draft,omitempty"`
	Weather     string   `yaml:"weather,omitempty"`
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// extractRichText extracts text from rich text, preserving links
func extractRichText(richText []notionapi.RichText) string {
	var text strings.Builder
	for _, rt := range richText {
		// Check if this rich text has a link
		if rt.Href != "" {
			// Format as markdown link: [text](url)
			text.WriteString(fmt.Sprintf("[%s](%s)", rt.PlainText, rt.Href))
		} else {
			// Just add the plain text
			text.WriteString(rt.PlainText)
		}
	}
	return text.String()
}

// retrievePageContent retrieves the content of a Notion page and converts it to markdown
func retrievePageContent(client *notionapi.Client, pageID notionapi.ObjectID) (string, error) {
	// Get the children blocks of the page
	resp, err := client.Block.GetChildren(context.Background(), notionapi.BlockID(pageID), nil)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve page content: %v", err)
	}

	// Convert blocks to markdown
	var markdown strings.Builder
	for _, block := range resp.Results {
		// Process each block based on its type
		blockType := block.GetType()

		switch blockType {
		case "paragraph":
			if paragraph, ok := block.(*notionapi.ParagraphBlock); ok {
				text := extractRichText(paragraph.Paragraph.RichText)
				markdown.WriteString(text + "  \n\n")
			}
		case "heading_1":
			if heading, ok := block.(*notionapi.Heading1Block); ok {
				text := extractRichText(heading.Heading1.RichText)
				markdown.WriteString("# " + text + "  \n\n")
			}
		case "heading_2":
			if heading, ok := block.(*notionapi.Heading2Block); ok {
				text := extractRichText(heading.Heading2.RichText)
				markdown.WriteString("## " + text + "  \n\n")
			}
		case "heading_3":
			if heading, ok := block.(*notionapi.Heading3Block); ok {
				text := extractRichText(heading.Heading3.RichText)
				markdown.WriteString("### " + text + "  \n\n")
			}
		case "bulleted_list_item":
			if item, ok := block.(*notionapi.BulletedListItemBlock); ok {
				text := extractRichText(item.BulletedListItem.RichText)
				markdown.WriteString("- " + text + "  \n")
			}
		case "numbered_list_item":
			if item, ok := block.(*notionapi.NumberedListItemBlock); ok {
				text := extractRichText(item.NumberedListItem.RichText)
				markdown.WriteString("1. " + text + "  \n")
			}
		case "to_do":
			if todo, ok := block.(*notionapi.ToDoBlock); ok {
				text := extractRichText(todo.ToDo.RichText)
				if todo.ToDo.Checked {
					markdown.WriteString("- [x] " + text + "  \n")
				} else {
					markdown.WriteString("- [ ] " + text + "  \n")
				}
			}
		case "code":
			if code, ok := block.(*notionapi.CodeBlock); ok {
				text := extractRichText(code.Code.RichText)
				language := string(code.Code.Language)
				markdown.WriteString("```" + language + "  \n" + text + "  \n```  \n\n")
			}
		case "quote":
			if quote, ok := block.(*notionapi.QuoteBlock); ok {
				text := extractRichText(quote.Quote.RichText)
				markdown.WriteString("> " + text + "  \n\n")
			}
		case "divider":
			markdown.WriteString("---  \n\n")
		case "image":
			if image, ok := block.(*notionapi.ImageBlock); ok {
				if image.Image.Type == "external" {
					markdown.WriteString("![Image](" + image.Image.External.URL + ")  \n\n")
				} else if image.Image.Type == "file" {
					markdown.WriteString("![Image](" + image.Image.File.URL + ")  \n\n")
				}
			}
		}
	}

	return markdown.String(), nil
}

// generateFrontmatterYAML generates YAML frontmatter
func generateFrontmatterYAML(frontmatter Frontmatter) (string, error) {
	// Create a custom YAML representation
	var yamlBuilder strings.Builder

	// Add ID if present
	if frontmatter.ID != "" {
		yamlBuilder.WriteString(fmt.Sprintf("id: %s\n", frontmatter.ID))
	}

	// Add title
	yamlBuilder.WriteString(fmt.Sprintf("title: %s\n", frontmatter.Title))

	// Add description if present
	if frontmatter.Description != "" {
		yamlBuilder.WriteString(fmt.Sprintf("description: %s\n", frontmatter.Description))
	}

	// Add publishedAt if present
	if frontmatter.PublishedAt != "" {
		yamlBuilder.WriteString(fmt.Sprintf("publishedAt: %s\n", frontmatter.PublishedAt))
	}

	// Add date if present (without quotes)
	if frontmatter.Date != "" {
		yamlBuilder.WriteString(fmt.Sprintf("date: %s\n", frontmatter.Date))
	}

	// Add tags if present (in the format ["tag1", "tag2", "tag3"])
	if len(frontmatter.Tags) > 0 {
		yamlBuilder.WriteString("tags: [")
		for i, tag := range frontmatter.Tags {
			if i > 0 {
				yamlBuilder.WriteString(", ")
			}
			yamlBuilder.WriteString(fmt.Sprintf("\"%s\"", tag))
		}
		yamlBuilder.WriteString("]\n")
	}

	// Add draft if true
	if frontmatter.Draft {
		yamlBuilder.WriteString("draft: true\n")
	}

	// Add weather if present
	if frontmatter.Weather != "" {
		yamlBuilder.WriteString(fmt.Sprintf("weather: %s\n", frontmatter.Weather))
	}

	return yamlBuilder.String(), nil
}

// processEmptyLines processes the content to handle empty lines according to requirements:
// - Remove single empty lines between sentences
// - If there are multiple consecutive empty lines, keep just one
func processEmptyLines(content string) string {
	// Split content by newline
	lines := strings.Split(content, "\n")

	// Process lines
	var result []string
	emptyLineCount := 0

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			// This is an empty line
			emptyLineCount++

			// Skip single empty lines
			if emptyLineCount == 1 {
				// Keep the first empty line after frontmatter
				if i > 0 && strings.TrimSpace(lines[i-1]) == "---" {
					result = append(result, line)
				}
				// Otherwise, skip it
			} else if emptyLineCount == 2 {
				// For multiple consecutive empty lines, keep one
				result = append(result, line)
			}
			// Skip any additional empty lines
		} else {
			// This is a non-empty line
			result = append(result, line)
			emptyLineCount = 0
		}
	}

	// Join lines back together
	return strings.Join(result, "\n")
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

// processPage processes a single Notion page and saves it as a markdown file
func processPage(client *notionapi.Client, page notionapi.Page, config Config) {
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
	} else if titleProp, ok := page.Properties["titile"]; ok { // Handle typo in field name
		if tp, ok := titleProp.(*notionapi.TitleProperty); ok && len(tp.Title) > 0 {
			title = tp.Title[0].PlainText
		}
	}

	if title == "" {
		log.Printf("Skipping page %s: no title found", page.ID)
		return
	}

	// Create frontmatter with page ID as fallback
	frontmatter := Frontmatter{
		ID:    page.ID.String(),
		Title: title,
	}

	// Try to get ID from properties (use the ID column value from Notion)
	var idProp notionapi.Property
	var ok bool

	// Check for "ID" or "id" property
	if idProp, ok = page.Properties["ID"]; !ok {
		idProp, ok = page.Properties["id"]
	}

	if ok {
		// Convert the property to string and extract the last part (the actual ID value)
		idStr := fmt.Sprintf("%v", idProp)
		parts := strings.Split(idStr, " ")
		if len(parts) > 0 {
			// Get the last part and remove any closing brace
			lastPart := strings.TrimSuffix(parts[len(parts)-1], "}")
			frontmatter.ID = lastPart
		} else {
			frontmatter.ID = idStr
		}
	}

	// Extract tags if available
	if tagsProp, ok := page.Properties["tags"]; ok {
		if mp, ok := tagsProp.(*notionapi.MultiSelectProperty); ok {
			tags := make([]string, len(mp.MultiSelect))
			for i, tag := range mp.MultiSelect {
				tags[i] = tag.Name
			}
			frontmatter.Tags = tags
		}
	} else if tagsProp, ok := page.Properties["Tags"]; ok {
		if mp, ok := tagsProp.(*notionapi.MultiSelectProperty); ok {
			tags := make([]string, len(mp.MultiSelect))
			for i, tag := range mp.MultiSelect {
				tags[i] = tag.Name
			}
			frontmatter.Tags = tags
		}
	}

	// For diary entries, extract description and weather
	if config.DatabaseType == "diary" {
		// Extract description
		if descProp, ok := page.Properties["description"]; ok {
			if rtp, ok := descProp.(*notionapi.RichTextProperty); ok && len(rtp.RichText) > 0 {
				frontmatter.Description = rtp.RichText[0].PlainText
			}
		}

		// Extract weather
		if weatherProp, ok := page.Properties["weather"]; ok {
			if rtp, ok := weatherProp.(*notionapi.RichTextProperty); ok && len(rtp.RichText) > 0 {
				frontmatter.Weather = rtp.RichText[0].PlainText
			}
		}
	}

	// Use CreatedTime as the date
	frontmatter.Date = page.CreatedTime.Format("2006-01-02")

	// Retrieve page content
	pageContent, err := retrievePageContent(client, page.ID)
	if err != nil {
		log.Printf("Failed to retrieve content for page %s: %v", page.ID, err)
		// If we can't retrieve the content, use a placeholder
		pageContent = "This content was imported from Notion, but the content could not be retrieved."
	}

	// For blog entries, set description as first 70 characters of content with newlines converted to spaces
	if config.DatabaseType == "blog" && pageContent != "" {
		// Replace newlines with spaces
		descriptionText := strings.ReplaceAll(pageContent, "\n", " ")
		// Remove extra spaces
		descriptionText = regexp.MustCompile(`\s+`).ReplaceAllString(descriptionText, " ")
		// Trim spaces
		descriptionText = strings.TrimSpace(descriptionText)
		// Get first 70 characters or less if content is shorter
		// Use runes to correctly handle multi-byte characters like Japanese
		runes := []rune(descriptionText)
		if len(runes) > 70 {
			frontmatter.Description = string(runes[:70]) + "..."
		} else {
			frontmatter.Description = descriptionText
		}
	} else if config.DatabaseType == "blog" {
		log.Printf("Not setting description for blog entry: %s (empty content)", title)
	}

	// Generate frontmatter YAML
	frontmatterYAML, err := generateFrontmatterYAML(frontmatter)
	if err != nil {
		log.Printf("Failed to generate frontmatter for page %s: %v", page.ID, err)
		return
	}

	// Create content with frontmatter
	content := fmt.Sprintf("---\n%s---\n\n%s", frontmatterYAML, pageContent)

	// Process empty lines: remove single empty lines, but keep one if there are multiple consecutive empty lines
	content = processEmptyLines(content)

	// Save to file
	filename := generateFilename(page)
	outputPath := filepath.Join(config.OutputDir, filename)
	if err := ioutil.WriteFile(outputPath, []byte(content), 0644); err != nil {
		log.Printf("Failed to write article to file %s: %v", outputPath, err)
		return
	}

	fmt.Printf("Successfully converted article: %s\n", outputPath)
}

// fetchDatabase initializes the Notion client, fetches the database, and queries it for pages
func fetchDatabase(config Config) (*notionapi.Client, []notionapi.Page) {
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

	return client, resp.Results
}

// loadConfig loads and validates the application configuration
func loadConfig() Config {
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

	return config
}

func main() {
	// Load and validate configuration
	config := loadConfig()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Fetch database and pages
	client, pages := fetchDatabase(config)

	// Process each article
	for _, page := range pages {
		processPage(client, page, config)
	}

	fmt.Println("Conversion completed!")
}
