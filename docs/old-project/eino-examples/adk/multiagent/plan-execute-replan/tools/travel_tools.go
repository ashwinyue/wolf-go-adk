/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tools

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// WeatherRequest 表示天气查询请求
type WeatherRequest struct {
	City string `json:"city" jsonschema_description:"获取天气信息的城市名称"`
	Date string `json:"date" jsonschema_description:"日期，格式为YYYY-MM-DD（可选）"`
}

// WeatherResponse 表示天气信息
type WeatherResponse struct {
	City        string `json:"city"`
	Temperature int    `json:"temperature"`
	Condition   string `json:"condition"`
	Date        string `json:"date"`
	Error       string `json:"error,omitempty"`
}

// FlightRequest 表示航班搜索请求
type FlightRequest struct {
	From       string `json:"from" jsonschema_description:"出发城市"`
	To         string `json:"to" jsonschema_description:"目的地城市"`
	Date       string `json:"date" jsonschema_description:"出发日期，格式为YYYY-MM-DD"`
	Passengers int    `json:"passengers" jsonschema_description:"乘客人数"`
}

// FlightResponse 表示航班搜索结果
type FlightResponse struct {
	Flights []Flight `json:"flights"`
	Error   string   `json:"error,omitempty"`
}

type Flight struct {
	Airline   string `json:"airline"`
	FlightNo  string `json:"flight_no"`
	Departure string `json:"departure"`
	Arrival   string `json:"arrival"`
	Price     int    `json:"price"`
	Duration  string `json:"duration"`
}

// HotelRequest 表示酒店搜索请求
type HotelRequest struct {
	City     string `json:"city" jsonschema_description:"搜索酒店的城市"`
	CheckIn  string `json:"check_in" jsonschema_description:"入住日期，格式为YYYY-MM-DD"`
	CheckOut string `json:"check_out" jsonschema_description:"退房日期，格式为YYYY-MM-DD"`
	Guests   int    `json:"guests" jsonschema_description:"客人数量"`
}

// HotelResponse 表示酒店搜索结果
type HotelResponse struct {
	Hotels []Hotel `json:"hotels"`
	Error  string  `json:"error,omitempty"`
}

type Hotel struct {
	Name      string   `json:"name"`
	Rating    float64  `json:"rating"`
	Price     int      `json:"price"`
	Location  string   `json:"location"`
	Amenities []string `json:"amenities"`
}

// AttractionRequest 表示旅游景点搜索请求
type AttractionRequest struct {
	City     string `json:"city" jsonschema_description:"搜索景点的城市"`
	Category string `json:"category" jsonschema_description:"景点类别（博物馆、公园、地标、历史古迹等）"`
}

// AttractionResponse 表示景点搜索结果
type AttractionResponse struct {
	Attractions []Attraction `json:"attractions"`
	Error       string       `json:"error,omitempty"`
}

type Attraction struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Rating      float64 `json:"rating"`
	OpenHours   string  `json:"open_hours"`
	TicketPrice int     `json:"ticket_price"`
	Category    string  `json:"category"`
}

// NewWeatherTool 模拟天气工具实现
func NewWeatherTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("get_weather", "获取特定城市和日期的天气信息",
		func(ctx context.Context, req *WeatherRequest) (*WeatherResponse, error) {
			if req.City == "" {
				return &WeatherResponse{Error: "城市名称是必需的"}, nil
			}

			// 模拟天气数据
			weathers := map[string]WeatherResponse{
				"北京": {City: "北京", Temperature: 15, Condition: "晴天", Date: req.Date},
				"上海": {City: "上海", Temperature: 20, Condition: "多云", Date: req.Date},
				"东京": {City: "东京", Temperature: 18, Condition: "雨天", Date: req.Date},
				"巴黎": {City: "巴黎", Temperature: 12, Condition: "阴天", Date: req.Date},
				"纽约": {City: "纽约", Temperature: 8, Condition: "雪天", Date: req.Date},
			}

			if weather, exists := weathers[req.City]; exists {
				return &weather, nil
			}

			// 根据城市和日期为未知城市生成一致的天气
			conditions := []string{"晴天", "多云", "雨天", "阴天"}
			hashInput := req.City + req.Date
			return &WeatherResponse{
				City:        req.City,
				Temperature: consistentHashing(hashInput+"temp", 5, 35), // 5-35°C
				Condition:   conditions[consistentHashing(hashInput+"cond", 0, len(conditions)-1)],
				Date:        req.Date,
			}, nil
		})
}

