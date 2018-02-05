package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	graphiql "github.com/mnmtanish/go-graphiql"
	graphql "github.com/neelance/graphql-go"
)

var (
	schema *graphql.Schema
)

var page = []byte(`
	<!DOCTYPE html>
	<html>
		<head>
			<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.11/graphiql.css" />
			<script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/1.1.0/fetch.min.js"></script>
			<script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react.min.js"></script>
			<script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react-dom.min.js"></script>
			<script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.11/graphiql.js"></script>
		</head>
		<body style="width: 100%; height: 100%; margin: 0; overflow: hidden;">
			<div id="graphiql" style="height: 100vh;">Loading...</div>
			<script>
				function graphQLFetcher(graphQLParams) {
					return fetch("/graphql", {
						method: "post",
						body: JSON.stringify(graphQLParams),
						credentials: "include",
					}).then(function (response) {
						return response.text();
					}).then(function (responseBody) {
						try {
							return JSON.parse(responseBody);
						} catch (error) {
							return responseBody;
						}
					});
				}
				ReactDOM.render(
					React.createElement(GraphiQL, {fetcher: graphQLFetcher}),
					document.getElementById("graphiql")
				);
			</script>
		</body>
	</html>
	`)

func ServeGraphiQL(w http.ResponseWriter, r *http.Request) {
	w.Write(page)
}

type Resolver struct{}

func (r *Resolver) Echo(args struct{ Text string }) string {

	return args.Text
}

func init() {

	var Schema = `

schema {
	query: Query
	
}

type Query {
	echo(text: String!): String!
}

`

	schema = graphql.MustParseSchema(Schema, &Resolver{})
}

func main() {

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {

		sendError := func(err error) {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		}

		req := &graphiql.Request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			sendError(err)
			return
		}

		c := context.Background()
		variables := map[string]interface{}{}
		result := schema.Exec(c, req.Query, "", variables)
		if len(result.Errors) != 0 {
			sendError(result.Errors[0])
			return
		}

		if err := json.NewEncoder(w).Encode(result); err != nil {
			sendError(err)
		}

	})

	http.HandleFunc("/", ServeGraphiQL)

	fmt.Println("listening on port 8089!")
	http.ListenAndServe(":8089", nil)

}
