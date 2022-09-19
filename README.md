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
Make sure you have Go installed, more info [here](https://go.dev/doc/install).

From a terminal where the main.go file is located just run `go run main.go`.

### Output
The script will output the 10 pools with the highest avg per dollar profit in the given time frame.
Pool data output for each pool in order is:

id  - token0 / token1 symbols - avg per dollar profit - APR

Output example for the time frame values:
```
Calculating most liquid pool in a period of 58 days starting from 2022-01-01 00:00:00 +0000 UTC
Disregarding pool with less than 29 day samples in for the period

Top ten most profitable pools between 2022-01-01 00:00:00 +0000 UTC and 2022-02-28 23:59:59 +0000 UTC

0x5b6e17b4eb1e86b04f41d25c457ee5b9f3edef13       CRPT / WETH     98.28921682659382       89688.91035426685
0x9396c357befc79abfef7f229a3bd8dd0ae8e6bfd       SHPING / WETH   0.06003576762168322     37.7811296239903
0x75099758a9d1f43198043825c8fbcf8a12be7a74       sifu / USDT     0.046611128583921156    29.332865401950382
0x88b468740da532ea93e687d3c5bfda5efc26f2f8       CHEDDAR / WETH          0.04347581479067895     46.67256587822887
0x1d2e8efae9fab4731028d4a90f0cca27e1a57c9f       USDC / rUSD     0.04318214127204921     36.654608289065024
0xb71008f10b4b126c43fad95257aa29c6e3b8ca37       DOP / WETH      0.041620880432658144    34.526412177091416
0x8fec7a391cd9838935f4d4fd516ba6a3b3d2cda7       MDT / WETH      0.038285001377225805    39.92578715053548
0x63805e5d951398bc1c1bec242d303f59fa7732e3       X2Y2 / WETH     0.03450279129758539     21.712963488997705
0xc7ec0dfee680c9fd6586b00cf739fbc54e9563e4       HIT / WETH      0.03292945937453993     22.67783522963599
0xb23256f709c9c1152e038bc5e4fc19fe40ee48de       wPPC / WETH     0.0301729954480925      36.710477795179216
```

## Observations
Querying for all the pool data in the given time frame from the subgraph takes some time. Be patient.
This could be improved in the future with parallel http calls using go routines.

More data than is actually needed to compute the profit and apr is queried from the subgraph, these fields were not removed from the graphql query
or the structures to have more data to experiment with. If speed was to be optimized, not querying for unneeded data may make the query more efficient.
