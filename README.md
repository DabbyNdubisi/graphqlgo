# graphqlgo
A go lang graphql client

## Usage
```
import graphqlgo

client := graphqlgo.Client{#insert-url-here, &http.Client{}}

query := getGraphQLQueryFromFile(graphQLFile)

f := func(data *[]byte) (interface{}, error) {
		var result struct {
				Val string
		}

		err := json.Unmarshal(*data, &result)
		return result, err
	}
  
gqlrequest := graphqlgo.GraphQLRequest{query, map[string]string{"Key": "value"}, f}
result, err := client.Execute(query)
```
