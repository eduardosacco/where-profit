package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
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
	Liquidity    float64 `json:",string"`
	SqrtPrice    string
	Token0Price  string
	Token1Price  string
	VolumeToken0 string
	VolumeToken1 string
	FeesUSD      float64 `json:",string"`
}

type poolStats struct {
	Id                     string
	AverageProfitPerDollar float64
	AverageAPR             float64
}

// Use https://www.epochconverter.com/ to get epoch
const startDate = 1640995200 //2022-01-01 12:00:00 AM
const endDate = 1646092799   //2022-02-28 11:59:59 PM
const subGraphUrl = "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v3"

func main() {
	t1 := time.Unix(startDate, 0)
	t2 := time.Unix(endDate, 0)
	days := int(t2.Sub(t1).Hours() / 24)

	fmt.Println("Calculating most liquid pool in a period of", strconv.Itoa(days), "days starting from", t1.UTC())

	pools := getLiquidityPools(t2.Unix())

	fmt.Println(pools[0]) // just testing

	// TODO: Add go routines
	var poolsStats []poolStats
	for _, p := range pools {
		pdd := getLiquidityPoolDaysData(p.Id, t1.Unix(), days)
		profitPerDollar := calculatePoolStats(pdd)
		poolsStats = append(poolsStats, poolStats{Id: p.Id, AverageProfitPerDollar: profitPerDollar})
	}

	sort.Slice(poolsStats, func(i, j int) bool {
		return poolsStats[i].AverageProfitPerDollar > poolsStats[j].AverageProfitPerDollar
	})

	fmt.Println(poolsStats)
	// pdd := getLiquidityPoolDaysData("0x1d42064fc4beb5f8aaf85f4617ae8b3b5b8bd801", t1.Unix(), days)
	// fmt.Println(pdd)

}

// Get all liquidity pools that were created before a limit date
func getLiquidityPools(limitDate int64) []pool {
	// excluding pools that were created after the limit date
	// to avoid processing data we don't need for the problem
	query := `query pools($skip:Int!, $limitDate: BigInt!) {
		pools(
			first: 1000
			skip: $skip
			where: {
				createdAtTimestamp_lt: $limitDate
		}) {
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

	var pools []pool
	var k = 0
	for {
		variables := map[string]any{
			"skip":      k * 1000,
			"limitDate": limitDate,
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

		if len(pr.Data.Pools) == 0 {
			break
		}

		pools = append(pools, pr.Data.Pools...)
		k++

		break // TODO: REMOVE, this is for testing purposes only
	}

	return pools
}

func getLiquidityPoolDaysData(pool string, date int64, days int) []poolDayDatas {
	// returns daily aggregated data for the first $days days
	// since the given &date timestamp for the $pool pool.
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
			feesUSD
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

	return pddr.Data.PoolDayDatas
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

// Will calculate only average values in the day span given
// TODO: Add APR calculation
func calculatePoolStats(pdds []poolDayDatas) float64 {
	var profitPerDollar float64
	for _, pdd := range pdds {
		dayProfitPerDollar := pdd.FeesUSD / float64(pdd.Liquidity)
		profitPerDollar = (profitPerDollar + dayProfitPerDollar) / 2
	}

	return profitPerDollar
}
