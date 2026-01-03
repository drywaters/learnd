package enricher

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestEstimateReadingTimeSecondsPrefersArticle(t *testing.T) {
	t.Parallel()

	articleWords := readingWordsPerMinute
	otherWords := readingWordsPerMinute

	htmlInput := fmt.Sprintf(
		"<html><body><div>%s</div><article>%s</article></body></html>",
		buildWords(otherWords),
		buildWords(articleWords),
	)

	doc, err := html.Parse(strings.NewReader(htmlInput))
	if err != nil {
		t.Fatalf("failed to parse html: %v", err)
	}

	seconds, words := estimateReadingTimeSeconds(doc)
	expectedSeconds := ((articleWords + readingWordsPerMinute - 1) / readingWordsPerMinute) * 60

	if words != articleWords {
		t.Fatalf("word count = %d, want %d", words, articleWords)
	}
	if seconds != expectedSeconds {
		t.Fatalf("seconds = %d, want %d", seconds, expectedSeconds)
	}
}

func TestEstimateReadingTimeSecondsIgnoresScriptStyle(t *testing.T) {
	t.Parallel()

	paragraphWords := 10
	htmlInput := fmt.Sprintf(
		"<html><head><style>%s</style></head><body><script>%s</script><p>%s</p></body></html>",
		buildWords(50),
		buildWords(50),
		buildWords(paragraphWords),
	)

	doc, err := html.Parse(strings.NewReader(htmlInput))
	if err != nil {
		t.Fatalf("failed to parse html: %v", err)
	}

	seconds, words := estimateReadingTimeSeconds(doc)
	expectedSeconds := ((paragraphWords + readingWordsPerMinute - 1) / readingWordsPerMinute) * 60

	if words != paragraphWords {
		t.Fatalf("word count = %d, want %d", words, paragraphWords)
	}
	if seconds != expectedSeconds {
		t.Fatalf("seconds = %d, want %d", seconds, expectedSeconds)
	}
}

func TestEstimateReadingTimeSecondsRoundsUp(t *testing.T) {
	t.Parallel()

	wordsCount := readingWordsPerMinute + 1
	htmlInput := fmt.Sprintf("<html><body><main>%s</main></body></html>", buildWords(wordsCount))

	doc, err := html.Parse(strings.NewReader(htmlInput))
	if err != nil {
		t.Fatalf("failed to parse html: %v", err)
	}

	seconds, words := estimateReadingTimeSeconds(doc)
	expectedSeconds := 2 * 60

	if words != wordsCount {
		t.Fatalf("word count = %d, want %d", words, wordsCount)
	}
	if seconds != expectedSeconds {
		t.Fatalf("seconds = %d, want %d", seconds, expectedSeconds)
	}
}

func buildWords(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.TrimSpace(strings.Repeat("word ", count))
}
