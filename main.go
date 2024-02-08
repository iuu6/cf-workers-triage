package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"regexp"
)

// 结构体定义用于解析GraphQL查询结果的JSON
type GraphQLResponse struct {
	Data struct {
		Viewer struct {
			Zones []struct {
				HTTPRequests1DGroups []struct {
					Date struct {
						Date string `json:"date"`
					} `json:"dimensions"`
					Sum struct {
						Bytes    int `json:"bytes"`
						Requests int `json:"requests"`
					} `json:"sum"`
				} `json:"httpRequests1dGroups"`
			} `json:"zones"`
		} `json:"viewer"`
	} `json:"data"`
}

// User 结构体用于保存用户信息
type User struct {
	Email string `json:"email"`
	Key   string `json:"key"`
}

// ResultData 结构体用于保存每次的数据
type ResultData struct {
	Pattern  string `json:"pattern"`
	Bytes    int    `json:"bytes"`
	Requests int    `json:"requests"`
}

func readConfig(filename string) ([]User, error) {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var users []User
	err = json.Unmarshal(fileContent, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// 获取账户信息
func listAccounts(email, key string) (*http.Response, error) {
	url := "https://api.cloudflare.com/client/v4/accounts"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Email", email)
	req.Header.Set("X-Auth-Key", key)

	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

// 获取域名信息
func listZones(email, key string) (*http.Response, error) {
	url := "https://api.cloudflare.com/client/v4/zones"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Email", email)
	req.Header.Set("X-Auth-Key", key)

	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

// 获取Workers路由（获取节点域名）
func listFilters(email, key, zoneID string) (*http.Response, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/workers/filters", zoneID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Email", email)
	req.Header.Set("X-Auth-Key", key)

	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

// GraphQL API调用
func graphqlAPI(email, key, zoneTag, start, end string) (*http.Response, error) {
	url := "https://api.cloudflare.com/client/v4/graphql"
	query := `
	{
		viewer {
			zones(filter: { zoneTag: $tag }) {
				httpRequests1dGroups(
					orderBy: [date_ASC]
					limit: 1000
					filter: { date_gt: $start, date_lt: $end }
				) {
					date: dimensions {
						date
					}
					sum {
						bytes
						requests
					}
				}
			}
		}
	}`
	variables := map[string]string{"tag": zoneTag, "start": start, "end": end}
	requestBody, err := json.Marshal(map[string]interface{}{"query": query, "variables": variables})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Email", email)
	req.Header.Set("X-Auth-Key", key)

	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

// 将当前日期移位n天，并返回移位日期和移位日期的前一个和下一个日期。
func dateShift(n int) (string, string, string) {
	today := time.Now().Local().AddDate(0, 0, n)
	shiftedDate := today.AddDate(0, 0, n)
	previousDate := shiftedDate.AddDate(0, 0, -1)
	nextDate := shiftedDate.AddDate(0, 0, 1)

	return shiftedDate.Format("2006-01-02"), previousDate.Format("2006-01-02"), nextDate.Format("2006-01-02")
}

var allUserData []ResultData  // 添加一个全局数组，用于保存所有用户的数据

func processAndSaveData(pattern, graphqlBody string, bytes, requests int) {
    // 提取pattern网址
	fmt.Println("pattern", pattern)

    // 创建 ResultData 结构体
    resultData := ResultData{
        Pattern:  pattern,
        Bytes:    bytes,
        Requests: requests,
    }

    // 将数据添加到全局数组中
    allUserData = append(allUserData, resultData)
}

func runForUser(user User) {
	email := user.Email
	key := user.Key
	fmt.Println("Processing for User:", email)

	var zoneTag string
	var domain string

	// 获取账户信息
	accountsResponse, err := listAccounts(email, key)
	if err != nil {
		fmt.Println("Error fetching accounts:", err)
		return
	}
	defer accountsResponse.Body.Close()

	// 输出账户信息
	accountsBody, _ := ioutil.ReadAll(accountsResponse.Body)
	fmt.Println("Accounts Response:", string(accountsBody))

	// 获取域名信息
	zonesResponse, err := listZones(email, key)
	if err != nil {
		fmt.Println("Error fetching zones:", err)
		return
	}
	defer zonesResponse.Body.Close()

	// 输出域名信息
	zonesBody, _ := ioutil.ReadAll(zonesResponse.Body)
	fmt.Println("Zones Response:", string(zonesBody))

	// 解析域名信息
	var zonesData map[string]interface{}
	if err := json.Unmarshal(zonesBody, &zonesData); err != nil {
    	fmt.Println("Error parsing Zones response:", err)
    	return
	}

	// 获取域名列表
	resultInfo, ok := zonesData["result_info"].(map[string]interface{})
	if !ok {
    	fmt.Println("Error getting result_info from response")
    	return
	}

	totalCount, ok := resultInfo["total_count"].(float64)
	if !ok {
    	fmt.Println("Error getting total_count from result_info")
    	return
	}

	zonesList, ok := zonesData["result"].([]interface{})
	if !ok {
    	fmt.Println("Error getting zones list from response")
    	return
	}
	
	// 创建一个保存所有zoneID的数组
	var allZoneIDs []string

	// 遍历所有域名，并提取ID
	for i := 0; i < int(totalCount); i++ {
    	zoneInfo, ok := zonesList[i].(map[string]interface{})
    	if !ok {
    	    fmt.Println("Error getting zone info")
    	    continue
    	}

    	// 获取当前域名的ID
    	zoneID, ok := zoneInfo["id"].(string)
    	if !ok {
    	    fmt.Println("Error getting zone ID")
    	    continue
    	}

    	// 使用当前域名的ID进行后续操作，例如打印或存储
    	fmt.Println("Zone ID:", zoneID)

	    // 将当前域名的ID添加到数组中
		allZoneIDs = append(allZoneIDs, zoneID)
		fmt.Println("allZoneIDs:", allZoneIDs)
	}

	// 循环遍历每个域名的ID
	for _, zoneID := range allZoneIDs {
		// 使用当前域名的ID调用listFilters函数
		filtersResponse, err := listFilters(email, key, zoneID)
		if err != nil {
			fmt.Println("Error fetching filters for zone ID", zoneID, ":", err)
			continue
		}
		defer filtersResponse.Body.Close()
	
		// 解析Filters响应
		var filtersData map[string]interface{}
		filtersBody, _ := ioutil.ReadAll(filtersResponse.Body)
		if err := json.Unmarshal(filtersBody, &filtersData); err != nil {
			fmt.Println("Error parsing Filters response for zone ID", zoneID, ":", err)
			continue
		}
	
		// 检查是否存在非空的result字段
		if result, ok := filtersData["result"].([]interface{}); ok && len(result) > 0 {
			// 获取第一个非空result字段的zoneID，并赋值给zoneTag
			if matchedZoneID, ok := result[0].(map[string]interface{})["id"].(string); ok {
				zoneTag = zoneID
				pattern := fmt.Sprintf("%v", result)
				// 输出符合条件的zoneID
				fmt.Println("Matching Zone ID:", matchedZoneID)
				fmt.Println("pattern:", pattern)
				re := regexp.MustCompile(`pattern:([^/\s]+)/*`)
				match := re.FindStringSubmatch(pattern)
				if len(match) > 1 {
					domain = match[1]
					fmt.Println(domain)
				} else {
					fmt.Println("未找到匹配项")
				}
				break
			}
		}
	}
	
	// 输出最终选择的zoneTag
	fmt.Println("Final Zone Tag:", zoneTag)
	fmt.Println("Final domain:", domain)

	// 获取当前日期和前后两天的日期
	_, start, end := dateShift(0)

	// 获取域名数据统计
	graphqlResponse, err := graphqlAPI(email, key, zoneTag, start, end)
	fmt.Println("Time:", start, end)
	if err != nil {
		fmt.Println("Error fetching GraphQL data:", err)
		return
	}
	defer graphqlResponse.Body.Close()

	var responseData GraphQLResponse
	graphqlBody, _ := ioutil.ReadAll(graphqlResponse.Body)
	err = json.Unmarshal(graphqlBody, &responseData)
	if err != nil {
		fmt.Println("Error parsing GraphQL response:", err)
		return
	}
	fmt.Println("GraphQL Response:", responseData)

    // 提取 pattern、bytes 和 requests
	for _, zone := range responseData.Data.Viewer.Zones {
    	// 检查是否存在 HTTPRequests1DGroups
    	if len(zone.HTTPRequests1DGroups) > 0 {
    	    for _, group := range zone.HTTPRequests1DGroups {
    	        // 直接引用 group 中的字段
    	        pattern := domain
    	        bytes := group.Sum.Bytes
    	        requests := group.Sum.Requests

            	// 处理数据并保存到 JSON 文件
            	processAndSaveData(pattern, string(graphqlBody), bytes, requests)
        	}
	    } else {
	        // 如果数据为空，将 bytes 和 requests 的值设置为 0
	        pattern := domain
	        bytes := 0
	        requests := 0

        	// 处理数据并保存到 JSON 文件
        	processAndSaveData(pattern, string(graphqlBody), bytes, requests)
    	}
	}
}

// 处理所有用户的数据并保存到文件
func processAndSaveAllUserData() error {
    // 获取当前日期并用作文件名
    fileName := fmt.Sprintf("data.json")

    // 将所有用户的数据转换为 JSON 格式
    jsonData, err := json.Marshal(allUserData)
    if err != nil {
        return err
    }

    // 将数据保存到文件
    err = ioutil.WriteFile(fileName, jsonData, 0644)
    if err != nil {
        return err
    }

    fmt.Printf("All user data saved to %s\n", fileName)
    return nil
}

func main() {
    // Read users from the config file
    users, err := readConfig("config.json")
    if err != nil {
        fmt.Println("Error reading config file:", err)
        return
    }

    // Loop through each user
    for _, user := range users {
        // Perform operations for each user
        runForUser(user)
    }

    // 处理所有用户的数据并保存到文件
    err = processAndSaveAllUserData()
    if err != nil {
        fmt.Println("Error processing and saving all user data:", err)
        return
    }
}