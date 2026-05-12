package wordreference

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Language struct {
	Name string
	Code string
}

var Languages = []Language{
	{Name: "English", Code: "en"},
	{Name: "Spanish", Code: "es"},
	{Name: "French", Code: "fr"},
}

type Result struct {
	Section         string
	Source          string
	SourceInfo      string
	Translation     string
	TranslationInfo string
	Notes           string
}

type Client struct {
	http *http.Client
}

func NewClient() Client {
	return Client{http: &http.Client{Timeout: 12 * time.Second}}
}

func (c Client) Lookup(ctx context.Context, origin, target Language, term string) ([]Result, error) {
	term = strings.TrimSpace(term)
	if term == "" {
		return nil, errors.New("enter a word or phrase to look up")
	}
	if origin.Code == target.Code {
		return nil, errors.New("origin and target languages must differ")
	}

	lookupURL := fmt.Sprintf("https://www.wordreference.com/%s%s/%s", origin.Code, target.Code, url.PathEscape(term))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, lookupURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "tuiference/0.1 (+https://wordreference.com)")

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wordreference returned %s", res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	results := parse(doc)
	if len(results) == 0 {
		return nil, errors.New("no table results found")
	}
	return results, nil
}

func parse(doc *goquery.Document) []Result {
	var results []Result
	currentSection := "Principal Translations"

	doc.Find("table.WRD tr").Each(func(_ int, row *goquery.Selection) {
		if heading := sectionHeading(row); heading != "" {
			currentSection = heading
			return
		}

		sourceCell := row.Find("td.FrWrd").First()
		translationCell := row.Find("td.ToWrd").First()
		if sourceCell.Length() == 0 || translationCell.Length() == 0 {
			return
		}

		source, sourceInfo := wordAndInfo(sourceCell)
		translation, translationInfo := wordAndInfo(translationCell)
		if source == "" || translation == "" {
			return
		}

		notes := collectNotes(row)

		results = append(results, Result{
			Section:         currentSection,
			Source:          source,
			SourceInfo:      sourceInfo,
			Translation:     translation,
			TranslationInfo: translationInfo,
			Notes:           notes,
		})
	})

	return dedupe(results)
}

func sectionHeading(row *goquery.Selection) string {
	if row.Find("td.FrWrd, td.ToWrd").Length() > 0 {
		return ""
	}

	text := clean(row.Text())
	if text == "" || len(text) > 90 {
		return ""
	}

	class, _ := row.Attr("class")
	if strings.Contains(class, "wrtopsection") || strings.Contains(class, "WRDheader") || strings.Contains(text, "Translations") || strings.Contains(text, "Forms") {
		return text
	}

	return ""
}

func wordAndInfo(cell *goquery.Selection) (string, string) {
	clone := cell.Clone()
	var info []string
	clone.Find("em, span.POS2, span.dsense, span.sense, span.roman, i").Each(func(_ int, s *goquery.Selection) {
		text := clean(s.Text())
		if text != "" {
			info = append(info, text)
		}
		s.Remove()
	})

	word := clean(clone.Text())
	return word, strings.Join(unique(info), ", ")
}

func collectNotes(row *goquery.Selection) string {
	var parts []string
	row.Find("td.POS2, td.sense, td.notePubl, td.note, em, span.dsense").Each(func(_ int, s *goquery.Selection) {
		text := clean(s.Text())
		if text != "" {
			parts = append(parts, text)
		}
	})
	return strings.Join(unique(parts), "; ")
}

func dedupe(results []Result) []Result {
	seen := map[string]bool{}
	out := make([]Result, 0, len(results))
	for _, result := range results {
		key := result.Section + "\x00" + result.Source + "\x00" + result.SourceInfo + "\x00" + result.Translation + "\x00" + result.TranslationInfo + "\x00" + result.Notes
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, result)
	}
	return out
}

func unique(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func clean(value string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(value, "⇒", "")), " ")
}
