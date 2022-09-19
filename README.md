# WHERE-PROFIT

## About

Found an online challenge for [Messari](https://engineering.messari.io/) a Web3 crypto research & data company
and thought it would be a good idea to attempt doing it to practice Go programming language which I am currently learning.

The purpose of the challenge is to calculate the most profitable Uniswap liquidity pool in a given time frame.
Data is obtained by querying Uniswap V3 subgraph.

This is actually my first Go script so the language best practices and conventions may not be top notch here.

## Resources I used on my investigation
* [Challenge](https://messari.notion.site/Messari-DeFi-Challenge-rev-03-17-2022-c5c6184e88dd44eab101be1f179a3ee0)
* [Uniswap V3 Subgraph explorer](https://thegraph.com/hosted-service/subgraph/uniswap/uniswap-v3)
* [Uniswap V3 Subgraph example queries](https://docs.uniswap.org/sdk/subgraph/subgraph-examples#pool-daily-aggregated)
* [Uniswap V3 Subgraph Schemas](https://github.com/Uniswap/v3-subgraph/blob/main/schema.graphql)
* [Uniswap V3 Whitepaper](https://uniswap.org/whitepaper-v3.pdf)
* [Epoch converter](https://www.epochconverter.com/)
* [APR calculation](https://www.investopedia.com/terms/a/apr.asp)

## How to use

### Variables
You can adjust the time frame to be computed by changing the `startDate` and `endDate` values in the main.go file.
These variables are expressed in epochs, use the epoch converter to get the epoch for specific dates.

The subgraph may or may not have indexed data for all days in the given time frame. To avoid having noisy computed values
derived from a small amount of data samples you can specify a minimum percentage of samples required for the pool stats to be computed.
This value can be modified via the `minimumDataSamplesPercentage` variable. By default 50% of the total days in the span
need to exist for the pool stats to be computed.

### Running the script
Make sure you have Go installed, here's a helpful link to do so https://go.dev/doc/install
From a terminal where the main.go file is located just run

`go run main.go`

### Output
The script will output the 10 pools with the highest avg per dollar profit in the given time frame.
Pool data output for each pool in order is:

id  - token0 / token1 symbols - avg per dollar profit - APR


## Observations
Querying for all the pool data in the given time frame from the subgraph takes some time. Be patient.
This could be improved in the future with parallel http calls using go routines.

More data than is actually needed to compute the profit and apr is queried from the subgraph, these fields were not removed from the graphql query
or the structures to have more data to experiment with. If speed was to be optimized, not querying for unneeded data may make the query more efficient.

