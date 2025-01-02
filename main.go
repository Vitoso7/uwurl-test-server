package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

type SitemapIndex struct {
	XMLName  xml.Name  `xml:"sitemapindex"`
	XMLNS    string    `xml:"xmlns,attr"`
	Sitemaps []Sitemap `xml:"sitemap"`
}

type Sitemap struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod"`
}

type URLSet struct {
	XMLName     xml.Name `xml:"urlset"`
	XMLNS       string   `xml:"xmlns,attr"`
	XMLNSNews   string   `xml:"xmlns:news,attr,omitempty"`
	XMLNSXhtml  string   `xml:"xmlns:xhtml,attr,omitempty"`
	XMLNSMobile string   `xml:"xmlns:mobile,attr,omitempty"`
	XMLNSImage  string   `xml:"xmlns:image,attr,omitempty"`
	XMLNSVideo  string   `xml:"xmlns:video,attr,omitempty"`
	URLs        []URL    `xml:"url"`
}

type URL struct {
	Loc        string      `xml:"loc"`
	LastMod    string      `xml:"lastmod,omitempty"`
	ChangeFreq string      `xml:"changefreq,omitempty"`
	Priority   string      `xml:"priority,omitempty"`
	XHTMLLinks []XHTMLLink `xml:"xhtml:link"`
}

type XHTMLLink struct {
	Rel      string `xml:"rel,attr"`
	Hreflang string `xml:"hreflang,attr"`
	Href     string `xml:"href,attr"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}

	serverPort := os.Getenv("SERVER_PORT")

	host := "localhost:" + serverPort

	if err := generateSitemaps("sitemaps", host); err != nil {
		fmt.Printf("Failed to generate sitemaps: %v\n", err)
		return
	}

	mux := http.NewServeMux()

	htmlFS := http.FileServer(http.Dir("pages"))
	sitemapFS := http.FileServer(http.Dir("sitemaps"))

	mux.Handle("/pages/", http.StripPrefix("/pages/", htmlFS))
	mux.Handle("/sitemaps/", http.StripPrefix("/sitemaps/", sitemapFS))

	server := &http.Server{
		Addr:    ":" + serverPort,
		Handler: mux,
	}

	fmt.Println("Server running at " + serverPort)

	server.ListenAndServe()
}

func generateSitemaps(outputDir, host string) error {
	sitemapData := map[string]map[string]interface{}{
		// Crunchyroll Mock
		"crunchyroll": {
			"sitemap.xml": SitemapIndex{
				XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
				Sitemaps: []Sitemap{
					{Loc: fmt.Sprintf("http://%s/sitemaps/crunchyroll/series/S.xml", host), LastMod: time.Now().Format("2006-01-02")},
				},
			},
			"series/S.xml": URLSet{
				XMLNS:      "http://www.sitemaps.org/schemas/sitemap/0.9",
				XMLNSXhtml: "http://www.w3.org/1999/xhtml",
				URLs: []URL{
					{
						Loc:        fmt.Sprintf("http://%s/pages/crunchyroll/series/spy_x_family.html", host),
						LastMod:    time.Now().Format("2006-01-02"),
						ChangeFreq: "daily",
					},
				},
			},
		},
	}

	// Iterate over domains and their files
	for domain, files := range sitemapData {
		domainDir := filepath.Join(outputDir, domain)
		for filename, content := range files {
			// Get the full file path
			filePath := filepath.Join(domainDir, filename)

			// Ensure the parent directory exists
			parentDir := filepath.Dir(filePath)
			if err := os.MkdirAll(parentDir, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", parentDir, err)
			}

			// Write the XML file
			if err := writeXMLFile(filePath, content); err != nil {
				return fmt.Errorf("failed to write file %s: %v", filePath, err)
			}
		}
	}

	return nil
}

// writeXMLFile encodes the content as XML and writes it to a file
func writeXMLFile(filePath string, content interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	if err := encoder.Encode(content); err != nil {
		return err
	}
	return nil
}
