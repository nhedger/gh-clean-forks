package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-clean-forks",
	Short: "A CLI tool to clean up your GitHub forks",
	Long:  "A CLI tool to clean up your GitHub forks",
	Run: func(cmd *cobra.Command, args []string) {
		var runningDry, _ = cmd.Flags().GetBool("dry-run")
		if runningDry {
			color.Yellow("Running in dry-run mode. No repositories will be deleted.")
		}

		var force, _ = cmd.Flags().GetBool("force")
		if force {
			color.Red("Force mode enabled. Forks with open pull requests will be deleted.")
		}

		// Create clients
		token := cmd.Flags().Lookup("token").Value.String()
		graphqlClient, restClient, error := createClients(&token)
		if error != nil {
			log.Fatal(error)
		}

		username := cmd.Flags().Lookup("username").Value.String()
		if username == "" {
			username = getRunningUser(restClient)
		}
		fmt.Println("Running as: " + username)

		// Retrieve the user's forks
		allForks := getForks(username, graphqlClient)
		reposWithOpenPullRequests := getReposWithOpenPullRequests(username, graphqlClient)

		// Create an array of struct that contains the name of the fork and whether or not it has open pull requests
		type Fork struct {
			name       string
			hasOpenPRs bool
		}

		var forks []Fork = []Fork{}

		for _, fork := range allForks {
			var hasOpenPRs bool = false

			for _, repo := range reposWithOpenPullRequests {
				if fork == repo {
					hasOpenPRs = true
					break
				}
			}

			forks = append(forks, Fork{name: fork, hasOpenPRs: hasOpenPRs})
		}

		for _, fork := range forks {
			if force || !fork.hasOpenPRs {
				deleteFork(fork.name, runningDry, restClient)
				color.Green("[DELETED] " + fork.name)
			} else {
				color.Yellow("[SKIPPED] " + fork.name + " (has open pull requests)")
			}
		}
	},
}

func Execute() {
	if error := rootCmd.Execute(); error != nil {
		fmt.Println(error)
		os.Exit(1)
	}
}

func createClients(token *string) (*api.GraphQLClient, *api.RESTClient, error) {
	// Create the GraphQL client
	graphqlClient, error := api.NewGraphQLClient(api.ClientOptions{
		AuthToken: *token,
	})
	if error != nil {
		return nil, nil, error
	}

	// Create the REST client
	restClient, error := api.NewRESTClient(api.ClientOptions{
		AuthToken: *token,
	})
	if error != nil {
		return nil, nil, error
	}

	return graphqlClient, restClient, nil
}

func getForks(username string, client *api.GraphQLClient) []string {
	// Create a slice to store the user's forks
	var forks []string = []string{}

	// Retrieve the user's forks
	var forksQuery struct {
		User struct {
			Repositories struct {
				Nodes []struct {
					NameWithOwner string
				}
				PageInfo struct {
					HasNextPage bool
					EndCursor   string
				}
			} `graphql:"repositories(first: 50, after: $endCursor, affiliations: [OWNER], isFork: true)"`
		} `graphql:"user(login: $login)"`
	}

	variables := map[string]interface{}{
		"login":     graphql.String(username),
		"endCursor": (*graphql.String)(nil),
	}

	for {
		if error := client.Query("Forks", &forksQuery, variables); error != nil {
			log.Fatal(error)
		}

		for _, fork := range forksQuery.User.Repositories.Nodes {
			forks = append(forks, fork.NameWithOwner)
		}

		if !forksQuery.User.Repositories.PageInfo.HasNextPage {
			break
		}
		variables["endCursor"] = graphql.String(forksQuery.User.Repositories.PageInfo.EndCursor)
	}

	return forks
}

func getReposWithOpenPullRequests(username string, client *api.GraphQLClient) []string {
	var repositories []string = []string{}

	var pullRequestsQuery struct {
		User struct {
			PullRequests struct {
				Nodes []struct {
					HeadRef struct {
						Name       string
						Repository struct {
							NameWithOwner string
						}
					}
				}
				PageInfo struct {
					HasNextPage bool
					EndCursor   string
				}
			} `graphql:"pullRequests(states: [OPEN], first: 50, after: $endCursor)"`
		} `graphql:"user(login: $login)"`
	}

	variables := map[string]interface{}{
		"login":     graphql.String(username),
		"endCursor": (*graphql.String)(nil),
	}

	for {
		if error := client.Query("PullRequests", &pullRequestsQuery, variables); error != nil {
			log.Fatal(error)
		}

		for _, pullRequests := range pullRequestsQuery.User.PullRequests.Nodes {
			repositories = append(repositories, pullRequests.HeadRef.Repository.NameWithOwner)
		}

		if !pullRequestsQuery.User.PullRequests.PageInfo.HasNextPage {
			break
		}
		variables["endCursor"] = graphql.String(pullRequestsQuery.User.PullRequests.PageInfo.EndCursor)
	}

	return repositories
}

func deleteFork(name string, dryRun bool, client *api.RESTClient) {
	if !dryRun {
		error := client.Delete("repos/"+name, nil)
		if error != nil {
			log.Fatal(error)
		}
	}
}

func getRunningUser(client *api.RESTClient) string {
	response := struct{ Login string }{}
	error := client.Get("user", &response)
	if error != nil {
		fmt.Println(error)
	}
	return response.Login
}

func init() {
	rootCmd.Flags().StringP("username", "u", "", "GitHub username")
	rootCmd.Flags().StringP("token", "t", "", "GitHub personal access token")
	rootCmd.Flags().BoolP("dry-run", "d", false, "Dry run mode")
	rootCmd.Flags().BoolP("force", "f", false, "Force delete forks that have open pull requests")
}
