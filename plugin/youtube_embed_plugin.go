// This file has been modified to support youtube-nocookie.com domain
// Original source: https://github.com/JohannesKaufmann/html-to-markdown/blob/master/plugin/iframe_youtube.go

package plugin

import (
	"fmt"
	"regexp"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
)

var youtubeID = regexp.MustCompile(`(?:youtube\.com|youtube-nocookie\.com)\/embed\/([^\&\?\/]+)`)

// YoutubeEmbed registers a rule (for iframes) and
// returns a markdown compatible representation (link to video, ...).
func YoutubeEmbed() md.Plugin {
	return func(c *md.Converter) []md.Rule {
		return []md.Rule{
			{
				Filter: []string{"iframe"},
				Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
					src := selec.AttrOr("src", "")
					if !strings.Contains(src, "youtube.com") && !strings.Contains(src, "youtube-nocookie.com") {
						fmt.Println("Not a YouTube iframe:", src)
						return nil
					}
					alt := selec.AttrOr("title", "")
					parts := youtubeID.FindStringSubmatch(src)
					if len(parts) != 2 {
						fmt.Println("YouTube ID not found in src:", src)
						return nil
					}
					id := parts[1]
					text := fmt.Sprintf("[![%s](https://img.youtube.com/vi/%s/0.jpg)](https://www.youtube.com/watch?v=%s)", alt, id, id)
					fmt.Println("YouTube video embedded:", text)
					return &text
				},
			},
		}
	}
}
