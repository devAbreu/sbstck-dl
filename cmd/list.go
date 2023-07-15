package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List the posts of a Substack",
	Long:  `List all the posts of a Substack`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		parsedURL, err := parseURL(args[0])
		if err != nil {
			log.Fatal(err)
		}
		mainWebsite := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
		if verbose {
			fmt.Printf("Main website: %s\n", mainWebsite)
			fmt.Println("Getting all posts URLs...")
		}
		urls, err := extractor.GetAllPostsURLs(ctx, mainWebsite)
		if err != nil {
			log.Fatal(err)
		}
		if verbose {
			fmt.Printf("Found %d posts.\n", len(urls))
		}
		for _, url := range urls {
			fmt.Println(url)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}