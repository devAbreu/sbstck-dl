package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexferrari88/sbstck-dl/lib"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// downloadCmd represents the download command
// downloadCmd represents the download command
var (
	downloadUrl  string
	format       string
	outputFolder string
	dryRun       bool
	force        bool
	logFile      string
	downloadCmd  = &cobra.Command{
		Use:   "download",
		Short: "Download individual posts or the entire public archive",
		Long:  `You can provide the url of a single post or the main url of the Substack you want to download.`,
		Run: func(cmd *cobra.Command, args []string) {
			startTime := time.Now()

			extractor, err := lib.NewExtractor(fetcher, logFile)
			if err != nil {
				log.Fatalf("Failed to create extractor: %v", err)
			}

			if strings.Contains(downloadUrl, "/p/") {
				if verbose {
					fmt.Printf("Downloading post %s\n", downloadUrl)
				}
				if dryRun {
					fmt.Println("Dry run, exiting...")
					return
				}
				if (beforeDate != "" || afterDate != "") && verbose {
					fmt.Println("Warning: --before and --after flags are ignored when downloading a single post")
				}

				post, err := extractor.ExtractPost(ctx, downloadUrl, outputFolder, force)
				if err != nil {
					log.Fatalln(err)
				}
				if post.Slug == "" {
					fmt.Println("No post was downloaded. Skipping...")
					return
				}

				downloadTime := time.Since(startTime)
				if verbose {
					fmt.Printf("Downloaded post %s in %s\n", downloadUrl, downloadTime)
				}

				postFolder := filepath.Join(outputFolder, post.Slug)
				path := filepath.Join(postFolder, fmt.Sprintf("%s.%s", post.Slug, format))
				if verbose {
					fmt.Printf("Writing post to file %s\n", path)
				}

				err = post.WriteToFile(path, format)
				if err != nil {
					log.Fatalln(err)
				}

				if verbose {
					fmt.Println("Done in ", time.Since(startTime))
				}
			} else {
				var downloadedPostsCount int
				dateFilterfunc := lib.MakeDateFilterFunc(beforeDate, afterDate)
				urls, err := extractor.GetAllPostsURLs(ctx, downloadUrl, dateFilterfunc)
				urlsCount := len(urls)
				if err != nil {
					log.Fatalln(err)
				}
				if urlsCount == 0 {
					if verbose {
						fmt.Println("No posts found, exiting...")
					}
					return
				}
				if verbose {
					fmt.Printf("Found %d posts\n", urlsCount)
				}
				if dryRun {
					fmt.Printf("Found %d posts\n", urlsCount)
					fmt.Println("Dry run, exiting...")
					return
				}
				bar := progressbar.NewOptions(len(urls),
					progressbar.OptionSetWidth(25),
					progressbar.OptionSetDescription("downloading"),
					progressbar.OptionShowBytes(true))
				for result := range extractor.ExtractAllPosts(ctx, urls, outputFolder, force) {
					select {
					case <-ctx.Done():
						log.Fatalln("context cancelled")
					default:
					}
					if result.Err != nil {
						if verbose {
							fmt.Printf("Error downloading post %s: %s\n", result.Post.CanonicalUrl, result.Err)
							fmt.Println("Skipping...")
						}
						continue
					}
					if result.Post.Slug == "" {
						continue
					}
					bar.Add(1)
					downloadedPostsCount++
					if verbose {
						fmt.Printf("Downloading post %s\n", result.Post.CanonicalUrl)
					}
					post := result.Post

					postFolder := filepath.Join(outputFolder, post.Slug)
					path := filepath.Join(postFolder, fmt.Sprintf("%s.%s", post.Slug, format))
					if verbose {
						fmt.Printf("Writing post to file %s\n", path)
					}

					err = post.WriteToFile(path, format)
					if err != nil {
						log.Fatalln(err)
					}
				}
				if verbose {
					fmt.Println("Downloaded", downloadedPostsCount, "posts, out of", len(urls))
					fmt.Println("Done in ", time.Since(startTime))
				}
			}
		},
	}
)

func init() {
	downloadCmd.Flags().StringVarP(&downloadUrl, "url", "u", "", "Specify the Substack url")
	downloadCmd.Flags().StringVarP(&format, "format", "f", "html", "Specify the output format (options: \"html\", \"md\", \"txt\"")
	downloadCmd.Flags().StringVarP(&outputFolder, "output", "o", ".", "Specify the download directory")
	downloadCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Enable dry run")
	downloadCmd.Flags().BoolVarP(&force, "force", "", false, "Force re-download of posts")
	downloadCmd.Flags().StringVar(&logFile, "log-file", "downloaded_posts.log", "Specify the log file to track downloaded posts")
	downloadCmd.MarkFlagRequired("url")
}

func convertDateTime(datetime string) string {
	// Parse the datetime string
	parsedTime, err := time.Parse(time.RFC3339, datetime)
	if err != nil {
		// Return an empty string or an error message if parsing fails
		return ""
	}

	// Format the datetime to the desired format
	formattedDateTime := fmt.Sprintf("%d%02d%02d_%02d%02d%02d",
		parsedTime.Year(), parsedTime.Month(), parsedTime.Day(),
		parsedTime.Hour(), parsedTime.Minute(), parsedTime.Second())

	return formattedDateTime
}

func parseURL(toTest string) (*url.URL, error) {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, err
	}

	return u, err
}

func makePath(post lib.Post, outputFolder string, format string) string {
	return fmt.Sprintf("%s/%s_%s.%s", outputFolder, convertDateTime(post.PostDate), post.Slug, format)
}

// extractSlug extracts the slug from a Substack post URL
// e.g. https://example.substack.com/p/this-is-the-post-title -> this-is-the-post-title
func extractSlug(url string) string {
	split := strings.Split(url, "/")
	return split[len(split)-1]
}

// filterExistingPosts filters out posts that already exist in the output folder.
// It looks for files whose name ends with the post slug.
func filterExistingPosts(urls []string, outputFolder string, format string) ([]string, error) {
	var filtered []string
	for _, url := range urls {
		slug := extractSlug(url)
		path := fmt.Sprintf("%s/%s_%s.%s", outputFolder, "*", slug, format)
		matches, err := filepath.Glob(path)
		if err != nil {
			return urls, err
		}
		if len(matches) == 0 {
			filtered = append(filtered, url)
		}
	}
	return filtered, nil
}

// Añadir lógica para descargar archivos multimedia
func downloadMedia(urls []string, outputFolder string) error {
	for _, mediaUrl := range urls {
		resp, err := http.Get(mediaUrl)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		fileName := filepath.Base(mediaUrl)
		outputPath := filepath.Join(outputFolder, fileName)
		out, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return err
		}
	}
	return nil
}