// NewFlightSearchTool 模拟航班搜索工具实现
func NewFlightSearchTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("search_flights", "搜索城市之间的航班",
		func(ctx context.Context, req *FlightRequest) (*FlightResponse, error) {
			if req.From == "" || req.To == "" {
				return &FlightResponse{Error: "出发城市和目的地城市是必需的"}, nil
			}

			// 模拟航班数据
			airlines := []string{"中国国航", "中国东方航空", "中国南方航空", "联合航空", "达美航空"}

			flights := make([]Flight, 3)
			hashInput := req.From + req.To + req.Date
			for i := 0; i < 3; i++ {
				flightHash := fmt.Sprintf("%s%d", hashInput, i)
				airlineIdx := consistentHashing(flightHash+"airline", 0, len(airlines)-1)

				// 生成出发和到达时间
				depHour := consistentHashing(flightHash+"dephour", 0, 23)
				depMin := consistentHashing(flightHash+"depmin", 0, 59)
				arrHour := consistentHashing(flightHash+"arrhour", 0, 23)
				arrMin := consistentHashing(flightHash+"arrmin", 0, 59)

				// 根据出发和到达时间计算持续时间
				depTotalMin := depHour*60 + depMin
				arrTotalMin := arrHour*60 + arrMin

				// 处理到达时间为第二天的情况（如果到达时间早于出发时间）
				if arrTotalMin <= depTotalMin {
					arrTotalMin += 24 * 60 // 增加24小时
				}

				durationMin := arrTotalMin - depTotalMin
				durationHour := durationMin / 60
				durationMinRemainder := durationMin % 60

				flights[i] = Flight{
					Airline:   airlines[airlineIdx],
					FlightNo:  fmt.Sprintf("%s%d", airlines[airlineIdx][:2], consistentHashing(flightHash+"flightno", 1000, 9999)),
					Departure: fmt.Sprintf("%02d:%02d", depHour, depMin),
					Arrival:   fmt.Sprintf("%02d:%02d", arrHour, arrMin),
					Price:     consistentHashing(flightHash+"price", 500, 2500), // $500-2500
					Duration:  fmt.Sprintf("%dh %dm", durationHour, durationMinRemainder),
				}
			}

			return &FlightResponse{Flights: flights}, nil
		})
}

// NewHotelSearchTool 模拟酒店搜索工具实现
func NewHotelSearchTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("search_hotels", "在城市中搜索酒店",
		func(ctx context.Context, req *HotelRequest) (*HotelResponse, error) {
			if req.City == "" {
				return &HotelResponse{Error: "城市名称是必需的"}, nil
			}

			// 模拟酒店数据
			hotelNames := []string{"豪华大酒店", "市中心酒店", "度假村", "经济型酒店", "商务酒店"}
			amenities := [][]string{
				{"WiFi", "游泳池", "健身房", "水疗中心"},
				{"WiFi", "早餐", "停车场"},
				{"WiFi", "游泳池", "餐厅", "酒吧", "礼宾服务"},
				{"WiFi", "早餐"},
				{"WiFi", "商务中心", "会议室"},
			}

			hotels := make([]Hotel, 4)
			hashInput := req.City + req.CheckIn + req.CheckOut
			for i := 0; i < 4; i++ {
				hotelHash := fmt.Sprintf("%s%d", hashInput, i)
				hotels[i] = Hotel{
					Name:      fmt.Sprintf("%s %s", req.City, hotelNames[consistentHashing(hotelHash+"name", 0, len(hotelNames)-1)]),
					Rating:    float64(consistentHashing(hotelHash+"rating", 20, 50)) / 10.0, // 2.0-5.0
					Price:     consistentHashing(hotelHash+"price", 50, 350),                 // $50-350 per night
					Location:  fmt.Sprintf("%s Downtown", req.City),
					Amenities: amenities[consistentHashing(hotelHash+"amenities", 0, len(amenities)-1)],
				}
			}

			return &HotelResponse{Hotels: hotels}, nil
		})
}

