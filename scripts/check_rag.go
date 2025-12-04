//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"

	milvusClient "github.com/milvus-io/milvus-sdk-go/v2/client"
)

func main() {
	ctx := context.Background()

	// 连接 Milvus
	cli, err := milvusClient.NewClient(ctx, milvusClient.Config{
		Address: "localhost:19530",
	})
	if err != nil {
		log.Fatalf("连接 Milvus 失败: %v", err)
	}
	defer cli.Close()

	// 检查 collection 是否存在
	exists, err := cli.HasCollection(ctx, "wolf_episodes")
	if err != nil {
		log.Fatalf("检查 Collection 失败: %v", err)
	}

	if !exists {
		fmt.Println("❌ Collection 'wolf_episodes' 不存在")
		fmt.Println("   RAG 系统可能未正确初始化")
		return
	}

	fmt.Println("✅ Collection 'wolf_episodes' 存在")

	// 获取 collection 统计信息
	stats, err := cli.GetCollectionStatistics(ctx, "wolf_episodes")
	if err != nil {
		log.Printf("获取统计信息失败: %v", err)
	} else {
		fmt.Printf("   统计信息: %v\n", stats)
	}

	// 列出所有 collections
	collections, err := cli.ListCollections(ctx)
	if err != nil {
		log.Printf("列出 Collections 失败: %v", err)
	} else {
		fmt.Printf("\n📦 所有 Collections:\n")
		for _, col := range collections {
			fmt.Printf("   - %s\n", col.Name)
		}
	}
}
