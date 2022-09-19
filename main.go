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
	Id       string
	Name     string
	Symbol   string
	Decimals string
}

type poolDayData struct {
	Date         int
	Liquidity    float64 `json:",string"`
	SqrtPrice    string
	Token0Price  string
	Token1Price  string
	VolumeToken0 string
	VolumeToken1 string
	FeesUSD      float64 `json:",string"`
	TvlUSD       float64 `json:",string"`
}

type poolStats struct {
	Id                     string
	AverageProfitPerDollar float64
	APR                    float64
	DaysWithInfo           int
}

type pool struct {
	Id                  string
	TotalValueLockedUSD string
	VolumeUSD           string
	TxCount             string
	Token0              token
	Token1              token
	PoolDayData         []poolDayData
	Stats               poolStats
}

// Use https://www.epochconverter.com/ to get epoch
const startDate = 1640995200 //2022-01-01 12:00:00 AM
const endDate = 1646092799   //2022-02-28 11:59:59 PM

// Minimum pool day data samples to exist in the given period in order for the pool stats to be taken into account
const minimumDataSamplesPercentage = 0.5

// Uniswap Subgraph URL
const subGraphUrl = "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v3"

func main() {
	t1 := time.Unix(startDate, 0)
	t2 := time.Unix(endDate, 0)
	days := int(t2.Sub(t1).Hours() / 24)
	minimumSamples := int(minimumDataSamplesPercentage * float64(days))

	fmt.Println("Calculating most liquid pool in a period of", strconv.Itoa(days), "days starting from", t1.UTC())
	fmt.Println("Disregarding pool with less than", minimumSamples, "day samples in for the period")
	fmt.Println()

	pools := getLiquidityPoolsWithDaysData(endDate, startDate, days)

	for k := 0; k < len(pools); k++ {
		// Calculate stats only for pools with minimum specified samples of daily data
		if len(pools[k].PoolDayData) > minimumSamples {
			pools[k].Stats = calculatePoolStats(pools[k])
		}
	}

	// sort pools by per dollar profit descending
	sort.Slice(pools, func(i, j int) bool {
		return pools[i].Stats.AverageProfitPerDollar > pools[j].Stats.AverageProfitPerDollar
	})

	fmt.Println("Top ten most profitable pools between", t1.UTC(), "and", t2.UTC())
	fmt.Println()

	for i := 0; i < 10; i++ {
		fmt.Println(pools[i].Id, "\t", pools[i].Token0.Symbol, "/", pools[i].Token1.Symbol, "\t",
			pools[i].Stats.AverageProfitPerDollar, "\t", pools[i].Stats.APR)
	}
}

// Get all liquidity pools that were created before a limit date including
// daily aggregated data for the first $lpDayDataDays days
// since the given &lpDayDataStartDate timestamp.
func getLiquidityPoolsWithDaysData(lpCreationDateLimit int, lpDayDataStartDate int, lpDayDataDays int) []pool {
	// excluding pools that were created after the limit date
	// to avoid processing data we don't need for the problem
	query := `query pools($first: Int!, $skip:Int!, $lpCreationDateLimit: BigInt!, $lpDayDataStartDate: Int!, $lpDayDataDays: Int!) {
    pools(
			first: $first
			skip: $skip
			where: {
				createdAtTimestamp_lt: $lpCreationDateLimit
			}
		) {
			id
			totalValueLockedUSD
			volumeUSD
			txCount
			token0 {
				id
				name
				symbol
				decimals
			}
			token1 {
				id
				name
				symbol
				decimals
			}
			poolDayData(first: $lpDayDataDays, orderBy: date, where: {
				date_gte: $lpDayDataStartDate
			} ) {
				date
				liquidity
				sqrtPrice
				token0Price
				token1Price
				volumeToken0
				volumeToken1
				feesUSD
				tvlUSD
			}
		}
	}`

	var pools []pool
	var k = 0
	for {
		variables := map[string]any{
			"first":               1000,
			"skip":                k * 1000,
			"lpCreationDateLimit": lpCreationDateLimit,
			"lpDayDataStartDate":  lpDayDataStartDate,
			"lpDayDataDays":       lpDayDataDays,
		}

		b := gqlRequest{Query: query, Variables: variables}
		body, err := json.Marshal(b)

		if err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}

		responseData := queryUniswapSubGraph(body)
		var pr *poolsResponse
		err = json.Unmarshal(responseData, &pr)

		if err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}

		if len(pr.Data.Pools) == 0 {
			break
		}

		pools = append(pools, pr.Data.Pools...)
		k++
	}

	return pools
}

// Send a GraphQl query to the Uniswap subgraph and return the raw response
func queryUniswapSubGraph(body []byte) []byte {
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

// Calculate average daily profit per dollar and APR in the given days span
func calculatePoolStats(p pool) poolStats {
	totalDays := len(p.PoolDayData)
	var sumProfitPerDollar float64
	for _, pdd := range p.PoolDayData {
		dayProfitPerDollar := pdd.FeesUSD / pdd.TvlUSD
		sumProfitPerDollar += dayProfitPerDollar
	}

	avgProfitPerDollar := sumProfitPerDollar / float64(totalDays)

	// uniswap v3 does not have compound interest
	apr := (avgProfitPerDollar) / float64(totalDays) * 365 * 100

	return poolStats{Id: p.Id, AverageProfitPerDollar: avgProfitPerDollar, APR: apr, DaysWithInfo: totalDays}
}