// NewAttractionSearchTool 模拟景点搜索工具实现
func NewAttractionSearchTool(ctx context.Context) (tool.BaseTool, error) {
	return utils.InferTool("search_attractions", "在城市中搜索旅游景点",
		func(ctx context.Context, req *AttractionRequest) (*AttractionResponse, error) {
			if req.City == "" {
				return &AttractionResponse{Error: "城市名称是必需的"}, nil
			}

			// 根据城市模拟景点数据
			attractionsByCity := map[string][]Attraction{
				"北京": {
					{Name: "故宫", Description: "古代皇宫", Rating: 4.8, OpenHours: "8:30-17:00", TicketPrice: 60, Category: "历史古迹"},
					{Name: "长城", Description: "历史防御工事", Rating: 4.9, OpenHours: "6:00-18:00", TicketPrice: 45, Category: "地标"},
					{Name: "天坛", Description: "皇家祭祀坛", Rating: 4.6, OpenHours: "6:00-22:00", TicketPrice: 35, Category: "公园"},
				},
				"巴黎": {
					{Name: "埃菲尔铁塔", Description: "标志性的铁格子塔", Rating: 4.7, OpenHours: "9:30-23:45", TicketPrice: 25, Category: "地标"},
					{Name: "卢浮宫博物馆", Description: "世界最大的艺术博物馆", Rating: 4.8, OpenHours: "9:00-18:00", TicketPrice: 17, Category: "博物馆"},
					{Name: "巴黎圣母院", Description: "中世纪天主教大教堂", Rating: 4.5, OpenHours: "8:00-18:45", TicketPrice: 0, Category: "地标"},
				},
				"东京": {
					{Name: "浅草寺", Description: "古老的佛教寺庙", Rating: 4.4, OpenHours: "6:00-17:00", TicketPrice: 0, Category: "地标"},
					{Name: "东京国立博物馆", Description: "最大的文化文物收藏", Rating: 4.3, OpenHours: "9:30-17:00", TicketPrice: 1000, Category: "博物馆"},
					{Name: "上野公园", Description: "拥有博物馆的大型公共公园", Rating: 4.2, OpenHours: "5:00-23:00", TicketPrice: 0, Category: "公园"},
				},
			}

			if attractions, exists := attractionsByCity[req.City]; exists {
				// 如果指定了类别，则进行过滤
				if req.Category != "" {
					var filtered []Attraction
					for _, attraction := range attractions {
						if attraction.Category == req.Category {
							filtered = append(filtered, attraction)
						}
					}
					return &AttractionResponse{Attractions: filtered}, nil
				}
				return &AttractionResponse{Attractions: attractions}, nil
			}

			// 为未知城市生成随机景点
			attractionNames := []string{"中央博物馆", "城市公园", "历史广场", "艺术画廊", "文化中心"}
			categories := []string{"博物馆", "公园", "地标", "历史古迹", "文化"}

			attractions := make([]Attraction, 3)
			hashInput := req.City + req.Category
			for i := 0; i < 3; i++ {
				attractionHash := fmt.Sprintf("%s%d", hashInput, i)
				attractions[i] = Attraction{
					Name:        fmt.Sprintf("%s %s", req.City, attractionNames[consistentHashing(attractionHash+"name", 0, len(attractionNames)-1)]),
					Description: "热门旅游景点",
					Rating:      float64(consistentHashing(attractionHash+"rating", 30, 50)) / 10.0, // 3.0-5.0
					OpenHours:   "9:00-17:00",
					TicketPrice: consistentHashing(attractionHash+"price", 0, 50),
					Category:    categories[consistentHashing(attractionHash+"category", 0, len(categories)-1)],
				}
			}

			return &AttractionResponse{Attractions: attractions}, nil
		})
}

// GetAllTravelTools 返回所有旅行相关工具
func GetAllTravelTools(ctx context.Context) ([]tool.BaseTool, error) {
	weatherTool, err := NewWeatherTool(ctx)
	if err != nil {
		return nil, err
	}

	flightTool, err := NewFlightSearchTool(ctx)
	if err != nil {
		return nil, err
	}

	hotelTool, err := NewHotelSearchTool(ctx)
	if err != nil {
		return nil, err
	}

	attractionTool, err := NewAttractionSearchTool(ctx)
	if err != nil {
		return nil, err
	}

	askForClarificationTool := NewAskForClarificationTool()

	return []tool.BaseTool{weatherTool, flightTool, hotelTool, attractionTool, askForClarificationTool}, nil
}

// consistentHashing 使用Go标准库hash/fnv实现一致性哈希
func consistentHashing(s string, min, max int) int {
	// 使用Go标准库的FNV-1a哈希算法
	h := fnv.New32a()
	h.Write([]byte(s))
	hash := h.Sum32()

	// 映射到范围[min, max]
	return min + int(hash)%(max-min+1)
}
