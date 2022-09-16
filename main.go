package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type gqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type poolsResponse struct {
	Data poolsData `json:"data"`
}

type poolsData struct {
	Pools []pool
}

type token struct {
	Name string
}

type pool struct {
	Id                  string
	TotalValueLockedUSD string
	VolumeUSD           string
	TxCount             string
	Token0              token
	Token1              token
}

type poolDayDatasResponse struct {
	Data poolDayDatasData `json:"data"`
}

type poolDayDatasData struct {
	PoolDayDatas []poolDayDatas
}

type poolDayDatas struct {
	Date         int
	Liquidity    string
	SqrtPrice    string
	Token0Price  string
	Token1Price  string
	VolumeToken0 string
	VolumeToken1 string
}

const startDate = "2022-01-01"
const endDate = "2022-02-28"

const subGraphUrl = "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v3"

func main() {
	const layout = "2006-01-02"
	t1, _ := time.Parse(layout, startDate)
	t2, _ := time.Parse(layout, endDate)
	days := int(t2.Sub(t1).Hours() / 24)

	fmt.Println("Calculating most liquid pool in a period of " + strconv.Itoa(days) + " days starting from " + startDate)

	pr := getLiquidityPools() // currently only using the 1000 most liquid pools
	pools := pr.Data.Pools
	// for _, p := range pools {

	// }

	p0Id := pools[0].Id
	fmt.Println(p0Id)
	pddr := getLiquidityPoolDaysData("0x1d42064fc4beb5f8aaf85f4617ae8b3b5b8bd801", t1.Unix(), days)

	fmt.Println(pddr.Data.PoolDayDatas[0])

}

func getLiquidityPools() poolsResponse {
	// refactor in order for pagination to be more performant than using skip
	//https://thegraph.com/docs/en/querying/graphql-api/#example-using-and-2
	query := `query pools($skip:Int!) {
		pools(
			first: 1000
			skip: $skip
		) {
			id
			totalValueLockedUSD
			volumeUSD
			txCount
			token0 {
				name
			}
			token1 {
				name
			}
		}
	}`

	variables := map[string]any{
		"skip": 1000,
	}

	b := gqlRequest{Query: query, Variables: variables}
	body, err := json.Marshal(b)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData := queryTheGraph(body)

	var pr *poolsResponse
	err = json.Unmarshal(responseData, &pr)

	return *pr
}

func getLiquidityPoolDaysData(pool string, date int64, days int) poolDayDatasResponse {
	query := `query poolDayDatas($pool: String!, $date: Int!, $days: Int!) {
		poolDayDatas(first: $days, orderBy: date, where: {
			pool: $pool,
			date_gt: $date
		} ) {
			date
			liquidity
			sqrtPrice
			token0Price
			token1Price
			volumeToken0
			volumeToken1
		}
	}`

	variables := map[string]any{
		"pool": pool,
		"date": date,
		"days": days,
	}

	b := gqlRequest{Query: query, Variables: variables}
	body, err := json.Marshal(b)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData := queryTheGraph(body)

	var pddr *poolDayDatasResponse
	err = json.Unmarshal(responseData, &pddr)

	return *pddr
}

func queryTheGraph(body []byte) []byte {
	response, err := http.Post(subGraphUrl, "application/json", bytes.NewBuffer(body))

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	return responseData
}
